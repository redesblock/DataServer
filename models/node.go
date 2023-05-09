package models

import (
	"gorm.io/gorm"
)

type Node struct {
	gorm.Model
	Name      string `json:"name"`
	IP        string `json:"ip"`
	Port      int    `json:"port"`
	Zone      string `json:"zone"`
	VoucherID string `json:"voucher_id" gorm:"unique"`
	Usable    bool   `json:"usable"`

	Created string `json:"created_at" gorm:"-"`
	Updated string `json:"updated_at" gorm:"-"`
}

func (item *Node) AfterFind(tx *gorm.DB) (err error) {
	item.Created = item.CreatedAt.Format(TIME_FORMAT)
	item.Updated = item.UpdatedAt.Format(TIME_FORMAT)
	return
}
