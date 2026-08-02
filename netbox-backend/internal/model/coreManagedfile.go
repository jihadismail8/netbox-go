package model

import (
	"github.com/go-dev-frame/sponge/pkg/sgorm"
	"time"
)

type CoreManagedfile struct {
	ID              uint64      `gorm:"column:id;type:int8;primary_key" json:"id"`
	DataPath        string      `gorm:"column:data_path;type:varchar(1000);not null" json:"dataPath"`
	DataSynced      *time.Time  `gorm:"column:data_synced;type:timestamptz" json:"dataSynced"`
	Created         *time.Time  `gorm:"column:created;type:timestamptz;not null" json:"created"`
	LastUpdated     *time.Time  `gorm:"column:last_updated;type:timestamptz" json:"lastUpdated"`
	FileRoot        string      `gorm:"column:file_root;type:varchar(1000);not null" json:"fileRoot"`
	FilePath        string      `gorm:"column:file_path;type:varchar(100);not null" json:"filePath"`
	DataFileID      int64       `gorm:"column:data_file_id;type:int8" json:"dataFileID"`
	DataSourceID    int64       `gorm:"column:data_source_id;type:int8" json:"dataSourceID"`
	AutoSyncEnabled *sgorm.Bool `gorm:"column:auto_sync_enabled;type:bool;not null" json:"autoSyncEnabled"`
}

// TableName table name
func (m *CoreManagedfile) TableName() string {
	return "core_managedfile"
}

// CoreManagedfileColumnNames Whitelist for custom query fields to prevent sql injection attacks
var CoreManagedfileColumnNames = map[string]bool{
	"id":                true,
	"data_path":         true,
	"data_synced":       true,
	"created":           true,
	"last_updated":      true,
	"file_root":         true,
	"file_path":         true,
	"data_file_id":      true,
	"data_source_id":    true,
	"auto_sync_enabled": true,
}
