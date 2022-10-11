package dataservice

import (
	"gorm.io/gorm"
	"time"
)

type UserAction struct {
	ID        uint      `json:"-" gorm:"primaryKey"`
	Email     string    `json:"email" gorm:"index"`
	Action    string    `json:"action"`
	IP        string    `json:"ip"`
	CreatedAt time.Time `json:"-"`
	UserID    uint      `json:"-"`

	Created string `json:"created_at" gorm:"-"`
}

func (u *UserAction) AfterFind(tx *gorm.DB) (err error) {
	u.Created = u.CreatedAt.Format(TIME_FORMAT)
	return
}

func (s *DataService) FindUserActions(userID uint, offset int64, limit int64) (total int64, items []*UserAction, err error) {
	err = s.Model(&UserAction{}).Where("user_id = ?", userID).Order("id DESC").Count(&total).Offset(int(offset)).Limit(int(limit)).Find(&items).Error
	return
}
