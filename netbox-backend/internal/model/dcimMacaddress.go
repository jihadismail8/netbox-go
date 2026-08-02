package model

import (
	"gorm.io/datatypes"
	"time"
)

type DcimMacaddress struct {
	ID                   uint64          `gorm:"column:id;type:int8;primary_key" json:"id"`
	Created              *time.Time      `gorm:"column:created;type:timestamptz" json:"created"`
	LastUpdated          *time.Time      `gorm:"column:last_updated;type:timestamptz" json:"lastUpdated"`
	CustomFieldData      *datatypes.JSON `gorm:"column:custom_field_data;type:jsonb;not null" json:"customFieldData"`
	Description          string          `gorm:"column:description;type:varchar(200);not null" json:"description"`
	Comments             string          `gorm:"column:comments;type:text;not null" json:"comments"`
	MacAddress           string          `gorm:"column:mac_address;type:macaddr;not null" json:"macAddress"`
	AssignedObjectID     int64           `gorm:"column:assigned_object_id;type:int8" json:"assignedObjectID"`
	AssignedObjectTypeID int             `gorm:"column:assigned_object_type_id;type:int4" json:"assignedObjectTypeID"`
}

// TableName table name
func (m *DcimMacaddress) TableName() string {
	return "dcim_macaddress"
}

// DcimMacaddressColumnNames Whitelist for custom query fields to prevent sql injection attacks
var DcimMacaddressColumnNames = map[string]bool{
	"id":                      true,
	"created":                 true,
	"last_updated":            true,
	"custom_field_data":       true,
	"description":             true,
	"comments":                true,
	"mac_address":             true,
	"assigned_object_id":      true,
	"assigned_object_type_id": true,
}
