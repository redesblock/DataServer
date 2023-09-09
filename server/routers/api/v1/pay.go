package v1

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-pay/gopay"
	"github.com/go-pay/gopay/wechat/v3"
	"github.com/redesblock/dataserver/models"
	"github.com/redesblock/dataserver/server/pay"
	"github.com/shopspring/decimal"
	"github.com/smartwalle/alipay/v3"
	"github.com/spf13/viper"
	"github.com/stripe/stripe-go/v75"
	"github.com/stripe/stripe-go/v75/checkout/session"
	"github.com/stripe/stripe-go/v75/webhook"
	"gorm.io/gorm"
	"io/ioutil"
	"net/http"
	"os"
	"reflect"
	"sort"
	"strings"
	"time"
)

func AlipayTest(db *gorm.DB) func(c *gin.Context) {
	return func(c *gin.Context) {
		res, err := pay.AliPayTrade("test", generateOrderID(), "0.1")
		if err != nil {
			c.JSON(OKCode, NewResponse(c, ExecuteCode, err))
		}

		c.Redirect(http.StatusTemporaryRedirect, res)
	}
}

func WxpayTest(db *gorm.DB) func(c *gin.Context) {
	return func(c *gin.Context) {
		res, err := pay.WXTrade("test", generateOrderID(), "0.1")
		if err != nil {
			c.JSON(OKCode, NewResponse(c, ExecuteCode, err))
		}

		c.Redirect(http.StatusTemporaryRedirect, res)
	}
}

func StripeTest(db *gorm.DB) func(c *gin.Context) {
	return func(c *gin.Context) {
		res, err := pay.StripeTrade("test", generateOrderID(), "1.01", "usd")
		if err != nil {
			c.JSON(OKCode, NewResponse(c, ExecuteCode, err))
		}

		c.Redirect(http.StatusTemporaryRedirect, res)
	}
}

func AlipayNotify(db *gorm.DB) func(c *gin.Context) {
	return func(c *gin.Context) {
		var noti, err = pay.AlipayClient.DecodeNotification(c.Request)
		if err != nil {
			c.JSON(OKCode, NewResponse(c, ExecuteCode, err))
			return
		}
		b, _ := json.Marshal(noti)
		fmt.Println("alipay notify", string(b))

		var order models.Order
		ret := db.Model(&models.Order{}).Where("order_id = ?", noti.OutTradeNo).Find(&order)
		if err := ret.Error; err != nil {
			c.JSON(OKCode, NewResponse(c, ExecuteCode, err))
			return
		}
		switch noti.TradeStatus {
		case alipay.TradeStatusWaitBuyerPay:
			if order.Status != models.OrderPending {
				order.Status = models.OrderPending
				if err := db.Save(&order).Error; err != nil {
					c.JSON(OKCode, NewResponse(c, ExecuteCode, err))
					return
				}
			}
		case alipay.TradeStatusClosed:
			if order.Status != models.OrderCancel {
				order.Status = models.OrderCancel
				if err := db.Save(&order).Error; err != nil {
					c.JSON(OKCode, NewResponse(c, ExecuteCode, err))
					return
				}
			}
		case alipay.TradeStatusSuccess:
			if order.Status != models.OrderSuccess {
				order.Status = models.OrderSuccess
				order.PaymentID = noti.TradeNo
				order.PaymentAccount = noti.BuyerLogonId
				if len(order.PaymentAccount) == 0 {
					order.PaymentAccount = noti.BuyerId
				}
				order.ReceiveAccount = noti.SellerEmail
				if len(order.ReceiveAccount) == 0 {
					order.ReceiveAccount = noti.SellerId
				}
				order.PaymentAmount = noti.TotalAmount
				order.PaymentTime, _ = time.Parse(models.TIME_FORMAT, noti.GmtPayment)
				if err := db.Transaction(func(tx *gorm.DB) error {
					var user models.User
					if ret := tx.Model(&models.User{}).Where("id = ?", order.UserID).Find(&user); ret.Error != nil {
						return ret.Error
					} else if ret.RowsAffected == 0 {
						return fmt.Errorf("not found user")
					}
					user.TotalStorage += order.Quantity
					if err := tx.Save(&user).Error; err != nil {
						return err
					}
					if order.UserCouponID > 0 {
						var userCoupon models.UserCoupon
						if ret := tx.Model(&models.UserCoupon{}).Where("id = ?", order.UserCouponID).Find(&userCoupon); ret.Error != nil {
							return ret.Error
						} else if ret.RowsAffected == 0 {
							return fmt.Errorf("not found user coupon")
						}
						userCoupon.Status = models.UserCouponStatus_Used
						if err := tx.Save(&userCoupon).Error; err != nil {
							return err
						}
					}

					return tx.Save(&order).Error
				}); err != nil {
					c.JSON(OKCode, NewResponse(c, ExecuteCode, err))
					return
				}
			}
		case alipay.TradeStatusFinished:
		}
		alipay.ACKNotification(c.Writer)
	}
}

