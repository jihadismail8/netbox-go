package model

import (
	"github.com/go-dev-frame/sponge/pkg/sgorm"
	"gorm.io/datatypes"
	"time"
)

type ExtrasEventrule struct {
	ID                 uint64          `gorm:"column:id;type:int8;primary_key" json:"id"`
	Created            *time.Time      `gorm:"column:created;type:timestamptz" json:"created"`
	LastUpdated        *time.Time      `gorm:"column:last_updated;type:timestamptz" json:"lastUpdated"`
	CustomFieldData    *datatypes.JSON `gorm:"column:custom_field_data;type:jsonb;not null" json:"customFieldData"`
	Name               string          `gorm:"column:name;type:varchar(150);not null" json:"name"`
	Description        string          `gorm:"column:description;type:varchar(200);not null" json:"description"`
	Enabled            *sgorm.Bool     `gorm:"column:enabled;type:bool;not null" json:"enabled"`
	Conditions         *datatypes.JSON `gorm:"column:conditions;type:jsonb" json:"conditions"`
	ActionType         string          `gorm:"column:action_type;type:varchar(30);not null" json:"actionType"`
	ActionObjectID     int64           `gorm:"column:action_object_id;type:int8" json:"actionObjectID"`
	ActionData         *datatypes.JSON `gorm:"column:action_data;type:jsonb" json:"actionData"`
	Comments           string          `gorm:"column:comments;type:text;not null" json:"comments"`
	ActionObjectTypeID int             `gorm:"column:action_object_type_id;type:int4;not null" json:"actionObjectTypeID"`
	EventTypes         string          `gorm:"column:event_types;type:_varchar;not null" json:"eventTypes"`
}

// TableName table name
func (m *ExtrasEventrule) TableName() string {
	return "extras_eventrule"
}

// ExtrasEventruleColumnNames Whitelist for custom query fields to prevent sql injection attacks
var ExtrasEventruleColumnNames = map[string]bool{
	"id":                    true,
	"created":               true,
	"last_updated":          true,
	"custom_field_data":     true,
	"name":                  true,
	"description":           true,
	"enabled":               true,
	"conditions":            true,
	"action_type":           true,
	"action_object_id":      true,
	"action_data":           true,
	"comments":              true,
	"action_object_type_id": true,
	"event_types":           true,
}
