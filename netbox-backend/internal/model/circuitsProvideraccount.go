package model

import (
	"gorm.io/datatypes"
	"time"
)

type CircuitsProvideraccount struct {
	ID              uint64          `gorm:"column:id;type:int8;primary_key" json:"id"`
	Created         *time.Time      `gorm:"column:created;type:timestamptz" json:"created"`
	LastUpdated     *time.Time      `gorm:"column:last_updated;type:timestamptz" json:"lastUpdated"`
	CustomFieldData *datatypes.JSON `gorm:"column:custom_field_data;type:jsonb;not null" json:"customFieldData"`
	Description     string          `gorm:"column:description;type:varchar(200);not null" json:"description"`
	Comments        string          `gorm:"column:comments;type:text;not null" json:"comments"`
	Account         string          `gorm:"column:account;type:varchar(100);not null" json:"account"`
	Name            string          `gorm:"column:name;type:varchar(100);not null" json:"name"`
	ProviderID      int64           `gorm:"column:provider_id;type:int8;not null" json:"providerID"`
}

// TableName table name
func (m *CircuitsProvideraccount) TableName() string {
	return "circuits_provideraccount"
}

// CircuitsProvideraccountColumnNames Whitelist for custom query fields to prevent sql injection attacks
var CircuitsProvideraccountColumnNames = map[string]bool{
	"id":                true,
	"created":           true,
	"last_updated":      true,
	"custom_field_data": true,
	"description":       true,
	"comments":          true,
	"account":           true,
	"name":              true,
	"provider_id":       true,
}
