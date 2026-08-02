package model

import (
	"gorm.io/datatypes"
	"time"
)

type IpamVlantranslationrule struct {
	ID              uint64          `gorm:"column:id;type:int8;primary_key" json:"id"`
	Created         *time.Time      `gorm:"column:created;type:timestamptz" json:"created"`
	LastUpdated     *time.Time      `gorm:"column:last_updated;type:timestamptz" json:"lastUpdated"`
	CustomFieldData *datatypes.JSON `gorm:"column:custom_field_data;type:jsonb;not null" json:"customFieldData"`
	LocalVid        int             `gorm:"column:local_vid;type:int2;not null" json:"localVid"`
	RemoteVid       int             `gorm:"column:remote_vid;type:int2;not null" json:"remoteVid"`
	PolicyID        int64           `gorm:"column:policy_id;type:int8;not null" json:"policyID"`
	Description     string          `gorm:"column:description;type:varchar(200);not null" json:"description"`
}

// TableName table name
func (m *IpamVlantranslationrule) TableName() string {
	return "ipam_vlantranslationrule"
}

// IpamVlantranslationruleColumnNames Whitelist for custom query fields to prevent sql injection attacks
var IpamVlantranslationruleColumnNames = map[string]bool{
	"id":                true,
	"created":           true,
	"last_updated":      true,
	"custom_field_data": true,
	"local_vid":         true,
	"remote_vid":        true,
	"policy_id":         true,
	"description":       true,
}
