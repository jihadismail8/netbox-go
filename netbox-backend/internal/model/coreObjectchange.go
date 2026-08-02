package model

import (
	"gorm.io/datatypes"
	"time"
)

type CoreObjectchange struct {
	ID                  uint64          `gorm:"column:id;type:int8;primary_key" json:"id"`
	Time                *time.Time      `gorm:"column:time;type:timestamptz;not null" json:"time"`
	UserName            string          `gorm:"column:user_name;type:varchar(150);not null" json:"userName"`
	RequestID           string          `gorm:"column:request_id;type:uuid;not null" json:"requestID"`
	Action              string          `gorm:"column:action;type:varchar(50);not null" json:"action"`
	ChangedObjectID     int64           `gorm:"column:changed_object_id;type:int8;not null" json:"changedObjectID"`
	RelatedObjectID     int64           `gorm:"column:related_object_id;type:int8" json:"relatedObjectID"`
	ObjectRepr          string          `gorm:"column:object_repr;type:varchar(200);not null" json:"objectRepr"`
	PrechangeData       *datatypes.JSON `gorm:"column:prechange_data;type:jsonb" json:"prechangeData"`
	PostchangeData      *datatypes.JSON `gorm:"column:postchange_data;type:jsonb" json:"postchangeData"`
	ChangedObjectTypeID int             `gorm:"column:changed_object_type_id;type:int4;not null" json:"changedObjectTypeID"`
	RelatedObjectTypeID int             `gorm:"column:related_object_type_id;type:int4" json:"relatedObjectTypeID"`
	UserID              int64           `gorm:"column:user_id;type:int8" json:"userID"`
	Message             string          `gorm:"column:message;type:varchar(200);not null" json:"message"`
}

// TableName table name
func (m *CoreObjectchange) TableName() string {
	return "core_objectchange"
}

// CoreObjectchangeColumnNames Whitelist for custom query fields to prevent sql injection attacks
var CoreObjectchangeColumnNames = map[string]bool{
	"id":                     true,
	"time":                   true,
	"user_name":              true,
	"request_id":             true,
	"action":                 true,
	"changed_object_id":      true,
	"related_object_id":      true,
	"object_repr":            true,
	"prechange_data":         true,
	"postchange_data":        true,
	"changed_object_type_id": true,
	"related_object_type_id": true,
	"user_id":                true,
	"message":                true,
}
