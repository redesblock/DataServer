package models

import "gorm.io/gorm"

type UserRole uint

const (
	UserRole_Admin UserRole = iota
	UserRole_Oper
	UserRole_User
)

var UserRoleMsgs = []string{
	"Admin",
	"Operator",
	"User",
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
	Email     string     `json:"email" gorm:"unique"`
	Password  string     `json:"-"`
	FirstName string     `json:"first_name"`
	LastName  string     `json:"last_name"`
	Role      UserRole   `json:"role"`
	Status    UserStatus `json:"status"`

	RoleStr   string `json:"role_str" gorm:"-"`
	StatusStr string `json:"status_str" gorm:"-"`
	Created   string `json:"created_at" gorm:"-"`
	Updated   string `json:"updated_at" gorm:"-"`
}

func (item *User) AfterFind(tx *gorm.DB) (err error) {
	item.RoleStr = UserRoleMsgs[item.Role]
	item.StatusStr = UserStatusMsgs[item.Status]
	item.Created = item.CreatedAt.Format(TIME_FORMAT)
	item.Updated = item.UpdatedAt.Format(TIME_FORMAT)
	return
}
