package v1

import (
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
	"github.com/stripe/stripe-go/v74"
	"github.com/stripe/stripe-go/v74/checkout/session"
	"github.com/stripe/stripe-go/v74/webhook"
	"gorm.io/gorm"
	"io/ioutil"
	"net/http"
	"os"
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
		res, err := pay.StripeTrade("test", generateOrderID(), "1.01")
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
				order.ReceiveAccount = noti.SellerEmail
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

		switch notifyReq.EventType {
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
				order.ReceiveAccount = decimal.NewFromInt(int64(noti.Amount.PayerTotal)).Div(decimal.NewFromInt(100)).String()
				order.PaymentAmount = decimal.NewFromInt(int64(noti.Amount.PayerTotal)).Div(decimal.NewFromInt(100)).String()
				order.PaymentTime, _ = time.Parse(models.TIME_FORMAT, noti.SuccessTime)
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

				order.ReceiveAccount = decimal.NewFromInt(int64(item.AmountTotal)).Div(decimal.NewFromInt(100)).String()
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
