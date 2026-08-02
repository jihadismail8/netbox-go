package model

import (
	"time"
)

type ExtrasCachedvalue struct {
	ID           string     `gorm:"column:id;type:uuid;primary_key" json:"id"`
	Timestamp    *time.Time `gorm:"column:timestamp;type:timestamptz;not null" json:"timestamp"`
	ObjectID     int64      `gorm:"column:object_id;type:int8;not null" json:"objectID"`
	Field        string     `gorm:"column:field;type:varchar(200);not null" json:"field"`
	Type         string     `gorm:"column:type;type:varchar(30);not null" json:"type"`
	Value        string     `gorm:"column:value;type:text;not null" json:"value"`
	Weight       int        `gorm:"column:weight;type:int2;not null" json:"weight"`
	ObjectTypeID int        `gorm:"column:object_type_id;type:int4;not null" json:"objectTypeID"`
}

// TableName table name
func (m *ExtrasCachedvalue) TableName() string {
	return "extras_cachedvalue"
}

// ExtrasCachedvalueColumnNames Whitelist for custom query fields to prevent sql injection attacks
var ExtrasCachedvalueColumnNames = map[string]bool{
	"id":             true,
	"timestamp":      true,
	"object_id":      true,
	"field":          true,
	"type":           true,
	"value":          true,
	"weight":         true,
	"object_type_id": true,
}
