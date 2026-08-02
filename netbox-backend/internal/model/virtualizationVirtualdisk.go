package model

import (
	"gorm.io/datatypes"
	"time"
)

type VirtualizationVirtualdisk struct {
	ID               uint64          `gorm:"column:id;type:int8;primary_key" json:"id"`
	Created          *time.Time      `gorm:"column:created;type:timestamptz" json:"created"`
	LastUpdated      *time.Time      `gorm:"column:last_updated;type:timestamptz" json:"lastUpdated"`
	CustomFieldData  *datatypes.JSON `gorm:"column:custom_field_data;type:jsonb;not null" json:"customFieldData"`
	Name             string          `gorm:"column:name;type:varchar(64);not null" json:"name"`
	Description      string          `gorm:"column:description;type:varchar(200);not null" json:"description"`
	Size             int             `gorm:"column:size;type:int4;not null" json:"size"`
	VirtualMachineID int64           `gorm:"column:virtual_machine_id;type:int8;not null" json:"virtualMachineID"`
}

// TableName table name
func (m *VirtualizationVirtualdisk) TableName() string {
	return "virtualization_virtualdisk"
}

// VirtualizationVirtualdiskColumnNames Whitelist for custom query fields to prevent sql injection attacks
var VirtualizationVirtualdiskColumnNames = map[string]bool{
	"id":                 true,
	"created":            true,
	"last_updated":       true,
	"custom_field_data":  true,
	"name":               true,
	"description":        true,
	"size":               true,
	"virtual_machine_id": true,
}
