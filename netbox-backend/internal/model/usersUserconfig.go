package model

import (
	"gorm.io/datatypes"
)

type UsersUserconfig struct {
	ID     uint64          `gorm:"column:id;type:int8;primary_key" json:"id"`
	Data   *datatypes.JSON `gorm:"column:data;type:jsonb;not null" json:"data"`
	UserID int64           `gorm:"column:user_id;type:int8;not null" json:"userID"`
}

// TableName table name
func (m *UsersUserconfig) TableName() string {
	return "users_userconfig"
}

// UsersUserconfigColumnNames Whitelist for custom query fields to prevent sql injection attacks
var UsersUserconfigColumnNames = map[string]bool{
	"id":      true,
	"data":    true,
	"user_id": true,
}
