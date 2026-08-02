package model

import (
	"gorm.io/datatypes"
	"time"
)

type CircuitsCircuitgroupassignment struct {
	ID              uint64          `gorm:"column:id;type:int8;primary_key" json:"id"`
	Created         *time.Time      `gorm:"column:created;type:timestamptz" json:"created"`
	LastUpdated     *time.Time      `gorm:"column:last_updated;type:timestamptz" json:"lastUpdated"`
	CustomFieldData *datatypes.JSON `gorm:"column:custom_field_data;type:jsonb;not null" json:"customFieldData"`
	Priority        string          `gorm:"column:priority;type:varchar(50)" json:"priority"`
	MemberID        int64           `gorm:"column:member_id;type:int8;not null" json:"memberID"`
	GroupID         int64           `gorm:"column:group_id;type:int8;not null" json:"groupID"`
	MemberTypeID    int             `gorm:"column:member_type_id;type:int4;not null" json:"memberTypeID"`
}

// TableName table name
func (m *CircuitsCircuitgroupassignment) TableName() string {
	return "circuits_circuitgroupassignment"
}

// CircuitsCircuitgroupassignmentColumnNames Whitelist for custom query fields to prevent sql injection attacks
var CircuitsCircuitgroupassignmentColumnNames = map[string]bool{
	"id":                true,
	"created":           true,
	"last_updated":      true,
	"custom_field_data": true,
	"priority":          true,
	"member_id":         true,
	"group_id":          true,
	"member_type_id":    true,
}
