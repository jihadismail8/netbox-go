package model

import (
	"time"
)

type DcimRearporttemplate struct {
	Created      *time.Time `gorm:"column:created;type:timestamptz" json:"created"`
	LastUpdated  *time.Time `gorm:"column:last_updated;type:timestamptz" json:"lastUpdated"`
	ID           uint64     `gorm:"column:id;type:int8;primary_key" json:"id"`
	Name         string     `gorm:"column:name;type:varchar(64);not null" json:"name"`
	Label        string     `gorm:"column:label;type:varchar(64);not null" json:"label"`
	Description  string     `gorm:"column:description;type:varchar(200);not null" json:"description"`
	Type         string     `gorm:"column:type;type:varchar(50);not null" json:"type"`
	Positions    int        `gorm:"column:positions;type:int2;not null" json:"positions"`
	DeviceTypeID int64      `gorm:"column:device_type_id;type:int8" json:"deviceTypeID"`
	Color        string     `gorm:"column:color;type:varchar(6);not null" json:"color"`
	ModuleTypeID int64      `gorm:"column:module_type_id;type:int8" json:"moduleTypeID"`
}

// TableName table name
func (m *DcimRearporttemplate) TableName() string {
	return "dcim_rearporttemplate"
}

// DcimRearporttemplateColumnNames Whitelist for custom query fields to prevent sql injection attacks
var DcimRearporttemplateColumnNames = map[string]bool{
	"created":        true,
	"last_updated":   true,
	"id":             true,
	"name":           true,
	"label":          true,
	"description":    true,
	"type":           true,
	"positions":      true,
	"device_type_id": true,
	"color":          true,
	"module_type_id": true,
}