func WxPayNotify(db *gorm.DB) func(c *gin.Context) {
	return func(c *gin.Context) {
		notifyReq, err := wechat.V3ParseNotify(c.Request)
		if err != nil {
			c.JSON(OKCode, NewResponse(c, ExecuteCode, err))
			return
		}

		// 获取微信平台证书
		certMap := pay.WXClient.WxPublicKeyMap()
		// 验证异步通知的签名
		err = notifyReq.VerifySignByPKMap(certMap)
		if err != nil {
			c.JSON(OKCode, NewResponse(c, ExecuteCode, err))
			return
		}

		noti, err := notifyReq.DecryptCipherText(string(pay.WXClient.ApiV3Key))
		if err != nil {
			c.JSON(OKCode, NewResponse(c, ExecuteCode, err))
			return
		}

		b, _ := json.Marshal(noti)
		fmt.Println("wxpay notify", string(b))

		var order models.Order
		ret := db.Model(&models.Order{}).Where("order_id = ?", noti.OutTradeNo).Find(&order)
		if err := ret.Error; err != nil {
			c.JSON(OKCode, NewResponse(c, ExecuteCode, err))
			return
		}

		switch noti.TradeState {
		case wechat.TradeStateClosed:
			if order.Status != models.OrderCancel {
				order.Status = models.OrderCancel
				if err := db.Save(&order).Error; err != nil {
					c.JSON(OKCode, NewResponse(c, ExecuteCode, err))
					return
				}
			}
		case wechat.TradeStateNoPay:
			if order.Status != models.OrderPending {
				order.Status = models.OrderPending
				if err := db.Save(&order).Error; err != nil {
					c.JSON(OKCode, NewResponse(c, ExecuteCode, err))
					return
				}
			}
		case wechat.TradeStateSuccess:
			if order.Status != models.OrderSuccess {
				order.Status = models.OrderSuccess
				order.PaymentID = noti.TransactionId
				order.PaymentAccount = noti.Payer.Openid
				order.PaymentTime, _ = time.Parse(time.RFC3339, noti.SuccessTime)
				if err := db.Transaction(func(tx *gorm.DB) error {
					var user models.User
					if ret := tx.Model(&models.User{}).Where("id = ?", order.UserID).Find(&user); ret.Error != nil {
						return ret.Error
					} else if ret.RowsAffected == 0 {
						return fmt.Errorf("not found user")
					}
					user.TotalStorage += order.Quantity
					if err := tx.Save(&user).Error; err != nil {
						return err
					}
					if order.UserCouponID > 0 {
						var userCoupon models.UserCoupon
						if ret := tx.Model(&models.UserCoupon{}).Where("id = ?", order.UserCouponID).Find(&userCoupon); ret.Error != nil {
							return ret.Error
						} else if ret.RowsAffected == 0 {
							return fmt.Errorf("not found user coupon")
						}
						userCoupon.Status = models.UserCouponStatus_Used
						if err := tx.Save(&userCoupon).Error; err != nil {
							return err
						}
					}
					return tx.Save(&order).Error
				}); err != nil {
					c.JSON(OKCode, NewResponse(c, ExecuteCode, err))
					return
				}
			}
		}
		c.JSON(http.StatusOK, &wechat.V3NotifyRsp{Code: gopay.SUCCESS, Message: "成功"})
	}
}

type RequestData struct {
	Amount     int64      `form:"amount"`
	Currency   string     `form:"currency"`
	ID         string     `form:"id"`
	Note       string     `form:"note"`
	Reference  string     `form:"reference"`
	RmbAmount  int64      `form:"rmb_amount"`
	Status     string     `form:"status"`
	SysReserve SysReserve `form:"sys_reserve"`
	Time       string     `form:"time"`
	VerifySign string     `form:"verify_sign"`
}
type SysReserve struct {
	VendorID string `json:"vendor_id"`
}

func getFieldValue(data RequestData, key string) string {
	keys := strings.Split(key, ".")
	value := reflect.ValueOf(data)
	for _, k := range keys {
		if value.Kind() == reflect.Ptr || value.Kind() == reflect.Interface {
			value = value.Elem()
		}
		value = value.FieldByName(k)
	}
	return fmt.Sprintf("%v", value)
}

