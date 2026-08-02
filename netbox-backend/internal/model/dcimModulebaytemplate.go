package model

import (
	"time"
)

type DcimModulebaytemplate struct {
	ID           uint64     `gorm:"column:id;type:int8;primary_key" json:"id"`
	Created      *time.Time `gorm:"column:created;type:timestamptz" json:"created"`
	LastUpdated  *time.Time `gorm:"column:last_updated;type:timestamptz" json:"lastUpdated"`
	Name         string     `gorm:"column:name;type:varchar(64);not null" json:"name"`
	Label        string     `gorm:"column:label;type:varchar(64);not null" json:"label"`
	Position     string     `gorm:"column:position;type:varchar(30);not null" json:"position"`
	Description  string     `gorm:"column:description;type:varchar(200);not null" json:"description"`
	DeviceTypeID int64      `gorm:"column:device_type_id;type:int8" json:"deviceTypeID"`
	ModuleTypeID int64      `gorm:"column:module_type_id;type:int8" json:"moduleTypeID"`
}

// TableName table name
func (m *DcimModulebaytemplate) TableName() string {
	return "dcim_modulebaytemplate"
}

// DcimModulebaytemplateColumnNames Whitelist for custom query fields to prevent sql injection attacks
var DcimModulebaytemplateColumnNames = map[string]bool{
	"id":             true,
	"created":        true,
	"last_updated":   true,
	"name":           true,
	"label":          true,
	"position":       true,
	"description":    true,
	"device_type_id": true,
	"module_type_id": true,
}
