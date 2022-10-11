package dataservice

import (
	"github.com/dustin/go-humanize"
	"gorm.io/gorm"
	"time"
)

type BillStorage struct {
	ID uint `json:"-" gorm:"primaryKey"`
	//Email       string    `json:"email" gorm:"index"`
	Hash        string    `json:"hash" gorm:"unique"`
	Amount      string    `json:"amount"`
	Size        uint64    `json:"size"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"-"`
	UserID      uint      `json:"-"`

	Created string `json:"created_at" gorm:"-"`
	SizeStr string `json:"size_str" gorm:"-"`
}

func (u *BillStorage) AfterFind(tx *gorm.DB) (err error) {
	u.Created = u.CreatedAt.Format(TIME_FORMAT)
	u.SizeStr = humanize.Bytes(u.Size)
	return
}

func (s *DataService) FindBillsStorage(userID uint, offset int64, limit int64) (total int64, items []*BillTraffic, err error) {
	err = s.Model(&BillStorage{}).Where("user_id = ?", userID).Order("id DESC").Count(&total).Offset(int(offset)).Limit(int(limit)).Find(&items).Error
	return
}
