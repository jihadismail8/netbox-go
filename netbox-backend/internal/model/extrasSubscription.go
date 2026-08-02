package model

import (
	"time"
)

type ExtrasSubscription struct {
	ID           uint64     `gorm:"column:id;type:int8;primary_key" json:"id"`
	Created      *time.Time `gorm:"column:created;type:timestamptz;not null" json:"created"`
	ObjectID     int64      `gorm:"column:object_id;type:int8;not null" json:"objectID"`
	ObjectTypeID int        `gorm:"column:object_type_id;type:int4;not null" json:"objectTypeID"`
	UserID       int64      `gorm:"column:user_id;type:int8;not null" json:"userID"`
}

// TableName table name
func (m *ExtrasSubscription) TableName() string {
	return "extras_subscription"
}

// ExtrasSubscriptionColumnNames Whitelist for custom query fields to prevent sql injection attacks
var ExtrasSubscriptionColumnNames = map[string]bool{
	"id":             true,
	"created":        true,
	"object_id":      true,
	"object_type_id": true,
	"user_id":        true,
}
