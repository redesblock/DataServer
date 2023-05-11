package models

import (
	"fmt"
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
}

var mutex sync.Mutex

type Order struct {
	gorm.Model
	OrderID        string          `json:"order_id"`
	PType          ProductType     `json:"type" gorm:"unique"`
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

	UserID uint `json:"user_id"`
	User   User

	Created     string `json:"created_at" gorm:"-"`
	Updated     string `json:"updated_at" gorm:"-"`
	QuantityStr string `json:"size_str" gorm:"-"`
	StatusStr   string `json:"status_str" gorm:"-"`
	URL         string `json:"url" gorm:"-"`
}

func (item *Order) BeforeCreate(tx *gorm.DB) (err error) {
	var num int64
	mutex.Lock()
	defer mutex.Unlock()
	if err := tx.Count(&num).Error; err != nil {
		return err
	}
	item.OrderID = fmt.Sprintf("%s%d", time.Now().Format("20060102150405"), num)
	return
}

func (item *Order) AfterFind(tx *gorm.DB) (err error) {
	item.Created = item.CreatedAt.Format(TIME_FORMAT)
	item.Updated = item.UpdatedAt.Format(TIME_FORMAT)
	item.QuantityStr = ByteSize(item.Quantity)
	item.StatusStr = OrderStatusMsgs[item.Status]
	item.URL = viper.GetString("bsc.browser") + "tx/" + item.Hash
	return
}
