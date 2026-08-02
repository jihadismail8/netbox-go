package model

import (
	"time"
)

type ExtrasNotification struct {
	ID           uint64     `gorm:"column:id;type:int8;primary_key" json:"id"`
	Created      *time.Time `gorm:"column:created;type:timestamptz;not null" json:"created"`
	Read         *time.Time `gorm:"column:read;type:timestamptz" json:"read"`
	ObjectID     int64      `gorm:"column:object_id;type:int8;not null" json:"objectID"`
	EventType    string     `gorm:"column:event_type;type:varchar(50);not null" json:"eventType"`
	ObjectTypeID int        `gorm:"column:object_type_id;type:int4;not null" json:"objectTypeID"`
	ObjectRepr   string     `gorm:"column:object_repr;type:varchar(200);not null" json:"objectRepr"`
	UserID       int64      `gorm:"column:user_id;type:int8;not null" json:"userID"`
}

// TableName table name
func (m *ExtrasNotification) TableName() string {
	return "extras_notification"
}

// ExtrasNotificationColumnNames Whitelist for custom query fields to prevent sql injection attacks
var ExtrasNotificationColumnNames = map[string]bool{
	"id":             true,
	"created":        true,
	"read":           true,
	"object_id":      true,
	"event_type":     true,
	"object_type_id": true,
	"object_repr":    true,
	"user_id":        true,
}
