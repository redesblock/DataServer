package models

import "gorm.io/gorm"

type UserRole uint

const (
	UserRole_User UserRole = iota
	UserRole_Oper
	UserRole_Admin
)

var UserRoleMsgs = []string{
	"User",
	"Operator",
	"Admin",
}

type UserStatus uint

const (
	UserStaus_Normal UserStatus = iota
	UserStaus_Disabled
)

var UserStatusMsgs = []string{
	"Enable",
	"Disabled",
}

type User struct {
	gorm.Model
	Email        string     `json:"email" gorm:"unique"`
	Password     string     `json:"-"`
	FirstName    string     `json:"first_name"`
	LastName     string     `json:"last_name"`
	Role         UserRole   `json:"role" gorm:"default:0"`
	Status       UserStatus `json:"status" gorm:"default:0"`
	TotalStorage uint64     `json:"total_storage"`
	TotalTraffic uint64     `json:"total_traffic"`

	RoleStr         string `json:"role_str" gorm:"-"`
	StatusStr       string `json:"status_str" gorm:"-"`
	TotalStorageStr string `json:"total_storage_str" gorm:"-"`
	TotalTrafficStr string `json:"total_traffic_str" gorm:"-"`
	Created         string `json:"created_at" gorm:"-"`
	Updated         string `json:"updated_at" gorm:"-"`
}

func (item *User) AfterFind(tx *gorm.DB) (err error) {
	item.RoleStr = UserRoleMsgs[item.Role]
	item.StatusStr = UserStatusMsgs[item.Status]
	item.TotalStorageStr = ByteSize(item.TotalStorage)
	item.TotalTrafficStr = ByteSize(item.TotalTraffic)
	item.Created = item.CreatedAt.Format(TIME_FORMAT)
	item.Updated = item.UpdatedAt.Format(TIME_FORMAT)
	return
}
