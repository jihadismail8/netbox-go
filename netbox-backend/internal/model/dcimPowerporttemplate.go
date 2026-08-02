package model

import (
	"time"
)

type DcimPowerporttemplate struct {
	Created       *time.Time `gorm:"column:created;type:timestamptz" json:"created"`
	LastUpdated   *time.Time `gorm:"column:last_updated;type:timestamptz" json:"lastUpdated"`
	ID            uint64     `gorm:"column:id;type:int8;primary_key" json:"id"`
	Name          string     `gorm:"column:name;type:varchar(64);not null" json:"name"`
	Label         string     `gorm:"column:label;type:varchar(64);not null" json:"label"`
	Description   string     `gorm:"column:description;type:varchar(200);not null" json:"description"`
	Type          string     `gorm:"column:type;type:varchar(50)" json:"type"`
	MaximumDraw   int        `gorm:"column:maximum_draw;type:int4" json:"maximumDraw"`
	AllocatedDraw int        `gorm:"column:allocated_draw;type:int4" json:"allocatedDraw"`
	DeviceTypeID  int64      `gorm:"column:device_type_id;type:int8" json:"deviceTypeID"`
	ModuleTypeID  int64      `gorm:"column:module_type_id;type:int8" json:"moduleTypeID"`
}

// TableName table name
func (m *DcimPowerporttemplate) TableName() string {
	return "dcim_powerporttemplate"
}

// DcimPowerporttemplateColumnNames Whitelist for custom query fields to prevent sql injection attacks
var DcimPowerporttemplateColumnNames = map[string]bool{
	"created":        true,
	"last_updated":   true,
	"id":             true,
	"name":           true,
	"label":          true,
	"description":    true,
	"type":           true,
	"maximum_draw":   true,
	"allocated_draw": true,
	"device_type_id": true,
	"module_type_id": true,
}
