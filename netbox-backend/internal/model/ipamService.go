package model

import (
	"gorm.io/datatypes"
	"time"
)

type IpamService struct {
	Created            *time.Time      `gorm:"column:created;type:timestamptz" json:"created"`
	LastUpdated        *time.Time      `gorm:"column:last_updated;type:timestamptz" json:"lastUpdated"`
	CustomFieldData    *datatypes.JSON `gorm:"column:custom_field_data;type:jsonb;not null" json:"customFieldData"`
	ID                 uint64          `gorm:"column:id;type:int8;primary_key" json:"id"`
	Name               string          `gorm:"column:name;type:varchar(100);not null" json:"name"`
	Protocol           string          `gorm:"column:protocol;type:varchar(50);not null" json:"protocol"`
	Ports              string          `gorm:"column:ports;type:_int4;not null" json:"ports"`
	Description        string          `gorm:"column:description;type:varchar(200);not null" json:"description"`
	Comments           string          `gorm:"column:comments;type:text;not null" json:"comments"`
	ParentObjectID     int64           `gorm:"column:parent_object_id;type:int8;not null" json:"parentObjectID"`
	ParentObjectTypeID int             `gorm:"column:parent_object_type_id;type:int4;not null" json:"parentObjectTypeID"`
}

// TableName table name
func (m *IpamService) TableName() string {
	return "ipam_service"
}

// IpamServiceColumnNames Whitelist for custom query fields to prevent sql injection attacks
var IpamServiceColumnNames = map[string]bool{
	"created":               true,
	"last_updated":          true,
	"custom_field_data":     true,
	"id":                    true,
	"name":                  true,
	"protocol":              true,
	"ports":                 true,
	"description":           true,
	"comments":              true,
	"parent_object_id":      true,
	"parent_object_type_id": true,
}
