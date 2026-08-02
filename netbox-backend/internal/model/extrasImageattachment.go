package model

import (
	"time"
)

type ExtrasImageattachment struct {
	ID           uint64     `gorm:"column:id;type:int8;primary_key" json:"id"`
	ObjectID     int64      `gorm:"column:object_id;type:int8;not null" json:"objectID"`
	Image        string     `gorm:"column:image;type:varchar(100);not null" json:"image"`
	ImageHeight  int        `gorm:"column:image_height;type:int2;not null" json:"imageHeight"`
	ImageWidth   int        `gorm:"column:image_width;type:int2;not null" json:"imageWidth"`
	Name         string     `gorm:"column:name;type:varchar(50);not null" json:"name"`
	Created      *time.Time `gorm:"column:created;type:timestamptz" json:"created"`
	ObjectTypeID int        `gorm:"column:object_type_id;type:int4;not null" json:"objectTypeID"`
	LastUpdated  *time.Time `gorm:"column:last_updated;type:timestamptz" json:"lastUpdated"`
	Description  string     `gorm:"column:description;type:varchar(200);not null" json:"description"`
}

// TableName table name
func (m *ExtrasImageattachment) TableName() string {
	return "extras_imageattachment"
}

// ExtrasImageattachmentColumnNames Whitelist for custom query fields to prevent sql injection attacks
var ExtrasImageattachmentColumnNames = map[string]bool{
	"id":             true,
	"object_id":      true,
	"image":          true,
	"image_height":   true,
	"image_width":    true,
	"name":           true,
	"created":        true,
	"object_type_id": true,
	"last_updated":   true,
	"description":    true,
}
