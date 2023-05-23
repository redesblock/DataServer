package v1

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/redesblock/dataserver/models"
	"github.com/redesblock/dataserver/server/pay"
	"github.com/smartwalle/alipay/v3"
	"gorm.io/gorm"
	"time"
)

func AlipayNotify(db *gorm.DB) func(c *gin.Context) {
	return func(c *gin.Context) {
		var noti, err = pay.AlipayClient.DecodeNotification(c.Request)
		if err != nil {
			c.JSON(OKCode, NewResponse(c, ExecuteCode, err))
			return
		}
		bts, _ := json.Marshal(noti)
		fmt.Println("notfiy", string(bts))
		_ = noti
		var order models.Order
		ret := db.Model(&models.Order{}).Where("order_id = ?", noti.OutTradeNo).Find(&order)
		if err := ret.Error; err != nil {
			c.JSON(OKCode, NewResponse(c, ExecuteCode, err))
			return
		}
		switch noti.TradeStatus {
		case alipay.TradeStatusWaitBuyerPay:
			order.Status = models.OrderPending
		case alipay.TradeStatusClosed:
			order.Status = models.OrderCancel
		case alipay.TradeStatusSuccess:
			order.Status = models.OrderSuccess
			order.PaymentID = noti.TradeNo
			order.PaymentAccount = noti.BuyerLogonId
			order.ReceiveAccount = noti.SellerEmail
			order.PaymentAmount = noti.TotalAmount
			order.PaymentTime, _ = time.Parse(models.TIME_FORMAT, noti.NotifyTime)
		case alipay.TradeStatusFinished:
		}
		if err := db.Save(&order).Error; err != nil {
			c.JSON(OKCode, NewResponse(c, ExecuteCode, err))
			return
		}
		alipay.ACKNotification(c.Writer)
	}
}
