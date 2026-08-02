package model

import (
	"github.com/go-dev-frame/sponge/pkg/sgorm"
	"time"
)

type UsersUser struct {
	ID          uint64      `gorm:"column:id;type:int8;primary_key" json:"id"`
	Password    string      `gorm:"column:password;type:varchar(128);not null" json:"password"`
	LastLogin   *time.Time  `gorm:"column:last_login;type:timestamptz" json:"lastLogin"`
	IsSuperuser *sgorm.Bool `gorm:"column:is_superuser;type:bool;not null" json:"isSuperuser"`
	Username    string      `gorm:"column:username;type:varchar(150);not null" json:"username"`
	FirstName   string      `gorm:"column:first_name;type:varchar(150);not null" json:"firstName"`
	LastName    string      `gorm:"column:last_name;type:varchar(150);not null" json:"lastName"`
	Email       string      `gorm:"column:email;type:varchar(254);not null" json:"email"`
	IsStaff     *sgorm.Bool `gorm:"column:is_staff;type:bool;not null" json:"isStaff"`
	IsActive    *sgorm.Bool `gorm:"column:is_active;type:bool;not null" json:"isActive"`
	DateJoined  *time.Time  `gorm:"column:date_joined;type:timestamptz;not null" json:"dateJoined"`
}

// TableName table name
func (m *UsersUser) TableName() string {
	return "users_user"
}

// UsersUserColumnNames Whitelist for custom query fields to prevent sql injection attacks
var UsersUserColumnNames = map[string]bool{
	"id":           true,
	"password":     true,
	"last_login":   true,
	"is_superuser": true,
	"username":     true,
	"first_name":   true,
	"last_name":    true,
	"email":        true,
	"is_staff":     true,
	"is_active":    true,
	"date_joined":  true,
}
