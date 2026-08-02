package model

import (
	"time"
)

type DcimPoweroutlettemplate struct {
	Created      *time.Time `gorm:"column:created;type:timestamptz" json:"created"`
	LastUpdated  *time.Time `gorm:"column:last_updated;type:timestamptz" json:"lastUpdated"`
	ID           uint64     `gorm:"column:id;type:int8;primary_key" json:"id"`
	Name         string     `gorm:"column:name;type:varchar(64);not null" json:"name"`
	Label        string     `gorm:"column:label;type:varchar(64);not null" json:"label"`
	Description  string     `gorm:"column:description;type:varchar(200);not null" json:"description"`
	Type         string     `gorm:"column:type;type:varchar(50)" json:"type"`
	FeedLeg      string     `gorm:"column:feed_leg;type:varchar(50)" json:"feedLeg"`
	DeviceTypeID int64      `gorm:"column:device_type_id;type:int8" json:"deviceTypeID"`
	PowerPortID  int64      `gorm:"column:power_port_id;type:int8" json:"powerPortID"`
	ModuleTypeID int64      `gorm:"column:module_type_id;type:int8" json:"moduleTypeID"`
}

// TableName table name
func (m *DcimPoweroutlettemplate) TableName() string {
	return "dcim_poweroutlettemplate"
}

// DcimPoweroutlettemplateColumnNames Whitelist for custom query fields to prevent sql injection attacks
var DcimPoweroutlettemplateColumnNames = map[string]bool{
	"created":        true,
	"last_updated":   true,
	"id":             true,
	"name":           true,
	"label":          true,
	"description":    true,
	"type":           true,
	"feed_leg":       true,
	"device_type_id": true,
	"power_port_id":  true,
	"module_type_id": true,
}
