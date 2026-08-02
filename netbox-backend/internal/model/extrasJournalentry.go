package model

import (
	"gorm.io/datatypes"
	"time"
)

type ExtrasJournalentry struct {
	LastUpdated          *time.Time      `gorm:"column:last_updated;type:timestamptz" json:"lastUpdated"`
	ID                   uint64          `gorm:"column:id;type:int8;primary_key" json:"id"`
	AssignedObjectID     int64           `gorm:"column:assigned_object_id;type:int8;not null" json:"assignedObjectID"`
	Created              *time.Time      `gorm:"column:created;type:timestamptz" json:"created"`
	Kind                 string          `gorm:"column:kind;type:varchar(30);not null" json:"kind"`
	Comments             string          `gorm:"column:comments;type:text;not null" json:"comments"`
	AssignedObjectTypeID int             `gorm:"column:assigned_object_type_id;type:int4;not null" json:"assignedObjectTypeID"`
	CreatedByID          int64           `gorm:"column:created_by_id;type:int8" json:"createdByID"`
	CustomFieldData      *datatypes.JSON `gorm:"column:custom_field_data;type:jsonb;not null" json:"customFieldData"`
}

// TableName table name
func (m *ExtrasJournalentry) TableName() string {
	return "extras_journalentry"
}

// ExtrasJournalentryColumnNames Whitelist for custom query fields to prevent sql injection attacks
var ExtrasJournalentryColumnNames = map[string]bool{
	"last_updated":            true,
	"id":                      true,
	"assigned_object_id":      true,
	"created":                 true,
	"kind":                    true,
	"comments":                true,
	"assigned_object_type_id": true,
	"created_by_id":           true,
	"custom_field_data":       true,
}
