package dataservice

import (
	"gorm.io/gorm"
)

type UsedStorage struct {
	ID     uint   `json:"-" gorm:"primaryKey"`
	Num    uint64 `json:"-"`
	Time   string `json:"timestamp"`
	UserID uint   `json:"-"`
	NumStr uint64 `json:"num" gorm:"-"`
}

func (u *UsedStorage) AfterFind(tx *gorm.DB) (err error) {
	u.NumStr = u.Num / 1024
	return
}

type UsedTraffic struct {
	ID     uint   `json:"-" gorm:"primaryKey"`
	Num    uint64 `json:"-"`
	Time   string `json:"timestamp"`
	UserID uint   `json:"-"`

	NumStr uint64 `json:"num" gorm:"-"`
}

func (u *UsedTraffic) AfterFind(tx *gorm.DB) (err error) {
	u.NumStr = u.Num / 1024
	return
}
