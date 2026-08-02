package model

import (
	"gorm.io/datatypes"
	"time"
)

type DcimModuletypeprofile struct {
	ID              uint64          `gorm:"column:id;type:int8;primary_key" json:"id"`
	Created         *time.Time      `gorm:"column:created;type:timestamptz" json:"created"`
	LastUpdated     *time.Time      `gorm:"column:last_updated;type:timestamptz" json:"lastUpdated"`
	CustomFieldData *datatypes.JSON `gorm:"column:custom_field_data;type:jsonb;not null" json:"customFieldData"`
	Description     string          `gorm:"column:description;type:varchar(200);not null" json:"description"`
	Comments        string          `gorm:"column:comments;type:text;not null" json:"comments"`
	Name            string          `gorm:"column:name;type:varchar(100);not null" json:"name"`
	Schema          *datatypes.JSON `gorm:"column:schema;type:jsonb" json:"schema"`
}

// TableName table name
func (m *DcimModuletypeprofile) TableName() string {
	return "dcim_moduletypeprofile"
}

// DcimModuletypeprofileColumnNames Whitelist for custom query fields to prevent sql injection attacks
var DcimModuletypeprofileColumnNames = map[string]bool{
	"id":                true,
	"created":           true,
	"last_updated":      true,
	"custom_field_data": true,
	"description":       true,
	"comments":          true,
	"name":              true,
	"schema":            true,
}
