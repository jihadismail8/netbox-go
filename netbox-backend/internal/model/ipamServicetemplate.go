package model

import (
	"gorm.io/datatypes"
	"time"
)

type IpamServicetemplate struct {
	ID              uint64          `gorm:"column:id;type:int8;primary_key" json:"id"`
	Created         *time.Time      `gorm:"column:created;type:timestamptz" json:"created"`
	LastUpdated     *time.Time      `gorm:"column:last_updated;type:timestamptz" json:"lastUpdated"`
	CustomFieldData *datatypes.JSON `gorm:"column:custom_field_data;type:jsonb;not null" json:"customFieldData"`
	Protocol        string          `gorm:"column:protocol;type:varchar(50);not null" json:"protocol"`
	Ports           string          `gorm:"column:ports;type:_int4;not null" json:"ports"`
	Description     string          `gorm:"column:description;type:varchar(200);not null" json:"description"`
	Name            string          `gorm:"column:name;type:varchar(100);not null" json:"name"`
	Comments        string          `gorm:"column:comments;type:text;not null" json:"comments"`
}

// TableName table name
func (m *IpamServicetemplate) TableName() string {
	return "ipam_servicetemplate"
}

// IpamServicetemplateColumnNames Whitelist for custom query fields to prevent sql injection attacks
var IpamServicetemplateColumnNames = map[string]bool{
	"id":                true,
	"created":           true,
	"last_updated":      true,
	"custom_field_data": true,
	"protocol":          true,
	"ports":             true,
	"description":       true,
	"name":              true,
	"comments":          true,
}
