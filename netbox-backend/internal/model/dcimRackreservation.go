package model

import (
	"gorm.io/datatypes"
	"time"
)

type DcimRackreservation struct {
	Created         *time.Time      `gorm:"column:created;type:timestamptz" json:"created"`
	LastUpdated     *time.Time      `gorm:"column:last_updated;type:timestamptz" json:"lastUpdated"`
	CustomFieldData *datatypes.JSON `gorm:"column:custom_field_data;type:jsonb;not null" json:"customFieldData"`
	ID              uint64          `gorm:"column:id;type:int8;primary_key" json:"id"`
	Units           string          `gorm:"column:units;type:_int2;not null" json:"units"`
	Description     string          `gorm:"column:description;type:varchar(200);not null" json:"description"`
	RackID          int64           `gorm:"column:rack_id;type:int8;not null" json:"rackID"`
	TenantID        int64           `gorm:"column:tenant_id;type:int8" json:"tenantID"`
	UserID          int64           `gorm:"column:user_id;type:int8;not null" json:"userID"`
	Comments        string          `gorm:"column:comments;type:text;not null" json:"comments"`
	Status          string          `gorm:"column:status;type:varchar(50);not null" json:"status"`
}

// TableName table name
func (m *DcimRackreservation) TableName() string {
	return "dcim_rackreservation"
}

// DcimRackreservationColumnNames Whitelist for custom query fields to prevent sql injection attacks
var DcimRackreservationColumnNames = map[string]bool{
	"created":           true,
	"last_updated":      true,
	"custom_field_data": true,
	"id":                true,
	"units":             true,
	"description":       true,
	"rack_id":           true,
	"tenant_id":         true,
	"user_id":           true,
	"comments":          true,
	"status":            true,
}
