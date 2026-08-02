package model

import (
	"github.com/go-dev-frame/sponge/pkg/sgorm"
	"gorm.io/datatypes"
	"time"
)

type ExtrasSavedfilter struct {
	ID          uint64          `gorm:"column:id;type:int8;primary_key" json:"id"`
	Created     *time.Time      `gorm:"column:created;type:timestamptz" json:"created"`
	LastUpdated *time.Time      `gorm:"column:last_updated;type:timestamptz" json:"lastUpdated"`
	Name        string          `gorm:"column:name;type:varchar(100);not null" json:"name"`
	Slug        string          `gorm:"column:slug;type:varchar(100);not null" json:"slug"`
	Description string          `gorm:"column:description;type:varchar(200);not null" json:"description"`
	Weight      int             `gorm:"column:weight;type:int2;not null" json:"weight"`
	Enabled     *sgorm.Bool     `gorm:"column:enabled;type:bool;not null" json:"enabled"`
	Shared      *sgorm.Bool     `gorm:"column:shared;type:bool;not null" json:"shared"`
	Parameters  *datatypes.JSON `gorm:"column:parameters;type:jsonb;not null" json:"parameters"`
	UserID      int64           `gorm:"column:user_id;type:int8" json:"userID"`
}

// TableName table name
func (m *ExtrasSavedfilter) TableName() string {
	return "extras_savedfilter"
}

// ExtrasSavedfilterColumnNames Whitelist for custom query fields to prevent sql injection attacks
var ExtrasSavedfilterColumnNames = map[string]bool{
	"id":           true,
	"created":      true,
	"last_updated": true,
	"name":         true,
	"slug":         true,
	"description":  true,
	"weight":       true,
	"enabled":      true,
	"shared":       true,
	"parameters":   true,
	"user_id":      true,
}
