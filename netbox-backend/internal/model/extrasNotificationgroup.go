package model

import (
	"time"
)

type ExtrasNotificationgroup struct {
	ID          uint64     `gorm:"column:id;type:int8;primary_key" json:"id"`
	Created     *time.Time `gorm:"column:created;type:timestamptz" json:"created"`
	LastUpdated *time.Time `gorm:"column:last_updated;type:timestamptz" json:"lastUpdated"`
	Name        string     `gorm:"column:name;type:varchar(100);not null" json:"name"`
	Description string     `gorm:"column:description;type:varchar(200);not null" json:"description"`
}

// TableName table name
func (m *ExtrasNotificationgroup) TableName() string {
	return "extras_notificationgroup"
}

// ExtrasNotificationgroupColumnNames Whitelist for custom query fields to prevent sql injection attacks
var ExtrasNotificationgroupColumnNames = map[string]bool{
	"id":           true,
	"created":      true,
	"last_updated": true,
	"name":         true,
	"description":  true,
}
