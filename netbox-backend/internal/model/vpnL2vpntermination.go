package model

import (
	"gorm.io/datatypes"
	"time"
)

type VpnL2Vpntermination struct {
	ID                   uint64          `gorm:"column:id;type:int8;primary_key" json:"id"`
	Created              *time.Time      `gorm:"column:created;type:timestamptz" json:"created"`
	LastUpdated          *time.Time      `gorm:"column:last_updated;type:timestamptz" json:"lastUpdated"`
	CustomFieldData      *datatypes.JSON `gorm:"column:custom_field_data;type:jsonb;not null" json:"customFieldData"`
	AssignedObjectID     int64           `gorm:"column:assigned_object_id;type:int8;not null" json:"assignedObjectID"`
	AssignedObjectTypeID int             `gorm:"column:assigned_object_type_id;type:int4;not null" json:"assignedObjectTypeID"`
	L2VpnID              int64           `gorm:"column:l2vpn_id;type:int8;not null" json:"l2vpnID"`
}

// TableName table name
func (m *VpnL2Vpntermination) TableName() string {
	return "vpn_l2vpntermination"
}

// VpnL2VpnterminationColumnNames Whitelist for custom query fields to prevent sql injection attacks
var VpnL2VpnterminationColumnNames = map[string]bool{
	"id":                      true,
	"created":                 true,
	"last_updated":            true,
	"custom_field_data":       true,
	"assigned_object_id":      true,
	"assigned_object_type_id": true,
	"l2vpn_id":                true,
}
