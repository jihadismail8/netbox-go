package model

import (
	"github.com/go-dev-frame/sponge/pkg/sgorm"
	"time"
)

type UsersToken struct {
	ID           uint64      `gorm:"column:id;type:int8;primary_key" json:"id"`
	Created      *time.Time  `gorm:"column:created;type:timestamptz;not null" json:"created"`
	Expires      *time.Time  `gorm:"column:expires;type:timestamptz" json:"expires"`
	Key          string      `gorm:"column:key;type:varchar(40);not null" json:"key"`
	WriteEnabled *sgorm.Bool `gorm:"column:write_enabled;type:bool;not null" json:"writeEnabled"`
	Description  string      `gorm:"column:description;type:varchar(200);not null" json:"description"`
	UserID       int64       `gorm:"column:user_id;type:int8;not null" json:"userID"`
	AllowedIps   string      `gorm:"column:allowed_ips;type:_cidr" json:"allowedIps"`
	LastUsed     *time.Time  `gorm:"column:last_used;type:timestamptz" json:"lastUsed"`
}

// TableName table name
func (m *UsersToken) TableName() string {
	return "users_token"
}

// UsersTokenColumnNames Whitelist for custom query fields to prevent sql injection attacks
var UsersTokenColumnNames = map[string]bool{
	"id":            true,
	"created":       true,
	"expires":       true,
	"key":           true,
	"write_enabled": true,
	"description":   true,
	"user_id":       true,
	"allowed_ips":   true,
	"last_used":     true,
}
