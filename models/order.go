package models

import (
	"github.com/shopspring/decimal"
	"github.com/spf13/viper"
	"gorm.io/gorm"
	"sync"
	"time"
)

type OrderStatus uint

const (
	OrderWait OrderStatus = iota
	OrderCancel
	OrderPending
	OrderSuccess
	OrderFailed
	OrderComplete
)

var OrderStatusMsgs = []string{
	"Unpaid",
	"Cancel",
	"Pending",
	"Success",
	"Fail",
	"Complete",
}

var mutex sync.Mutex

type Order struct {
	ID             uint `json:"id" gorm:"primarykey"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
	DeletedAt      gorm.DeletedAt  `gorm:"index"`
	OrderID        string          `json:"order_id"`
	PType          ProductType     `json:"type"`
	Quantity       uint64          `json:"quantity"`
	Price          decimal.Decimal `json:"price"`
	PaymentID      string          `json:"payment_id"`
	Payment        PaymentChannel  `json:"payment_channel"`
	PaymentAccount string          `json:"payment_account"`
	ReceiveAccount string          `json:"receive_account"`
	PaymentAmount  string          `json:"payment_amount"`
	PaymentTime    time.Time       `json:"payment_time"`
	Status         OrderStatus     `json:"status"`
	Hash           string          `json:"hash"`
	Discount       decimal.Decimal `json:"-"`
	Discount1      decimal.Decimal `json:"discount"`

	UserCouponID uint `json:"-"`

	//CouponID uint `json:"-"`
	//Coupon   Coupon

	UserID uint `json:"-"`
	User   User

	CurrencyID uint `json:"-"`
	Currency   Currency

	Created        string `json:"created_at" gorm:"-"`
	Updated        string `json:"updated_at" gorm:"-"`
	QuantityStr    string `json:"size_str" gorm:"-"`
	PaymentStr     string `json:"payment_channel_str" gorm:"-"`
	PaymentTimeStr string `json:"payment_time_str"`
	PTypeStr       string `json:"type_str" gorm:"-"`

	StatusStr string `json:"status_str" gorm:"-"`
	URL       string `json:"url" gorm:"-"`
}

func (item *Order) AfterFind(tx *gorm.DB) (err error) {
	item.Created = item.CreatedAt.Format(TIME_FORMAT)
	item.Updated = item.UpdatedAt.Format(TIME_FORMAT)
	if item.PaymentTime.Unix() > 0 {
		item.PaymentTimeStr = item.PaymentTime.Format(TIME_FORMAT)
	}
	item.QuantityStr = ByteSize(item.Quantity)
	item.PTypeStr = ProductTypeMsgs[item.PType]
	item.StatusStr = OrderStatusMsgs[item.Status]
	if len(item.Hash) > 0 {
		item.URL = viper.GetString("bsc.browser") + "tx/" + item.Hash
	}
	item.PaymentStr = PaymentChannelMsgs[item.Payment]
	return
}
