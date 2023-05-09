package models

import (
	"gorm.io/gorm"
)

type SignInPeriod uint

const (
	SignInPeriod_Day SignInPeriod = 0
	SignInPeriod_Week
	SignInPeriod_Month
	SignInPeriod_Year

	SignInPeriod_End
)

var SignInPeriodMsgs = []string{
	"1 Day",
	"1 Week",
	"1 Month",
	"1 Year",
}

type SignIn struct {
	gorm.Model
	PType    ProductType  `json:"type" gorm:"unique"`
	Quantity uint64       `json:"quantity"`
	Period   SignInPeriod `json:"period"`
	Enable   bool         `json:"enable"`

	QuantityStr string `json:"quantity_str" gorm:"-"`
	PeriodStr   string `json:"period_str" gorm:"-"`
	PTypeStr    string `json:"product_type_str" gorm:"-"`
	Created     string `json:"created_at" gorm:"-"`
	Updated     string `json:"updated_at" gorm:"-"`
}

func (item *SignIn) AfterFind(tx *gorm.DB) (err error) {
	item.PTypeStr = ProductTypeMsgs[item.PType]
	item.PeriodStr = SignInPeriodMsgs[item.Period]
	item.QuantityStr = ByteSize(item.Quantity)
	item.Created = item.CreatedAt.Format(TIME_FORMAT)
	item.Updated = item.UpdatedAt.Format(TIME_FORMAT)
	return
}
