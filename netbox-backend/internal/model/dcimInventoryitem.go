package model

import (
	"github.com/go-dev-frame/sponge/pkg/sgorm"
	"gorm.io/datatypes"
	"time"
)

type DcimInventoryitem struct {
	Created         *time.Time      `gorm:"column:created;type:timestamptz" json:"created"`
	LastUpdated     *time.Time      `gorm:"column:last_updated;type:timestamptz" json:"lastUpdated"`
	CustomFieldData *datatypes.JSON `gorm:"column:custom_field_data;type:jsonb;not null" json:"customFieldData"`
	ID              uint64          `gorm:"column:id;type:int8;primary_key" json:"id"`
	Name            string          `gorm:"column:name;type:varchar(64);not null" json:"name"`
	Label           string          `gorm:"column:label;type:varchar(64);not null" json:"label"`
	Description     string          `gorm:"column:description;type:varchar(200);not null" json:"description"`
	PartID          string          `gorm:"column:part_id;type:varchar(50);not null" json:"partID"`
	Serial          string          `gorm:"column:serial;type:varchar(50);not null" json:"serial"`
	AssetTag        string          `gorm:"column:asset_tag;type:varchar(50)" json:"assetTag"`
	Discovered      *sgorm.Bool     `gorm:"column:discovered;type:bool;not null" json:"discovered"`
	Lft             int             `gorm:"column:lft;type:int4;not null" json:"lft"`
	Rght            int             `gorm:"column:rght;type:int4;not null" json:"rght"`
	TreeID          int             `gorm:"column:tree_id;type:int4;not null" json:"treeID"`
	Level           int             `gorm:"column:level;type:int4;not null" json:"level"`
	DeviceID        int64           `gorm:"column:device_id;type:int8;not null" json:"deviceID"`
	ManufacturerID  int64           `gorm:"column:manufacturer_id;type:int8" json:"manufacturerID"`
	ParentID        int64           `gorm:"column:parent_id;type:int8" json:"parentID"`
	RoleID          int64           `gorm:"column:role_id;type:int8" json:"roleID"`
	ComponentID     int64           `gorm:"column:component_id;type:int8" json:"componentID"`
	ComponentTypeID int             `gorm:"column:component_type_id;type:int4" json:"componentTypeID"`
	Status          string          `gorm:"column:status;type:varchar(50);not null" json:"status"`
	XLocationID     int64           `gorm:"column:_location_id;type:int8" json:"XLocationID"`
	XRackID         int64           `gorm:"column:_rack_id;type:int8" json:"XRackID"`
	XSiteID         int64           `gorm:"column:_site_id;type:int8" json:"XSiteID"`
}

// TableName table name
func (m *DcimInventoryitem) TableName() string {
	return "dcim_inventoryitem"
}

// DcimInventoryitemColumnNames Whitelist for custom query fields to prevent sql injection attacks
var DcimInventoryitemColumnNames = map[string]bool{
	"created":           true,
	"last_updated":      true,
	"custom_field_data": true,
	"id":                true,
	"name":              true,
	"label":             true,
	"description":       true,
	"part_id":           true,
	"serial":            true,
	"asset_tag":         true,
	"discovered":        true,
	"lft":               true,
	"rght":              true,
	"tree_id":           true,
	"level":             true,
	"device_id":         true,
	"manufacturer_id":   true,
	"parent_id":         true,
	"role_id":           true,
	"component_id":      true,
	"component_type_id": true,
	"status":            true,
	"_location_id":      true,
	"_rack_id":          true,
	"_site_id":          true,
}
