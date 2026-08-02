package model

import (
	"gorm.io/datatypes"
	"time"
)

type CircuitsVirtualcircuittermination struct {
	ID               uint64          `gorm:"column:id;type:int8;primary_key" json:"id"`
	Created          *time.Time      `gorm:"column:created;type:timestamptz" json:"created"`
	LastUpdated      *time.Time      `gorm:"column:last_updated;type:timestamptz" json:"lastUpdated"`
	CustomFieldData  *datatypes.JSON `gorm:"column:custom_field_data;type:jsonb;not null" json:"customFieldData"`
	Role             string          `gorm:"column:role;type:varchar(50);not null" json:"role"`
	Description      string          `gorm:"column:description;type:varchar(200);not null" json:"description"`
	InterfaceID      int64           `gorm:"column:interface_id;type:int8;not null" json:"interfaceID"`
	VirtualCircuitID int64           `gorm:"column:virtual_circuit_id;type:int8;not null" json:"virtualCircuitID"`
}

// TableName table name
func (m *CircuitsVirtualcircuittermination) TableName() string {
	return "circuits_virtualcircuittermination"
}

// CircuitsVirtualcircuitterminationColumnNames Whitelist for custom query fields to prevent sql injection attacks
var CircuitsVirtualcircuitterminationColumnNames = map[string]bool{
	"id":                 true,
	"created":            true,
	"last_updated":       true,
	"custom_field_data":  true,
	"role":               true,
	"description":        true,
	"interface_id":       true,
	"virtual_circuit_id": true,
}
