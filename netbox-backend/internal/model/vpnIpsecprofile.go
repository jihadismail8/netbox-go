package model

import (
	"gorm.io/datatypes"
	"time"
)

type VpnIpsecprofile struct {
	ID              uint64          `gorm:"column:id;type:int8;primary_key" json:"id"`
	Created         *time.Time      `gorm:"column:created;type:timestamptz" json:"created"`
	LastUpdated     *time.Time      `gorm:"column:last_updated;type:timestamptz" json:"lastUpdated"`
	CustomFieldData *datatypes.JSON `gorm:"column:custom_field_data;type:jsonb;not null" json:"customFieldData"`
	Description     string          `gorm:"column:description;type:varchar(200);not null" json:"description"`
	Comments        string          `gorm:"column:comments;type:text;not null" json:"comments"`
	Name            string          `gorm:"column:name;type:varchar(100);not null" json:"name"`
	Mode            string          `gorm:"column:mode;type:varchar;not null" json:"mode"`
	IkePolicyID     int64           `gorm:"column:ike_policy_id;type:int8;not null" json:"ikePolicyID"`
	IpsecPolicyID   int64           `gorm:"column:ipsec_policy_id;type:int8;not null" json:"ipsecPolicyID"`
}

// TableName table name
func (m *VpnIpsecprofile) TableName() string {
	return "vpn_ipsecprofile"
}

// VpnIpsecprofileColumnNames Whitelist for custom query fields to prevent sql injection attacks
var VpnIpsecprofileColumnNames = map[string]bool{
	"id":                true,
	"created":           true,
	"last_updated":      true,
	"custom_field_data": true,
	"description":       true,
	"comments":          true,
	"name":              true,
	"mode":              true,
	"ike_policy_id":     true,
	"ipsec_policy_id":   true,
}
