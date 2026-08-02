package model

import (
	"gorm.io/datatypes"
	"time"
)

type IpamVlan struct {
	Created         *time.Time      `gorm:"column:created;type:timestamptz" json:"created"`
	LastUpdated     *time.Time      `gorm:"column:last_updated;type:timestamptz" json:"lastUpdated"`
	CustomFieldData *datatypes.JSON `gorm:"column:custom_field_data;type:jsonb;not null" json:"customFieldData"`
	ID              uint64          `gorm:"column:id;type:int8;primary_key" json:"id"`
	Vid             int             `gorm:"column:vid;type:int2;not null" json:"vid"`
	Name            string          `gorm:"column:name;type:varchar(64);not null" json:"name"`
	Status          string          `gorm:"column:status;type:varchar(50);not null" json:"status"`
	Description     string          `gorm:"column:description;type:varchar(200);not null" json:"description"`
	GroupID         int64           `gorm:"column:group_id;type:int8" json:"groupID"`
	RoleID          int64           `gorm:"column:role_id;type:int8" json:"roleID"`
	SiteID          int64           `gorm:"column:site_id;type:int8" json:"siteID"`
	TenantID        int64           `gorm:"column:tenant_id;type:int8" json:"tenantID"`
	Comments        string          `gorm:"column:comments;type:text;not null" json:"comments"`
	QinqRole        string          `gorm:"column:qinq_role;type:varchar(50)" json:"qinqRole"`
	QinqSvlanID     int64           `gorm:"column:qinq_svlan_id;type:int8" json:"qinqSvlanID"`
}

// TableName table name
func (m *IpamVlan) TableName() string {
	return "ipam_vlan"
}

// IpamVlanColumnNames Whitelist for custom query fields to prevent sql injection attacks
var IpamVlanColumnNames = map[string]bool{
	"created":           true,
	"last_updated":      true,
	"custom_field_data": true,
	"id":                true,
	"vid":               true,
	"name":              true,
	"status":            true,
	"description":       true,
	"group_id":          true,
	"role_id":           true,
	"site_id":           true,
	"tenant_id":         true,
	"comments":          true,
	"qinq_role":         true,
	"qinq_svlan_id":     true,
}
