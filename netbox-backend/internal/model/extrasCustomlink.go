package model

import (
	"github.com/go-dev-frame/sponge/pkg/sgorm"
	"time"
)

type ExtrasCustomlink struct {
	ID          uint64      `gorm:"column:id;type:int8;primary_key" json:"id"`
	Name        string      `gorm:"column:name;type:varchar(100);not null" json:"name"`
	LinkText    string      `gorm:"column:link_text;type:text;not null" json:"linkText"`
	LinkURL     string      `gorm:"column:link_url;type:text;not null" json:"linkURL"`
	Weight      int         `gorm:"column:weight;type:int2;not null" json:"weight"`
	GroupName   string      `gorm:"column:group_name;type:varchar(50);not null" json:"groupName"`
	ButtonClass string      `gorm:"column:button_class;type:varchar(30);not null" json:"buttonClass"`
	NewWindow   *sgorm.Bool `gorm:"column:new_window;type:bool;not null" json:"newWindow"`
	Created     *time.Time  `gorm:"column:created;type:timestamptz" json:"created"`
	LastUpdated *time.Time  `gorm:"column:last_updated;type:timestamptz" json:"lastUpdated"`
	Enabled     *sgorm.Bool `gorm:"column:enabled;type:bool;not null" json:"enabled"`
}

// TableName table name
func (m *ExtrasCustomlink) TableName() string {
	return "extras_customlink"
}

// ExtrasCustomlinkColumnNames Whitelist for custom query fields to prevent sql injection attacks
var ExtrasCustomlinkColumnNames = map[string]bool{
	"id":           true,
	"name":         true,
	"link_text":    true,
	"link_url":     true,
	"weight":       true,
	"group_name":   true,
	"button_class": true,
	"new_window":   true,
	"created":      true,
	"last_updated": true,
	"enabled":      true,
}
