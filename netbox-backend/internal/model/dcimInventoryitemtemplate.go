package model

import (
	"time"
)

type DcimInventoryitemtemplate struct {
	ID              uint64     `gorm:"column:id;type:int8;primary_key" json:"id"`
	Created         *time.Time `gorm:"column:created;type:timestamptz" json:"created"`
	LastUpdated     *time.Time `gorm:"column:last_updated;type:timestamptz" json:"lastUpdated"`
	Name            string     `gorm:"column:name;type:varchar(64);not null" json:"name"`
	Label           string     `gorm:"column:label;type:varchar(64);not null" json:"label"`
	Description     string     `gorm:"column:description;type:varchar(200);not null" json:"description"`
	ComponentID     int64      `gorm:"column:component_id;type:int8" json:"componentID"`
	PartID          string     `gorm:"column:part_id;type:varchar(50);not null" json:"partID"`
	Lft             int        `gorm:"column:lft;type:int4;not null" json:"lft"`
	Rght            int        `gorm:"column:rght;type:int4;not null" json:"rght"`
	TreeID          int        `gorm:"column:tree_id;type:int4;not null" json:"treeID"`
	Level           int        `gorm:"column:level;type:int4;not null" json:"level"`
	ComponentTypeID int        `gorm:"column:component_type_id;type:int4" json:"componentTypeID"`
	DeviceTypeID    int64      `gorm:"column:device_type_id;type:int8;not null" json:"deviceTypeID"`
	ManufacturerID  int64      `gorm:"column:manufacturer_id;type:int8" json:"manufacturerID"`
	ParentID        int64      `gorm:"column:parent_id;type:int8" json:"parentID"`
	RoleID          int64      `gorm:"column:role_id;type:int8" json:"roleID"`
}

// TableName table name
func (m *DcimInventoryitemtemplate) TableName() string {
	return "dcim_inventoryitemtemplate"
}

// DcimInventoryitemtemplateColumnNames Whitelist for custom query fields to prevent sql injection attacks
var DcimInventoryitemtemplateColumnNames = map[string]bool{
	"id":                true,
	"created":           true,
	"last_updated":      true,
	"name":              true,
	"label":             true,
	"description":       true,
	"component_id":      true,
	"part_id":           true,
	"lft":               true,
	"rght":              true,
	"tree_id":           true,
	"level":             true,
	"component_type_id": true,
	"device_type_id":    true,
	"manufacturer_id":   true,
	"parent_id":         true,
	"role_id":           true,
}
