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
	"gorm.io/gorm"
	"net/http"
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
				order.PaymentTime, _ = time.Parse(models.TIME_FORMAT, noti.NotifyTime)
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

		case wechat.TradeStateSuccess:
			if order.Status != models.OrderSuccess {
				order.Status = models.OrderSuccess
				order.PaymentID = noti.TransactionId
				order.PaymentAccount = noti.Payer.Openid
				order.ReceiveAccount = decimal.NewFromInt(int64(noti.Amount.PayerTotal)).Div(decimal.NewFromInt(10)).String()
				order.PaymentAmount = decimal.NewFromInt(int64(noti.Amount.PayerTotal)).Div(decimal.NewFromInt(10)).String()
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
