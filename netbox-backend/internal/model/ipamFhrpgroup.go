package model

import (
	"gorm.io/datatypes"
	"time"
)

type IpamFhrpgroup struct {
	Created         *time.Time      `gorm:"column:created;type:timestamptz" json:"created"`
	LastUpdated     *time.Time      `gorm:"column:last_updated;type:timestamptz" json:"lastUpdated"`
	CustomFieldData *datatypes.JSON `gorm:"column:custom_field_data;type:jsonb;not null" json:"customFieldData"`
	ID              uint64          `gorm:"column:id;type:int8;primary_key" json:"id"`
	GroupID         int             `gorm:"column:group_id;type:int2;not null" json:"groupID"`
	Protocol        string          `gorm:"column:protocol;type:varchar(50);not null" json:"protocol"`
	AuthType        string          `gorm:"column:auth_type;type:varchar(50)" json:"authType"`
	AuthKey         string          `gorm:"column:auth_key;type:varchar(255);not null" json:"authKey"`
	Description     string          `gorm:"column:description;type:varchar(200);not null" json:"description"`
	Name            string          `gorm:"column:name;type:varchar(100);not null" json:"name"`
	Comments        string          `gorm:"column:comments;type:text;not null" json:"comments"`
}

// TableName table name
func (m *IpamFhrpgroup) TableName() string {
	return "ipam_fhrpgroup"
}

// IpamFhrpgroupColumnNames Whitelist for custom query fields to prevent sql injection attacks
var IpamFhrpgroupColumnNames = map[string]bool{
	"created":           true,
	"last_updated":      true,
	"custom_field_data": true,
	"id":                true,
	"group_id":          true,
	"protocol":          true,
	"auth_type":         true,
	"auth_key":          true,
	"description":       true,
	"name":              true,
	"comments":          true,
}
