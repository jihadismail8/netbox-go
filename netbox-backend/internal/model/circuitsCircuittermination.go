package model

import (
	"github.com/go-dev-frame/sponge/pkg/sgorm"
	"gorm.io/datatypes"
	"time"
)

type CircuitsCircuittermination struct {
	Created            *time.Time      `gorm:"column:created;type:timestamptz" json:"created"`
	LastUpdated        *time.Time      `gorm:"column:last_updated;type:timestamptz" json:"lastUpdated"`
	ID                 uint64          `gorm:"column:id;type:int8;primary_key" json:"id"`
	MarkConnected      *sgorm.Bool     `gorm:"column:mark_connected;type:bool;not null" json:"markConnected"`
	TermSide           string          `gorm:"column:term_side;type:varchar(1);not null" json:"termSide"`
	PortSpeed          int             `gorm:"column:port_speed;type:int4" json:"portSpeed"`
	UpstreamSpeed      int             `gorm:"column:upstream_speed;type:int4" json:"upstreamSpeed"`
	XconnectID         string          `gorm:"column:xconnect_id;type:varchar(50);not null" json:"xconnectID"`
	PpInfo             string          `gorm:"column:pp_info;type:varchar(100);not null" json:"ppInfo"`
	Description        string          `gorm:"column:description;type:varchar(200);not null" json:"description"`
	CableID            int64           `gorm:"column:cable_id;type:int8" json:"cableID"`
	CircuitID          int64           `gorm:"column:circuit_id;type:int8;not null" json:"circuitID"`
	XProviderNetworkID int64           `gorm:"column:_provider_network_id;type:int8" json:"XProviderNetworkID"`
	CustomFieldData    *datatypes.JSON `gorm:"column:custom_field_data;type:jsonb;not null" json:"customFieldData"`
	CableEnd           string          `gorm:"column:cable_end;type:varchar(1)" json:"cableEnd"`
	TerminationID      int64           `gorm:"column:termination_id;type:int8" json:"terminationID"`
	TerminationTypeID  int             `gorm:"column:termination_type_id;type:int4" json:"terminationTypeID"`
	XLocationID        int64           `gorm:"column:_location_id;type:int8" json:"XLocationID"`
	XRegionID          int64           `gorm:"column:_region_id;type:int8" json:"XRegionID"`
	XSiteID            int64           `gorm:"column:_site_id;type:int8" json:"XSiteID"`
	XSiteGroupID       int64           `gorm:"column:_site_group_id;type:int8" json:"XSiteGroupID"`
}

// TableName table name
func (m *CircuitsCircuittermination) TableName() string {
	return "circuits_circuittermination"
}

// CircuitsCircuitterminationColumnNames Whitelist for custom query fields to prevent sql injection attacks
var CircuitsCircuitterminationColumnNames = map[string]bool{
	"created":              true,
	"last_updated":         true,
	"id":                   true,
	"mark_connected":       true,
	"term_side":            true,
	"port_speed":           true,
	"upstream_speed":       true,
	"xconnect_id":          true,
	"pp_info":              true,
	"description":          true,
	"cable_id":             true,
	"circuit_id":           true,
	"_provider_network_id": true,
	"custom_field_data":    true,
	"cable_end":            true,
	"termination_id":       true,
	"termination_type_id":  true,
	"_location_id":         true,
	"_region_id":           true,
	"_site_id":             true,
	"_site_group_id":       true,
}
