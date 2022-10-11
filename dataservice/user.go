package dataservice

import (
	"gorm.io/gorm"
	"time"
)

type User struct {
	ID        uint      `json:"-" gorm:"primaryKey"`
	Email     string    `json:"email" gorm:"unique"`
	Password  string    `json:"-"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	UpdatedAt time.Time `json:"-"`
	CreatedAt time.Time `json:"-"`

	Created string `json:"created_at" gorm:"-"`
	Updated string `json:"updated_at" gorm:"-"`
}

func (u *User) AfterFind(tx *gorm.DB) (err error) {
	u.Created = u.CreatedAt.Format(TIME_FORMAT)
	u.Updated = u.UpdatedAt.Format(TIME_FORMAT)
	return
}

func (s *DataService) FindUserByEmail(email string) (item *User, err error) {
	ret := s.Model(&User{}).Where("email = ?", email).Find(&item)
	if ret.Error != nil {
		err = ret.Error
	}
	if ret.RowsAffected == 0 {
		item = nil
	}
	return
}

func (s *DataService) FindUserByID(id uint) (item *User, err error) {
	ret := s.Model(&User{}).Where("id = ?", id).Find(&item)
	if ret.Error != nil {
		err = ret.Error
	}
	if ret.RowsAffected == 0 {
		item = nil
	}
	return
}
