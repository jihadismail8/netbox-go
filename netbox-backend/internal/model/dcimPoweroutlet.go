package model

import (
	"github.com/go-dev-frame/sponge/pkg/sgorm"
	"gorm.io/datatypes"
	"time"
)

type DcimPoweroutlet struct {
	Created         *time.Time      `gorm:"column:created;type:timestamptz" json:"created"`
	LastUpdated     *time.Time      `gorm:"column:last_updated;type:timestamptz" json:"lastUpdated"`
	CustomFieldData *datatypes.JSON `gorm:"column:custom_field_data;type:jsonb;not null" json:"customFieldData"`
	ID              uint64          `gorm:"column:id;type:int8;primary_key" json:"id"`
	Name            string          `gorm:"column:name;type:varchar(64);not null" json:"name"`
	Label           string          `gorm:"column:label;type:varchar(64);not null" json:"label"`
	Description     string          `gorm:"column:description;type:varchar(200);not null" json:"description"`
	MarkConnected   *sgorm.Bool     `gorm:"column:mark_connected;type:bool;not null" json:"markConnected"`
	Type            string          `gorm:"column:type;type:varchar(50)" json:"type"`
	FeedLeg         string          `gorm:"column:feed_leg;type:varchar(50)" json:"feedLeg"`
	XPathID         int64           `gorm:"column:_path_id;type:int8" json:"XPathID"`
	CableID         int64           `gorm:"column:cable_id;type:int8" json:"cableID"`
	DeviceID        int64           `gorm:"column:device_id;type:int8;not null" json:"deviceID"`
	PowerPortID     int64           `gorm:"column:power_port_id;type:int8" json:"powerPortID"`
	ModuleID        int64           `gorm:"column:module_id;type:int8" json:"moduleID"`
	CableEnd        string          `gorm:"column:cable_end;type:varchar(1)" json:"cableEnd"`
	Color           string          `gorm:"column:color;type:varchar(6);not null" json:"color"`
	Status          string          `gorm:"column:status;type:varchar(50);not null" json:"status"`
	XLocationID     int64           `gorm:"column:_location_id;type:int8" json:"XLocationID"`
	XRackID         int64           `gorm:"column:_rack_id;type:int8" json:"XRackID"`
	XSiteID         int64           `gorm:"column:_site_id;type:int8" json:"XSiteID"`
}

// TableName table name
func (m *DcimPoweroutlet) TableName() string {
	return "dcim_poweroutlet"
}

// DcimPoweroutletColumnNames Whitelist for custom query fields to prevent sql injection attacks
var DcimPoweroutletColumnNames = map[string]bool{
	"created":           true,
	"last_updated":      true,
	"custom_field_data": true,
	"id":                true,
	"name":              true,
	"label":             true,
	"description":       true,
	"mark_connected":    true,
	"type":              true,
	"feed_leg":          true,
	"_path_id":          true,
	"cable_id":          true,
	"device_id":         true,
	"power_port_id":     true,
	"module_id":         true,
	"cable_end":         true,
	"color":             true,
	"status":            true,
	"_location_id":      true,
	"_rack_id":          true,
	"_site_id":          true,
}
