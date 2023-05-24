package models

import (
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type ProductType uint

const (
	ProductType_ALL     ProductType = 0
	ProductType_Storage ProductType = 1
	ProductType_Traffic ProductType = 1 << 1
)

var ProductTypeMsgs = map[ProductType]string{
	ProductType_ALL:     "All",
	ProductType_Storage: "Storage",
	ProductType_Traffic: "Traffic",
}

type Product struct {
	gorm.Model
	PType    ProductType     `json:"type" gorm:"unique"`
	Quantity uint64          `json:"quantity"`
	Price    decimal.Decimal `json:"price"`

	PTypeStr    string `json:"product_type_str" gorm:"-"`
	QuantityStr string `json:"quantity_str" gorm:"-"`
	Created     string `json:"created_at" gorm:"-"`
	Updated     string `json:"updated_at" gorm:"-"`

	CurrencyID int `json:"-"`
	Currency   Currency
}

func (item *Product) AfterFind(tx *gorm.DB) (err error) {
	item.QuantityStr = ByteSize(item.Quantity)
	item.PTypeStr = ProductTypeMsgs[item.PType]
	item.Created = item.CreatedAt.Format(TIME_FORMAT)
	item.Updated = item.UpdatedAt.Format(TIME_FORMAT)

	return
}

type SpecialProduct struct {
	gorm.Model
	Name     string          `json:"name"`
	PType    ProductType     `json:"type"`
	Quantity uint64          `json:"quantity"`
	Discount decimal.Decimal `json:"discount"`
	Sold     uint64          `json:"sold"`
	Reserve  uint64          `json:"reserve"`

	PTypeStr    string `json:"product_type_str" gorm:"-"`
	QuantityStr string `json:"quantity_str" gorm:"-"`
	Created     string `json:"created_at" gorm:"-"`
	Updated     string `json:"updated_at" gorm:"-"`

	ProductID uint `json:"-"`
	Product   Product
}

func (item *SpecialProduct) AfterFind(tx *gorm.DB) (err error) {
	item.QuantityStr = ByteSize(item.Quantity)
	item.PTypeStr = ProductTypeMsgs[item.PType]
	item.Created = item.CreatedAt.Format(TIME_FORMAT)
	item.Updated = item.UpdatedAt.Format(TIME_FORMAT)

	return
}
