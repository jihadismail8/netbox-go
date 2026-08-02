package model

import (
	"time"
)

type DjangoSession struct {
	SessionKey  string     `gorm:"column:session_key;type:varchar(40);primary_key" json:"sessionKey"`
	SessionData string     `gorm:"column:session_data;type:text;not null" json:"sessionData"`
	ExpireDate  *time.Time `gorm:"column:expire_date;type:timestamptz;not null" json:"expireDate"`
}

// TableName table name
func (m *DjangoSession) TableName() string {
	return "django_session"
}

// DjangoSessionColumnNames Whitelist for custom query fields to prevent sql injection attacks
var DjangoSessionColumnNames = map[string]bool{
	"session_key":  true,
	"session_data": true,
	"expire_date":  true,
}
