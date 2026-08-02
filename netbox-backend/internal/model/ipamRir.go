package model

import (
	"github.com/go-dev-frame/sponge/pkg/sgorm"
	"gorm.io/datatypes"
	"time"
)

type IpamRir struct {
	Created         *time.Time      `gorm:"column:created;type:timestamptz" json:"created"`
	LastUpdated     *time.Time      `gorm:"column:last_updated;type:timestamptz" json:"lastUpdated"`
	CustomFieldData *datatypes.JSON `gorm:"column:custom_field_data;type:jsonb;not null" json:"customFieldData"`
	ID              uint64          `gorm:"column:id;type:int8;primary_key" json:"id"`
	Name            string          `gorm:"column:name;type:varchar(100);not null" json:"name"`
	Slug            string          `gorm:"column:slug;type:varchar(100);not null" json:"slug"`
	IsPrivate       *sgorm.Bool     `gorm:"column:is_private;type:bool;not null" json:"isPrivate"`
	Description     string          `gorm:"column:description;type:varchar(200);not null" json:"description"`
}

// TableName table name
func (m *IpamRir) TableName() string {
	return "ipam_rir"
}

// IpamRirColumnNames Whitelist for custom query fields to prevent sql injection attacks
var IpamRirColumnNames = map[string]bool{
	"created":           true,
	"last_updated":      true,
	"custom_field_data": true,
	"id":                true,
	"name":              true,
	"slug":              true,
	"is_private":        true,
	"description":       true,
}
