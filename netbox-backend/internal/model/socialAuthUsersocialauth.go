package model

import (
	"gorm.io/datatypes"
	"time"
)

type SocialAuthUsersocialauth struct {
	ID        uint64          `gorm:"column:id;type:int8;primary_key" json:"id"`
	Provider  string          `gorm:"column:provider;type:varchar(32);not null" json:"provider"`
	UID       string          `gorm:"column:uid;type:varchar(255);not null" json:"uid"`
	UserID    int64           `gorm:"column:user_id;type:int8;not null" json:"userID"`
	Created   *time.Time      `gorm:"column:created;type:timestamptz;not null" json:"created"`
	Modified  *time.Time      `gorm:"column:modified;type:timestamptz;not null" json:"modified"`
	ExtraData *datatypes.JSON `gorm:"column:extra_data;type:jsonb;not null" json:"extraData"`
}

// TableName table name
func (m *SocialAuthUsersocialauth) TableName() string {
	return "social_auth_usersocialauth"
}

// SocialAuthUsersocialauthColumnNames Whitelist for custom query fields to prevent sql injection attacks
var SocialAuthUsersocialauthColumnNames = map[string]bool{
	"id":         true,
	"provider":   true,
	"uid":        true,
	"user_id":    true,
	"created":    true,
	"modified":   true,
	"extra_data": true,
}
