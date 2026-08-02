package model

import (
	"gorm.io/datatypes"
	"time"
)

type DcimModulebay struct {
	ID              uint64          `gorm:"column:id;type:int8;primary_key" json:"id"`
	Created         *time.Time      `gorm:"column:created;type:timestamptz" json:"created"`
	LastUpdated     *time.Time      `gorm:"column:last_updated;type:timestamptz" json:"lastUpdated"`
	CustomFieldData *datatypes.JSON `gorm:"column:custom_field_data;type:jsonb;not null" json:"customFieldData"`
	Name            string          `gorm:"column:name;type:varchar(64);not null" json:"name"`
	Label           string          `gorm:"column:label;type:varchar(64);not null" json:"label"`
	Position        string          `gorm:"column:position;type:varchar(30);not null" json:"position"`
	Description     string          `gorm:"column:description;type:varchar(200);not null" json:"description"`
	DeviceID        int64           `gorm:"column:device_id;type:int8;not null" json:"deviceID"`
	Level           int             `gorm:"column:level;type:int4;not null" json:"level"`
	Lft             int             `gorm:"column:lft;type:int4;not null" json:"lft"`
	ModuleID        int64           `gorm:"column:module_id;type:int8" json:"moduleID"`
	ParentID        int64           `gorm:"column:parent_id;type:int8" json:"parentID"`
	Rght            int             `gorm:"column:rght;type:int4;not null" json:"rght"`
	TreeID          int             `gorm:"column:tree_id;type:int4;not null" json:"treeID"`
	XLocationID     int64           `gorm:"column:_location_id;type:int8" json:"XLocationID"`
	XRackID         int64           `gorm:"column:_rack_id;type:int8" json:"XRackID"`
	XSiteID         int64           `gorm:"column:_site_id;type:int8" json:"XSiteID"`
}

// TableName table name
func (m *DcimModulebay) TableName() string {
	return "dcim_modulebay"
}

// DcimModulebayColumnNames Whitelist for custom query fields to prevent sql injection attacks
var DcimModulebayColumnNames = map[string]bool{
	"id":                true,
	"created":           true,
	"last_updated":      true,
	"custom_field_data": true,
	"name":              true,
	"label":             true,
	"position":          true,
	"description":       true,
	"device_id":         true,
	"level":             true,
	"lft":               true,
	"module_id":         true,
	"parent_id":         true,
	"rght":              true,
	"tree_id":           true,
	"_location_id":      true,
	"_rack_id":          true,
	"_site_id":          true,
}
