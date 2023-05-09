package models

import "gorm.io/gorm"

type DailyUsedStorage struct {
	ID     uint   `json:"-" gorm:"primaryKey"`
	Num    uint64 `json:"-"`
	Time   string `json:"timestamp"`
	UserID uint   `json:"-"`
	NumStr uint64 `json:"num" gorm:"-"`
}

func (item *DailyUsedStorage) AfterFind(tx *gorm.DB) (err error) {
	item.NumStr = item.Num / 1024
	return
}

type DailyUsedTraffic struct {
	ID     uint   `json:"-" gorm:"primaryKey"`
	Num    uint64 `json:"-"`
	Time   string `json:"timestamp"`
	UserID uint   `json:"-"`

	NumStr uint64 `json:"num" gorm:"-"`
}

func (item *DailyUsedTraffic) AfterFind(tx *gorm.DB) (err error) {
	item.NumStr = item.Num / 1024
	return
}
