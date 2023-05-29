package models

import (
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"strings"
	"time"
)

type CouponType uint

var UnlimitedTime = time.Unix(0, 0)

const (
	CouponType_Free = iota
	CouponType_Discount
)

var CouponTypeMsgs = map[CouponType]string{
	CouponType_Free:     "free",
	CouponType_Discount: "discount",
}

type UserCouponStatus uint

const (
	UserCouponStatus_Normal UserCouponStatus = iota
	UserCouponStatus_Used
	UserCouponStatus_Expired
)

type CouponStatus uint

const (
	CouponStatus_NotStart CouponStatus = iota
	CouponStatus_InProcess
	CouponStatus_Completed
	CouponStatus_Expired
)

var UserCouponStatusMsgs = []string{
	"usable",
	"used",
	"expired",
}

var CouponStatusMsgs = []string{
	"not started",
	"in process",
	"completed",
	"expired",
}

type Coupon struct {
	ID                 uint `json:"id" gorm:"primarykey"`
	CreatedAt          time.Time
	UpdatedAt          time.Time
	DeletedAt          gorm.DeletedAt  `gorm:"index"`
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
	Status             CouponStatus    `json:"status"`

	StartTimeStr  string `json:"start_time_str" gorm:"-"`
	EndTimeStr    string `json:"end_time_str" gorm:"-"`
	PTypeStr      string `json:"product_type_str" gorm:"-"`
	CouponTypeStr string `json:"coupon_type_str" gorm:"-"`
	StatusStr     string `json:"status_str" gorm:"-"`
	Created       string `json:"created_at" gorm:"-"`
	Updated       string `json:"updated_at" gorm:"-"`
}

func (item *Coupon) AfterFind(tx *gorm.DB) (err error) {
	item.StartTimeStr = item.StartTime.Format(TIME_FORMAT)
	item.EndTimeStr = item.EndTime.Format(TIME_FORMAT)
	//if item.StartTime.Unix() == 0 {
	//	item.StartTimeStr = "0"
	//}
	//if item.EndTime.Unix() == 0 {
	//	item.EndTimeStr = "0"
	//}
	item.StatusStr = CouponStatusMsgs[item.Status]
	item.CouponTypeStr = CouponTypeMsgs[item.CouponType]
	if str, ok := ProductTypeMsgs[item.PType]; ok {
		item.PTypeStr = str
	} else {
		var strs []string
		for p, s := range ProductTypeMsgs {
			if p&item.PType > 0 {
				strs = append(strs, s)
			}
		}
		item.PTypeStr = strings.Join(strs, ",")
	}

	item.Created = item.CreatedAt.Format(TIME_FORMAT)
	item.Updated = item.UpdatedAt.Format(TIME_FORMAT)
	return
}

type UserCoupon struct {
	ID        uint `json:"id" gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`

	UserID uint             `json:"-"`
	User   User             `json:"user"`
	Status UserCouponStatus `json:"status"`

	EndTime time.Time   `json:"end_time"`
	PType   ProductType `json:"product_type"`

	CouponID uint   `json:"-"`
	Coupon   Coupon `json:"coupon"`

	StatusStr string `json:"status_str" gorm:"-"`
	Created   string `json:"created_at" gorm:"-"`
	Updated   string `json:"updated_at" gorm:"-"`
}

func (item *UserCoupon) AfterFind(tx *gorm.DB) (err error) {
	item.StatusStr = UserCouponStatusMsgs[item.Status]
	item.Created = item.CreatedAt.Format(TIME_FORMAT)
	item.Updated = item.UpdatedAt.Format(TIME_FORMAT)
	return
}
