package model

import (
	"gorm.io/datatypes"
	"time"
)

type VpnTunneltermination struct {
	ID                uint64          `gorm:"column:id;type:int8;primary_key" json:"id"`
	Created           *time.Time      `gorm:"column:created;type:timestamptz" json:"created"`
	LastUpdated       *time.Time      `gorm:"column:last_updated;type:timestamptz" json:"lastUpdated"`
	CustomFieldData   *datatypes.JSON `gorm:"column:custom_field_data;type:jsonb;not null" json:"customFieldData"`
	Role              string          `gorm:"column:role;type:varchar(50);not null" json:"role"`
	TerminationID     int64           `gorm:"column:termination_id;type:int8" json:"terminationID"`
	TerminationTypeID int             `gorm:"column:termination_type_id;type:int4;not null" json:"terminationTypeID"`
	OutsideIpID       int64           `gorm:"column:outside_ip_id;type:int8" json:"outsideIpID"`
	TunnelID          int64           `gorm:"column:tunnel_id;type:int8;not null" json:"tunnelID"`
}

// TableName table name
func (m *VpnTunneltermination) TableName() string {
	return "vpn_tunneltermination"
}

// VpnTunnelterminationColumnNames Whitelist for custom query fields to prevent sql injection attacks
var VpnTunnelterminationColumnNames = map[string]bool{
	"id":                  true,
	"created":             true,
	"last_updated":        true,
	"custom_field_data":   true,
	"role":                true,
	"termination_id":      true,
	"termination_type_id": true,
	"outside_ip_id":       true,
	"tunnel_id":           true,
}
