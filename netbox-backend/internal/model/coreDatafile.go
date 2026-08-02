package model

import (
	"time"
)

type CoreDatafile struct {
	ID          uint64     `gorm:"column:id;type:int8;primary_key" json:"id"`
	Created     *time.Time `gorm:"column:created;type:timestamptz;not null" json:"created"`
	LastUpdated *time.Time `gorm:"column:last_updated;type:timestamptz;not null" json:"lastUpdated"`
	Path        string     `gorm:"column:path;type:varchar(1000);not null" json:"path"`
	Size        int        `gorm:"column:size;type:int4;not null" json:"size"`
	Hash        string     `gorm:"column:hash;type:varchar(64);not null" json:"hash"`
	Data        string     `gorm:"column:data;type:bytea;not null" json:"data"`
	SourceID    int64      `gorm:"column:source_id;type:int8;not null" json:"sourceID"`
}

// TableName table name
func (m *CoreDatafile) TableName() string {
	return "core_datafile"
}

// CoreDatafileColumnNames Whitelist for custom query fields to prevent sql injection attacks
var CoreDatafileColumnNames = map[string]bool{
	"id":           true,
	"created":      true,
	"last_updated": true,
	"path":         true,
	"size":         true,
	"hash":         true,
	"data":         true,
	"source_id":    true,
}
