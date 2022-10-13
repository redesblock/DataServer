package dataservice

import (
	"gorm.io/gorm"
)

type UsedStorage struct {
	ID     uint   `json:"-" gorm:"primaryKey"`
	Num    uint64 `json:"num"`
	Time   string `json:"timestamp"`
	UserID uint   `json:"-"`
}

func (u *UsedStorage) AfterFind(tx *gorm.DB) (err error) {
	u.Num /= 1024 * 1024
	return
}

type UsedTraffic struct {
	ID     uint   `json:"-" gorm:"primaryKey"`
	Num    uint64 `json:"num"`
	Time   string `json:"timestamp"`
	UserID uint   `json:"-"`
}

func (u *UsedTraffic) AfterFind(tx *gorm.DB) (err error) {
	u.Num /= 1024 * 1024
	return
}
