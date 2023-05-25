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
	gorm.Model
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
	Discount       decimal.Decimal `json:"discount"`

	UserID uint `json:"-"`
	User   User

	CurrencyID uint `json:"-"`
	Currency   Currency

	Created     string `json:"created_at" gorm:"-"`
	Updated     string `json:"updated_at" gorm:"-"`
	QuantityStr string `json:"size_str" gorm:"-"`
	PaymentStr  string `json:"payment_str" gorm:"-"`
	StatusStr   string `json:"status_str" gorm:"-"`
	URL         string `json:"url" gorm:"-"`
}

func (item *Order) AfterFind(tx *gorm.DB) (err error) {
	item.Created = item.CreatedAt.Format(TIME_FORMAT)
	item.Updated = item.UpdatedAt.Format(TIME_FORMAT)
	item.QuantityStr = ByteSize(item.Quantity)
	item.StatusStr = OrderStatusMsgs[item.Status]
	item.URL = viper.GetString("bsc.browser") + "tx/" + item.Hash
	item.PaymentStr = PaymentChannelMsgs[item.Payment]
	return
}
