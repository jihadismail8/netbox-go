package model

import (
	"time"
)

type IpamFhrpgroupassignment struct {
	Created         *time.Time `gorm:"column:created;type:timestamptz" json:"created"`
	LastUpdated     *time.Time `gorm:"column:last_updated;type:timestamptz" json:"lastUpdated"`
	ID              uint64     `gorm:"column:id;type:int8;primary_key" json:"id"`
	InterfaceID     int64      `gorm:"column:interface_id;type:int8;not null" json:"interfaceID"`
	Priority        int        `gorm:"column:priority;type:int2;not null" json:"priority"`
	GroupID         int64      `gorm:"column:group_id;type:int8;not null" json:"groupID"`
	InterfaceTypeID int        `gorm:"column:interface_type_id;type:int4;not null" json:"interfaceTypeID"`
}

// TableName table name
func (m *IpamFhrpgroupassignment) TableName() string {
	return "ipam_fhrpgroupassignment"
}

// IpamFhrpgroupassignmentColumnNames Whitelist for custom query fields to prevent sql injection attacks
var IpamFhrpgroupassignmentColumnNames = map[string]bool{
	"created":           true,
	"last_updated":      true,
	"id":                true,
	"interface_id":      true,
	"priority":          true,
	"group_id":          true,
	"interface_type_id": true,
}
