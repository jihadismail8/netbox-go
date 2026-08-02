package model

import (
	"github.com/go-dev-frame/sponge/pkg/sgorm"
	"gorm.io/datatypes"
	"time"
)

type DcimPowerfeed struct {
	Created         *time.Time      `gorm:"column:created;type:timestamptz" json:"created"`
	LastUpdated     *time.Time      `gorm:"column:last_updated;type:timestamptz" json:"lastUpdated"`
	CustomFieldData *datatypes.JSON `gorm:"column:custom_field_data;type:jsonb;not null" json:"customFieldData"`
	ID              uint64          `gorm:"column:id;type:int8;primary_key" json:"id"`
	MarkConnected   *sgorm.Bool     `gorm:"column:mark_connected;type:bool;not null" json:"markConnected"`
	Name            string          `gorm:"column:name;type:varchar(100);not null" json:"name"`
	Status          string          `gorm:"column:status;type:varchar(50);not null" json:"status"`
	Type            string          `gorm:"column:type;type:varchar(50);not null" json:"type"`
	Supply          string          `gorm:"column:supply;type:varchar(50);not null" json:"supply"`
	Phase           string          `gorm:"column:phase;type:varchar(50);not null" json:"phase"`
	Voltage         int             `gorm:"column:voltage;type:int2;not null" json:"voltage"`
	Amperage        int             `gorm:"column:amperage;type:int2;not null" json:"amperage"`
	MaxUtilization  int             `gorm:"column:max_utilization;type:int2;not null" json:"maxUtilization"`
	AvailablePower  int             `gorm:"column:available_power;type:int4;not null" json:"availablePower"`
	Comments        string          `gorm:"column:comments;type:text;not null" json:"comments"`
	XPathID         int64           `gorm:"column:_path_id;type:int8" json:"XPathID"`
	CableID         int64           `gorm:"column:cable_id;type:int8" json:"cableID"`
	PowerPanelID    int64           `gorm:"column:power_panel_id;type:int8;not null" json:"powerPanelID"`
	RackID          int64           `gorm:"column:rack_id;type:int8" json:"rackID"`
	CableEnd        string          `gorm:"column:cable_end;type:varchar(1)" json:"cableEnd"`
	Description     string          `gorm:"column:description;type:varchar(200);not null" json:"description"`
	TenantID        int64           `gorm:"column:tenant_id;type:int8" json:"tenantID"`
}

// TableName table name
func (m *DcimPowerfeed) TableName() string {
	return "dcim_powerfeed"
}

// DcimPowerfeedColumnNames Whitelist for custom query fields to prevent sql injection attacks
var DcimPowerfeedColumnNames = map[string]bool{
	"created":           true,
	"last_updated":      true,
	"custom_field_data": true,
	"id":                true,
	"mark_connected":    true,
	"name":              true,
	"status":            true,
	"type":              true,
	"supply":            true,
	"phase":             true,
	"voltage":           true,
	"amperage":          true,
	"max_utilization":   true,
	"available_power":   true,
	"comments":          true,
	"_path_id":          true,
	"cable_id":          true,
	"power_panel_id":    true,
	"rack_id":           true,
	"cable_end":         true,
	"description":       true,
	"tenant_id":         true,
}
