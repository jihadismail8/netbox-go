package model

import (
	"github.com/go-dev-frame/sponge/pkg/sgorm"
	"gorm.io/datatypes"
	"time"
)

type CoreDatasource struct {
	ID              uint64          `gorm:"column:id;type:int8;primary_key" json:"id"`
	Created         *time.Time      `gorm:"column:created;type:timestamptz" json:"created"`
	LastUpdated     *time.Time      `gorm:"column:last_updated;type:timestamptz" json:"lastUpdated"`
	CustomFieldData *datatypes.JSON `gorm:"column:custom_field_data;type:jsonb;not null" json:"customFieldData"`
	Description     string          `gorm:"column:description;type:varchar(200);not null" json:"description"`
	Comments        string          `gorm:"column:comments;type:text;not null" json:"comments"`
	Name            string          `gorm:"column:name;type:varchar(100);not null" json:"name"`
	Type            string          `gorm:"column:type;type:varchar(50);not null" json:"type"`
	SourceURL       string          `gorm:"column:source_url;type:varchar(200);not null" json:"sourceURL"`
	Status          string          `gorm:"column:status;type:varchar(50);not null" json:"status"`
	Enabled         *sgorm.Bool     `gorm:"column:enabled;type:bool;not null" json:"enabled"`
	IgnoreRules     string          `gorm:"column:ignore_rules;type:text;not null" json:"ignoreRules"`
	Parameters      *datatypes.JSON `gorm:"column:parameters;type:jsonb" json:"parameters"`
	LastSynced      *time.Time      `gorm:"column:last_synced;type:timestamptz" json:"lastSynced"`
	SyncInterval    int             `gorm:"column:sync_interval;type:int2" json:"syncInterval"`
}

// TableName table name
func (m *CoreDatasource) TableName() string {
	return "core_datasource"
}

// CoreDatasourceColumnNames Whitelist for custom query fields to prevent sql injection attacks
var CoreDatasourceColumnNames = map[string]bool{
	"id":                true,
	"created":           true,
	"last_updated":      true,
	"custom_field_data": true,
	"description":       true,
	"comments":          true,
	"name":              true,
	"type":              true,
	"source_url":        true,
	"status":            true,
	"enabled":           true,
	"ignore_rules":      true,
	"parameters":        true,
	"last_synced":       true,
	"sync_interval":     true,
}
