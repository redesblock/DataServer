package models

import "gorm.io/gorm"

type UserAction struct {
	gorm.Model
	Action string `json:"action"`
	IP     string `json:"ip"`

	UserID uint `json:"-"`
	User   User

	Created string `json:"created_at" gorm:"-"`
	Updated string `json:"updated_at" gorm:"-"`
}

func (item *UserAction) AfterFind(tx *gorm.DB) (err error) {
	item.Created = item.CreatedAt.Format(TIME_FORMAT)
	item.Updated = item.UpdatedAt.Format(TIME_FORMAT)
	return
}
