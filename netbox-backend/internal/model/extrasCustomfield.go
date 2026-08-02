package model

import (
	"github.com/go-dev-frame/sponge/pkg/sgorm"
	"github.com/shopspring/decimal"
	"gorm.io/datatypes"
	"time"
)

type ExtrasCustomfield struct {
	ID                  uint64           `gorm:"column:id;type:int8;primary_key" json:"id"`
	Type                string           `gorm:"column:type;type:varchar(50);not null" json:"type"`
	Name                string           `gorm:"column:name;type:varchar(50);not null" json:"name"`
	Label               string           `gorm:"column:label;type:varchar(50);not null" json:"label"`
	Description         string           `gorm:"column:description;type:varchar(200);not null" json:"description"`
	Required            *sgorm.Bool      `gorm:"column:required;type:bool;not null" json:"required"`
	FilterLogic         string           `gorm:"column:filter_logic;type:varchar(50);not null" json:"filterLogic"`
	Default             *datatypes.JSON  `gorm:"column:default;type:jsonb" json:"default"`
	Weight              int              `gorm:"column:weight;type:int2;not null" json:"weight"`
	ValidationMinimum   *decimal.Decimal `gorm:"column:validation_minimum;type:numeric" json:"validationMinimum"`
	ValidationMaximum   *decimal.Decimal `gorm:"column:validation_maximum;type:numeric" json:"validationMaximum"`
	ValidationRegex     string           `gorm:"column:validation_regex;type:varchar(500);not null" json:"validationRegex"`
	Created             *time.Time       `gorm:"column:created;type:timestamptz" json:"created"`
	LastUpdated         *time.Time       `gorm:"column:last_updated;type:timestamptz" json:"lastUpdated"`
	RelatedObjectTypeID int              `gorm:"column:related_object_type_id;type:int4" json:"relatedObjectTypeID"`
	GroupName           string           `gorm:"column:group_name;type:varchar(50);not null" json:"groupName"`
	SearchWeight        int              `gorm:"column:search_weight;type:int2;not null" json:"searchWeight"`
	IsCloneable         *sgorm.Bool      `gorm:"column:is_cloneable;type:bool;not null" json:"isCloneable"`
	ChoiceSetID         int64            `gorm:"column:choice_set_id;type:int8" json:"choiceSetID"`
	UiEditable          string           `gorm:"column:ui_editable;type:varchar(50);not null" json:"uiEditable"`
	UiVisible           string           `gorm:"column:ui_visible;type:varchar(50);not null" json:"uiVisible"`
	Comments            string           `gorm:"column:comments;type:text;not null" json:"comments"`
	Unique              *sgorm.Bool      `gorm:"column:unique;type:bool;not null" json:"unique"`
	RelatedObjectFilter *datatypes.JSON  `gorm:"column:related_object_filter;type:jsonb" json:"relatedObjectFilter"`
}

// TableName table name
func (m *ExtrasCustomfield) TableName() string {
	return "extras_customfield"
}

// ExtrasCustomfieldColumnNames Whitelist for custom query fields to prevent sql injection attacks
var ExtrasCustomfieldColumnNames = map[string]bool{
	"id":                     true,
	"type":                   true,
	"name":                   true,
	"label":                  true,
	"description":            true,
	"required":               true,
	"filter_logic":           true,
	"default":                true,
	"weight":                 true,
	"validation_minimum":     true,
	"validation_maximum":     true,
	"validation_regex":       true,
	"created":                true,
	"last_updated":           true,
	"related_object_type_id": true,
	"group_name":             true,
	"search_weight":          true,
	"is_cloneable":           true,
	"choice_set_id":          true,
	"ui_editable":            true,
	"ui_visible":             true,
	"comments":               true,
	"unique":                 true,
	"related_object_filter":  true,
}
