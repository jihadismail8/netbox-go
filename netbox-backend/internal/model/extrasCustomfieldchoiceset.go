package model

import (
	"github.com/go-dev-frame/sponge/pkg/sgorm"
	"time"
)

type ExtrasCustomfieldchoiceset struct {
	ID                  uint64      `gorm:"column:id;type:int8;primary_key" json:"id"`
	Created             *time.Time  `gorm:"column:created;type:timestamptz" json:"created"`
	LastUpdated         *time.Time  `gorm:"column:last_updated;type:timestamptz" json:"lastUpdated"`
	Name                string      `gorm:"column:name;type:varchar(100);not null" json:"name"`
	Description         string      `gorm:"column:description;type:varchar(200);not null" json:"description"`
	BaseChoices         string      `gorm:"column:base_choices;type:varchar(50)" json:"baseChoices"`
	ExtraChoices        string      `gorm:"column:extra_choices;type:_varchar" json:"extraChoices"`
	OrderAlphabetically *sgorm.Bool `gorm:"column:order_alphabetically;type:bool;not null" json:"orderAlphabetically"`
}

// TableName table name
func (m *ExtrasCustomfieldchoiceset) TableName() string {
	return "extras_customfieldchoiceset"
}

// ExtrasCustomfieldchoicesetColumnNames Whitelist for custom query fields to prevent sql injection attacks
var ExtrasCustomfieldchoicesetColumnNames = map[string]bool{
	"id":                   true,
	"created":              true,
	"last_updated":         true,
	"name":                 true,
	"description":          true,
	"base_choices":         true,
	"extra_choices":        true,
	"order_alphabetically": true,
}
