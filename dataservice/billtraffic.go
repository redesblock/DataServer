package dataservice

import (
	"gorm.io/gorm"
	"time"
)

type BillTraffic struct {
	ID uint `json:"-" gorm:"primaryKey"`
	//Email       string    `json:"email" gorm:"index"`
	Hash        string    `json:"hash"`
	Amount      string    `json:"amount"`
	Size        uint64    `json:"size"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"-"`
	UserID      uint      `json:"-"`

	Created string `json:"created_at" gorm:"-"`
	SizeStr string `json:"size_str" gorm:"-"`
	URL     string `json:"url" gorm:"-"`
}

func (u *BillTraffic) AfterFind(tx *gorm.DB) (err error) {
	u.Created = u.CreatedAt.Format(TIME_FORMAT)
	u.SizeStr = HumanateBytes(u.Size)
	u.URL = "https://testnet.bscscan.com/tx/" + u.Hash
	return
}

func (s *DataService) FindBillsTraffic(userID uint, offset int64, limit int64) (total int64, items []*BillTraffic, err error) {
	err = s.Model(&BillTraffic{}).Where("user_id = ?", userID).Order("id DESC").Count(&total).Offset(int(offset)).Limit(int(limit)).Find(&items).Error
	return
}
