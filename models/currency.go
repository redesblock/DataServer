package models

import (
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"strings"
)

type PaymentChannel uint

const (
	PaymentChannel_SignIn      PaymentChannel = 1
	PaymentChannel_Counon_Free PaymentChannel = 1 << 1
	PaymentChannel_Crypto      PaymentChannel = 1 << 2
	PaymentChannel_Alipay      PaymentChannel = 1 << 3
	PaymentChannel_WeChat      PaymentChannel = 1 << 4
)

var PaymentChannelMsgs = map[PaymentChannel]string{
	PaymentChannel_SignIn:      "SignIn",
	PaymentChannel_Counon_Free: "Free",
	PaymentChannel_Crypto:      "CryptoCurrency",
	PaymentChannel_Alipay:      "Alipay",
	PaymentChannel_WeChat:      "WeChat",
}

type Currency struct {
	gorm.Model
	Symbol    string          `json:"symbol" gorm:"unique"`
	Rate      decimal.Decimal `json:"rate"`
	Base      bool            `json:"base" gorm:"default: false"`
	Payment   PaymentChannel  `json:"payment_channel"`
	Receiptor string          `json:"receiptor"`

	PaymentStr string `json:"payment_channel_str" gorm:"-"`
	Created    string `json:"created_at" gorm:"-"`
	Updated    string `json:"updated_at" gorm:"-"`
}

func (item *Currency) AfterFind(tx *gorm.DB) (err error) {
	var strs []string
	for p, s := range PaymentChannelMsgs {
		if p&item.Payment > 0 {
			strs = append(strs, s)
		}
	}
	item.PaymentStr = strings.Join(strs, ",")
	item.Created = item.CreatedAt.Format(TIME_FORMAT)
	item.Updated = item.UpdatedAt.Format(TIME_FORMAT)
	return
}
