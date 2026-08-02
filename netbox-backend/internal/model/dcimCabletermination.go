package model

import (
	"time"
)

type DcimCabletermination struct {
	ID                uint64     `gorm:"column:id;type:int8;primary_key" json:"id"`
	CableEnd          string     `gorm:"column:cable_end;type:varchar(1);not null" json:"cableEnd"`
	TerminationID     int64      `gorm:"column:termination_id;type:int8;not null" json:"terminationID"`
	CableID           int64      `gorm:"column:cable_id;type:int8;not null" json:"cableID"`
	TerminationTypeID int        `gorm:"column:termination_type_id;type:int4;not null" json:"terminationTypeID"`
	XDeviceID         int64      `gorm:"column:_device_id;type:int8" json:"XDeviceID"`
	XRackID           int64      `gorm:"column:_rack_id;type:int8" json:"XRackID"`
	XLocationID       int64      `gorm:"column:_location_id;type:int8" json:"XLocationID"`
	XSiteID           int64      `gorm:"column:_site_id;type:int8" json:"XSiteID"`
	Created           *time.Time `gorm:"column:created;type:timestamptz" json:"created"`
	LastUpdated       *time.Time `gorm:"column:last_updated;type:timestamptz" json:"lastUpdated"`
}

// TableName table name
func (m *DcimCabletermination) TableName() string {
	return "dcim_cabletermination"
}

// DcimCableterminationColumnNames Whitelist for custom query fields to prevent sql injection attacks
var DcimCableterminationColumnNames = map[string]bool{
	"id":                  true,
	"cable_end":           true,
	"termination_id":      true,
	"cable_id":            true,
	"termination_type_id": true,
	"_device_id":          true,
	"_rack_id":            true,
	"_location_id":        true,
	"_site_id":            true,
	"created":             true,
	"last_updated":        true,
}
