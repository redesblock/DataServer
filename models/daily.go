package models

import "gorm.io/gorm"

type UsedStorage struct {
	ID     uint   `json:"-" gorm:"primaryKey"`
	Num    uint64 `json:"-"`
	Time   string `json:"timestamp"`
	UserID uint   `json:"-"`
	NumStr uint64 `json:"num" gorm:"-"`
}

func (item *UsedStorage) AfterFind(tx *gorm.DB) (err error) {
	item.NumStr = item.Num / 1024
	return
}

type UsedTraffic struct {
	ID     uint   `json:"-" gorm:"primaryKey"`
	Num    uint64 `json:"-"`
	Time   string `json:"timestamp"`
	UserID uint   `json:"-"`

	NumStr uint64 `json:"num" gorm:"-"`
}

func (item *UsedTraffic) AfterFind(tx *gorm.DB) (err error) {
	item.NumStr = item.Num / 1024
	return
}
