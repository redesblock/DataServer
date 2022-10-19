package dataservice

import (
	"gorm.io/gorm"
	"time"
)

type User struct {
	ID           uint      `json:"-" gorm:"primaryKey"`
	Email        string    `json:"email" gorm:"unique"`
	Password     string    `json:"-"`
	FirstName    string    `json:"first_name"`
	LastName     string    `json:"last_name"`
	UsedStorage  uint64    `json:"used_storage"`
	TotalStorage uint64    `json:"total_storage"`
	UsedTraffic  uint64    `json:"used_traffic"`
	TotalTraffic uint64    `json:"total_traffic"`
	UpdatedAt    time.Time `json:"-"`
	CreatedAt    time.Time `json:"-"`

	UsedStorageStr  string `json:"used_storage_str" gorm:"-"`
	TotalStorageStr string `json:"total_storage_str" gorm:"-"`
	UsedTrafficStr  string `json:"used_traffic_str" gorm:"-"`
	TotalTrafficStr string `json:"total_traffic_str" gorm:"-"`
	Created         string `json:"created_at" gorm:"-"`
	Updated         string `json:"updated_at" gorm:"-"`
}

func (u *User) AfterFind(tx *gorm.DB) (err error) {
	u.Created = u.CreatedAt.Format(TIME_FORMAT)
	u.Updated = u.UpdatedAt.Format(TIME_FORMAT)
	u.TotalStorageStr = ByteSize(u.TotalStorage)
	u.UsedStorageStr = ByteSize(u.UsedStorage)
	u.TotalTrafficStr = ByteSize(u.TotalTraffic)
	u.UsedTrafficStr = ByteSize(u.UsedTraffic)
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