func NihaoPayNotify(db *gorm.DB) func(c *gin.Context) {
	return func(c *gin.Context) {
		var requestData RequestData
		if err := c.ShouldBind(&requestData); err != nil {
			fmt.Println("NihaoPayNotify", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		b, _ := json.Marshal(requestData)
		fmt.Println("nihaopay notify", string(b))

		// 将请求体字段按键名升序排序
		keys := make([]string, 0, 8)
		keys = append(keys, "amount", "currency", "id", "note", "reference", "rmb_amount", "status", "sys_reserve.vendor_id")
		sort.Strings(keys)

		// 构建拼接后的字符串
		var sb strings.Builder
		for _, key := range keys {
			val := getFieldValue(requestData, key)
			if val != "" {
				sb.WriteString(fmt.Sprintf("%s=%s&", key, val))
			}
		}

		token := viper.GetString("nihaopay.key")
		tokenMD5 := md5.Sum([]byte(token))
		tokenStr := hex.EncodeToString(tokenMD5[:])
		sb.WriteString(fmt.Sprintf("MD5(Token)=%s", tokenStr))

		// 对整个字符串进行MD5哈希并转换为小写字符
		dataStr := sb.String()
		dataMD5 := md5.Sum([]byte(dataStr))
		dataMD5Str := hex.EncodeToString(dataMD5[:])

		fmt.Println("nihaopay verify", requestData.VerifySign, dataMD5Str, requestData.VerifySign == dataMD5Str)

		id := requestData.ID
		orderID := requestData.Reference
		status := requestData.Status
		success := requestData.Time
		amount := requestData.RmbAmount
		if amount == 0 {
			amount = requestData.Amount
		}

		var order models.Order
		ret := db.Model(&models.Order{}).Where("order_id = ?", orderID).Find(&order)
		if err := ret.Error; err != nil {
			c.JSON(OKCode, NewResponse(c, ExecuteCode, err))
			return
		}

		switch status {
		case "closed":
			if order.Status != models.OrderCancel {
				order.Status = models.OrderCancel
				if err := db.Save(&order).Error; err != nil {
					c.JSON(OKCode, NewResponse(c, ExecuteCode, err))
					return
				}
			}
		case "failure":
			if order.Status != models.OrderFailed {
				order.Status = models.OrderFailed
				if err := db.Save(&order).Error; err != nil {
					c.JSON(OKCode, NewResponse(c, ExecuteCode, err))
					return
				}
			}
		case "pending":
			if order.Status != models.OrderPending {
				order.Status = models.OrderPending
				if err := db.Save(&order).Error; err != nil {
					c.JSON(OKCode, NewResponse(c, ExecuteCode, err))
					return
				}
			}
		case "success":
			if order.Status != models.OrderSuccess {
				order.Status = models.OrderSuccess
				order.PaymentID = id
				order.PaymentAccount = decimal.NewFromInt(amount).Div(decimal.NewFromInt(100)).String()
				order.PaymentTime, _ = time.Parse(time.RFC3339, success)
				if err := db.Transaction(func(tx *gorm.DB) error {
					var user models.User
					if ret := tx.Model(&models.User{}).Where("id = ?", order.UserID).Find(&user); ret.Error != nil {
						return ret.Error
					} else if ret.RowsAffected == 0 {
						return fmt.Errorf("not found user")
					}
					user.TotalStorage += order.Quantity
					if err := tx.Save(&user).Error; err != nil {
						return err
					}
					if order.UserCouponID > 0 {
						var userCoupon models.UserCoupon
						if ret := tx.Model(&models.UserCoupon{}).Where("id = ?", order.UserCouponID).Find(&userCoupon); ret.Error != nil {
							return ret.Error
						} else if ret.RowsAffected == 0 {
							return fmt.Errorf("not found user coupon")
						}
						userCoupon.Status = models.UserCouponStatus_Used
						if err := tx.Save(&userCoupon).Error; err != nil {
							return err
						}
					}
					return tx.Save(&order).Error
				}); err != nil {
					c.JSON(OKCode, NewResponse(c, ExecuteCode, err))
					return
				}
			}
		}
		c.JSON(http.StatusOK, &wechat.V3NotifyRsp{Code: gopay.SUCCESS, Message: "成功"})
	}
}

func StripeNotify(db *gorm.DB) func(c *gin.Context) {
	return func(c *gin.Context) {
		fulfillOrder := func(s stripe.CheckoutSession) error {
			params := &stripe.CheckoutSessionParams{}
			params.AddExpand("line_items")
			sessionWithLineItems, _ := session.Get(s.ID, params)
			b, _ := json.Marshal(sessionWithLineItems)
			fmt.Println("stripe notify", string(b))

			for _, item := range sessionWithLineItems.LineItems.Data {
				orderID := item.Description
				var order models.Order
				ret := db.Model(&models.Order{}).Where("order_id = ?", orderID).Find(&order)
				if err := ret.Error; err != nil {
					return err
				}
				order.Status = models.OrderSuccess
				order.PaymentID = sessionWithLineItems.ID
				if sessionWithLineItems.CustomerDetails != nil {
					order.PaymentAccount = sessionWithLineItems.CustomerDetails.Email
				}

				order.PaymentAmount = decimal.NewFromInt(int64(item.AmountTotal)).Div(decimal.NewFromInt(100)).String()
				order.PaymentTime = time.Unix(sessionWithLineItems.Created, 0)
				if err := db.Transaction(func(tx *gorm.DB) error {
					var user models.User
					if ret := tx.Model(&models.User{}).Where("id = ?", order.UserID).Find(&user); ret.Error != nil {
						return ret.Error
					} else if ret.RowsAffected == 0 {
						return fmt.Errorf("not found user")
					}
					user.TotalStorage += order.Quantity
					if err := tx.Save(&user).Error; err != nil {
						return err
					}
					if order.UserCouponID > 0 {
						var userCoupon models.UserCoupon
						if ret := tx.Model(&models.UserCoupon{}).Where("id = ?", order.UserCouponID).Find(&userCoupon); ret.Error != nil {
							return ret.Error
						} else if ret.RowsAffected == 0 {
							return fmt.Errorf("not found user coupon")
						}
						userCoupon.Status = models.UserCouponStatus_Used
						if err := tx.Save(&userCoupon).Error; err != nil {
							return err
						}
					}
					return tx.Save(&order).Error
				}); err != nil {
					return err
				}
			}
			return nil
		}

		const MaxBodyBytes = int64(65536)
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, MaxBodyBytes)

		body, err := ioutil.ReadAll(c.Request.Body)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading request body: %v\n", err)
			c.JSON(http.StatusServiceUnavailable, NewResponse(c, ExecuteCode, err))
			return
		}

		// Pass the request body and Stripe-Signature header to ConstructEvent, along with the webhook signing key
		// You can find your endpoint's secret in your webhook settings
		endpointSecret := viper.GetString("stripe.secret")
		signature := c.Request.Header.Get("Stripe-Signature")
		event, err := webhook.ConstructEvent(body, signature, endpointSecret)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error verifying webhook signature: %v\n", err)
			c.JSON(http.StatusBadRequest, NewResponse(c, ExecuteCode, err))
			return
		}

		switch event.Type {
		case "checkout.session.completed":
			var session stripe.CheckoutSession
			err := json.Unmarshal(event.Data.Raw, &session)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error parsing webhook JSON: %v\n", err)
				c.JSON(http.StatusBadRequest, NewResponse(c, ExecuteCode, err))
				return
			}

			// Check if the order is paid (for example, from a card payment)
			//
			// A delayed notification payment will have an `unpaid` status, as
			// you're still waiting for funds to be transferred from the customer's
			// account.
			orderPaid := session.PaymentStatus == stripe.CheckoutSessionPaymentStatusPaid

			if orderPaid {
				// Fulfill the purchase
				fulfillOrder(session)
			}

		case "checkout.session.async_payment_succeeded":
			var session stripe.CheckoutSession
			err := json.Unmarshal(event.Data.Raw, &session)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error parsing webhook JSON: %v\n", err)
				c.JSON(http.StatusBadRequest, NewResponse(c, ExecuteCode, err))
				return
			}

			// Fulfill the purchase
			fulfillOrder(session)
		}

		c.JSON(OKCode, "")

		//var order models.Order
		//ret := db.Model(&models.Order{}).Where("order_id = ?", s. .OutTradeNo).Find(&order)
		//if err := ret.Error; err != nil {
		//	c.JSON(OKCode, NewResponse(c, ExecuteCode, err))
		//	return
		//}
		//switch s.Status {
		//case stripe.CheckoutSessionStatusComplete:
		//
		//}
		//if order.Status != models.OrderSuccess {
		//	order.Status = models.OrderSuccess
		//	order.PaymentID = s.ID
		//	order.PaymentAccount = s.CustomerEmail
		//	order.ReceiveAccount = decimal.NewFromInt(int64(s.AmountTotal)).Div(decimal.NewFromInt(100)).String()
		//	order.PaymentAmount = decimal.NewFromInt(int64(s.AmountTotal)).Div(decimal.NewFromInt(100)).String()
		//	//order.PaymentTime, _ = time.Parse(models.TIME_FORMAT, )
		//	if err := db.Transaction(func(tx *gorm.DB) error {
		//		var user models.User
		//		if ret := tx.Model(&models.User{}).Where("id = ?", order.UserID).Find(&user); ret.Error != nil {
		//			return ret.Error
		//		} else if ret.RowsAffected == 0 {
		//			return fmt.Errorf("not found user")
		//		}
		//		user.TotalStorage += order.Quantity
		//		if err := tx.Save(&user).Error; err != nil {
		//			return err
		//		}
		//		return tx.Save(&order).Error
		//	}); err != nil {
		//		c.JSON(OKCode, NewResponse(c, ExecuteCode, err))
		//		return
		//	}
		//}
	}
}
