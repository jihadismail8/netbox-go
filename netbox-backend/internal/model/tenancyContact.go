package model

import (
	"gorm.io/datatypes"
	"time"
)

type TenancyContact struct {
	ID              uint64          `gorm:"column:id;type:int8;primary_key" json:"id"`
	Created         *time.Time      `gorm:"column:created;type:timestamptz" json:"created"`
	LastUpdated     *time.Time      `gorm:"column:last_updated;type:timestamptz" json:"lastUpdated"`
	CustomFieldData *datatypes.JSON `gorm:"column:custom_field_data;type:jsonb;not null" json:"customFieldData"`
	Name            string          `gorm:"column:name;type:varchar(100);not null" json:"name"`
	Title           string          `gorm:"column:title;type:varchar(100);not null" json:"title"`
	Phone           string          `gorm:"column:phone;type:varchar(50);not null" json:"phone"`
	Email           string          `gorm:"column:email;type:varchar(254);not null" json:"email"`
	Address         string          `gorm:"column:address;type:varchar(200);not null" json:"address"`
	Comments        string          `gorm:"column:comments;type:text;not null" json:"comments"`
	Link            string          `gorm:"column:link;type:varchar(200);not null" json:"link"`
	Description     string          `gorm:"column:description;type:varchar(200);not null" json:"description"`
}

// TableName table name
func (m *TenancyContact) TableName() string {
	return "tenancy_contact"
}

// TenancyContactColumnNames Whitelist for custom query fields to prevent sql injection attacks
var TenancyContactColumnNames = map[string]bool{
	"id":                true,
	"created":           true,
	"last_updated":      true,
	"custom_field_data": true,
	"name":              true,
	"title":             true,
	"phone":             true,
	"email":             true,
	"address":           true,
	"comments":          true,
	"link":              true,
	"description":       true,
}
