package model

import (
	"gorm.io/datatypes"
	"time"
)

type DcimPowerpanel struct {
	Created         *time.Time      `gorm:"column:created;type:timestamptz" json:"created"`
	LastUpdated     *time.Time      `gorm:"column:last_updated;type:timestamptz" json:"lastUpdated"`
	CustomFieldData *datatypes.JSON `gorm:"column:custom_field_data;type:jsonb;not null" json:"customFieldData"`
	ID              uint64          `gorm:"column:id;type:int8;primary_key" json:"id"`
	Name            string          `gorm:"column:name;type:varchar(100);not null" json:"name"`
	LocationID      int64           `gorm:"column:location_id;type:int8" json:"locationID"`
	SiteID          int64           `gorm:"column:site_id;type:int8;not null" json:"siteID"`
	Comments        string          `gorm:"column:comments;type:text;not null" json:"comments"`
	Description     string          `gorm:"column:description;type:varchar(200);not null" json:"description"`
}

// TableName table name
func (m *DcimPowerpanel) TableName() string {
	return "dcim_powerpanel"
}

// DcimPowerpanelColumnNames Whitelist for custom query fields to prevent sql injection attacks
var DcimPowerpanelColumnNames = map[string]bool{
	"created":           true,
	"last_updated":      true,
	"custom_field_data": true,
	"id":                true,
	"name":              true,
	"location_id":       true,
	"site_id":           true,
	"comments":          true,
	"description":       true,
}
