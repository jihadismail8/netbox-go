package model

import (
	"github.com/go-dev-frame/sponge/pkg/sgorm"
	"time"
)

type SocialAuthCode struct {
	ID        uint64      `gorm:"column:id;type:int8;primary_key" json:"id"`
	Email     string      `gorm:"column:email;type:varchar(254);not null" json:"email"`
	Code      string      `gorm:"column:code;type:varchar(32);not null" json:"code"`
	Verified  *sgorm.Bool `gorm:"column:verified;type:bool;not null" json:"verified"`
	Timestamp *time.Time  `gorm:"column:timestamp;type:timestamptz;not null" json:"timestamp"`
}

// TableName table name
func (m *SocialAuthCode) TableName() string {
	return "social_auth_code"
}

// SocialAuthCodeColumnNames Whitelist for custom query fields to prevent sql injection attacks
var SocialAuthCodeColumnNames = map[string]bool{
	"id":        true,
	"email":     true,
	"code":      true,
	"verified":  true,
	"timestamp": true,
}
