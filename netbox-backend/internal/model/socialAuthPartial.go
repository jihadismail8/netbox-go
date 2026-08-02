package model

import (
	"gorm.io/datatypes"
	"time"
)

type SocialAuthPartial struct {
	ID        uint64          `gorm:"column:id;type:int8;primary_key" json:"id"`
	Token     string          `gorm:"column:token;type:varchar(32);not null" json:"token"`
	NextStep  int             `gorm:"column:next_step;type:int2;not null" json:"nextStep"`
	Backend   string          `gorm:"column:backend;type:varchar(32);not null" json:"backend"`
	Timestamp *time.Time      `gorm:"column:timestamp;type:timestamptz;not null" json:"timestamp"`
	Data      *datatypes.JSON `gorm:"column:data;type:jsonb;not null" json:"data"`
}

// TableName table name
func (m *SocialAuthPartial) TableName() string {
	return "social_auth_partial"
}

// SocialAuthPartialColumnNames Whitelist for custom query fields to prevent sql injection attacks
var SocialAuthPartialColumnNames = map[string]bool{
	"id":        true,
	"token":     true,
	"next_step": true,
	"backend":   true,
	"timestamp": true,
	"data":      true,
}
