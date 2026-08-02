package model

import (
	"gorm.io/datatypes"
	"time"
)

type TenancyContactassignment struct {
	ID              uint64          `gorm:"column:id;type:int8;primary_key" json:"id"`
	Created         *time.Time      `gorm:"column:created;type:timestamptz" json:"created"`
	LastUpdated     *time.Time      `gorm:"column:last_updated;type:timestamptz" json:"lastUpdated"`
	ObjectID        int64           `gorm:"column:object_id;type:int8;not null" json:"objectID"`
	Priority        string          `gorm:"column:priority;type:varchar(50)" json:"priority"`
	ContactID       int64           `gorm:"column:contact_id;type:int8;not null" json:"contactID"`
	ObjectTypeID    int             `gorm:"column:object_type_id;type:int4;not null" json:"objectTypeID"`
	RoleID          int64           `gorm:"column:role_id;type:int8;not null" json:"roleID"`
	CustomFieldData *datatypes.JSON `gorm:"column:custom_field_data;type:jsonb;not null" json:"customFieldData"`
}

// TableName table name
func (m *TenancyContactassignment) TableName() string {
	return "tenancy_contactassignment"
}

// TenancyContactassignmentColumnNames Whitelist for custom query fields to prevent sql injection attacks
var TenancyContactassignmentColumnNames = map[string]bool{
	"id":                true,
	"created":           true,
	"last_updated":      true,
	"object_id":         true,
	"priority":          true,
	"contact_id":        true,
	"object_type_id":    true,
	"role_id":           true,
	"custom_field_data": true,
}
