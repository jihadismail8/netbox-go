package model

import (
	"gorm.io/datatypes"
	"time"
)

type DcimModule struct {
	ID               uint64          `gorm:"column:id;type:int8;primary_key" json:"id"`
	Created          *time.Time      `gorm:"column:created;type:timestamptz" json:"created"`
	LastUpdated      *time.Time      `gorm:"column:last_updated;type:timestamptz" json:"lastUpdated"`
	CustomFieldData  *datatypes.JSON `gorm:"column:custom_field_data;type:jsonb;not null" json:"customFieldData"`
	LocalContextData *datatypes.JSON `gorm:"column:local_context_data;type:jsonb" json:"localContextData"`
	Serial           string          `gorm:"column:serial;type:varchar(50);not null" json:"serial"`
	AssetTag         string          `gorm:"column:asset_tag;type:varchar(50)" json:"assetTag"`
	Comments         string          `gorm:"column:comments;type:text;not null" json:"comments"`
	DeviceID         int64           `gorm:"column:device_id;type:int8;not null" json:"deviceID"`
	ModuleBayID      int64           `gorm:"column:module_bay_id;type:int8;not null" json:"moduleBayID"`
	ModuleTypeID     int64           `gorm:"column:module_type_id;type:int8;not null" json:"moduleTypeID"`
	Description      string          `gorm:"column:description;type:varchar(200);not null" json:"description"`
	Status           string          `gorm:"column:status;type:varchar(50);not null" json:"status"`
}

// TableName table name
func (m *DcimModule) TableName() string {
	return "dcim_module"
}

// DcimModuleColumnNames Whitelist for custom query fields to prevent sql injection attacks
var DcimModuleColumnNames = map[string]bool{
	"id":                 true,
	"created":            true,
	"last_updated":       true,
	"custom_field_data":  true,
	"local_context_data": true,
	"serial":             true,
	"asset_tag":          true,
	"comments":           true,
	"device_id":          true,
	"module_bay_id":      true,
	"module_type_id":     true,
	"description":        true,
	"status":             true,
}
