package models

import (
	"gorm.io/gorm"
	"time"
)

type UserActionType uint

const (
	UserActionType_Login = iota
	UserActionType_Forgot
	UserActionType_Reset
)

var UserActionTypeMsgs = []string{
	"Login",
	"Forgot",
	"Reset",
}

type UserAction struct {
	ID         uint `json:"id" gorm:"primarykey"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
	DeletedAt  gorm.DeletedAt `gorm:"index"`
	ActionType UserActionType `json:"action_type"`
	Email      string         `json:"email" gorm:"index"`
	IP         string         `json:"ip"`

	UserID uint `json:"-"`
	User   User

	ActionTypeStr string `json:"action" gorm:"-"`
	Created       string `json:"created_at" gorm:"-"`
	Updated       string `json:"updated_at" gorm:"-"`
}

func (item *UserAction) AfterFind(tx *gorm.DB) (err error) {
	item.Created = item.CreatedAt.Format(TIME_FORMAT)
	item.Updated = item.UpdatedAt.Format(TIME_FORMAT)
	item.ActionTypeStr = UserActionTypeMsgs[item.ActionType]
	return
}
