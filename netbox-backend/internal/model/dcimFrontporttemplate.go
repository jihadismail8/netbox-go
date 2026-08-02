package model

import (
	"time"
)

type DcimFrontporttemplate struct {
	Created          *time.Time `gorm:"column:created;type:timestamptz" json:"created"`
	LastUpdated      *time.Time `gorm:"column:last_updated;type:timestamptz" json:"lastUpdated"`
	ID               uint64     `gorm:"column:id;type:int8;primary_key" json:"id"`
	Name             string     `gorm:"column:name;type:varchar(64);not null" json:"name"`
	Label            string     `gorm:"column:label;type:varchar(64);not null" json:"label"`
	Description      string     `gorm:"column:description;type:varchar(200);not null" json:"description"`
	Type             string     `gorm:"column:type;type:varchar(50);not null" json:"type"`
	RearPortPosition int        `gorm:"column:rear_port_position;type:int2;not null" json:"rearPortPosition"`
	DeviceTypeID     int64      `gorm:"column:device_type_id;type:int8" json:"deviceTypeID"`
	RearPortID       int64      `gorm:"column:rear_port_id;type:int8;not null" json:"rearPortID"`
	Color            string     `gorm:"column:color;type:varchar(6);not null" json:"color"`
	ModuleTypeID     int64      `gorm:"column:module_type_id;type:int8" json:"moduleTypeID"`
}

// TableName table name
func (m *DcimFrontporttemplate) TableName() string {
	return "dcim_frontporttemplate"
}

// DcimFrontporttemplateColumnNames Whitelist for custom query fields to prevent sql injection attacks
var DcimFrontporttemplateColumnNames = map[string]bool{
	"created":            true,
	"last_updated":       true,
	"id":                 true,
	"name":               true,
	"label":              true,
	"description":        true,
	"type":               true,
	"rear_port_position": true,
	"device_type_id":     true,
	"rear_port_id":       true,
	"color":              true,
	"module_type_id":     true,
}
