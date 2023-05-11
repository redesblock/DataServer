package models

import (
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"time"
)

type CouponType uint

const (
	CouponType_Free = iota
	CouponType_Discount
)

var CouponTypeMsgs = map[CouponType]string{
	CouponType_Free:     "free",
	CouponType_Discount: "discount",
}

type Coupon struct {
	gorm.Model
	Name               string          `json:"name"`
	CouponType         CouponType      `json:"coupon_type"`
	PType              ProductType     `json:"product_type"`
	Discount           decimal.Decimal `json:"discount"`
	StorageQuantityMin uint64          `json:"storage_quantity_min"`
	StorageQuantityMax uint64          `json:"storage_quantity_max"`
	TrafficQuantityMin uint64          `json:"traffic_quantity_min"`
	TrafficQuantityMax uint64          `json:"traffic_quantity_max"`
	StartTime          time.Time       `json:"start_time"`
	EndTime            time.Time       `json:"end_time"`
	Sold               uint64          `json:"sold"`
	Reserve            uint64          `json:"reserve"`
	MaxClaim           uint64          `json:"max_claim"`

	StartTimeStr  string   `json:"start_time_str" gorm:"-"`
	EndTimeStr    string   `json:"end_time_str" gorm:"-"`
	PTypeStr      []string `json:"product_type_str" gorm:"-"`
	CouponTypeStr string   `json:"coupon_type_str" gorm:"-"`
	Created       string   `json:"created_at" gorm:"-"`
	Updated       string   `json:"updated_at" gorm:"-"`
}

func (item *Coupon) AfterFind(tx *gorm.DB) (err error) {
	item.StartTimeStr = item.StartTime.Format(TIME_FORMAT)
	item.EndTimeStr = item.EndTime.Format(TIME_FORMAT)
	item.CouponTypeStr = CouponTypeMsgs[item.CouponType]
	for p, s := range ProductTypeMsgs {
		if p^item.PType == 1 {
			item.PTypeStr = append(item.PTypeStr, s)
		}
	}
	item.Created = item.CreatedAt.Format(TIME_FORMAT)
	item.Updated = item.UpdatedAt.Format(TIME_FORMAT)
	return
}

type UserCoupon struct {
	gorm.Model

	UserID uint `json:"-"`
	User   User `json:"user"`
	Used   bool `json:"used"`

	CouponID uint   `json:"-"`
	Coupon   Coupon `json:"coupon"`
}
