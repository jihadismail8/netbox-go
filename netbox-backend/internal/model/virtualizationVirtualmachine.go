package model

import (
	"github.com/shopspring/decimal"
	"gorm.io/datatypes"
	"time"
)

type VirtualizationVirtualmachine struct {
	Created          *time.Time       `gorm:"column:created;type:timestamptz" json:"created"`
	LastUpdated      *time.Time       `gorm:"column:last_updated;type:timestamptz" json:"lastUpdated"`
	CustomFieldData  *datatypes.JSON  `gorm:"column:custom_field_data;type:jsonb;not null" json:"customFieldData"`
	ID               uint64           `gorm:"column:id;type:int8;primary_key" json:"id"`
	LocalContextData *datatypes.JSON  `gorm:"column:local_context_data;type:jsonb" json:"localContextData"`
	Name             string           `gorm:"column:name;type:varchar(64);not null" json:"name"`
	Status           string           `gorm:"column:status;type:varchar(50);not null" json:"status"`
	Vcpus            *decimal.Decimal `gorm:"column:vcpus;type:numeric" json:"vcpus"`
	Memory           int              `gorm:"column:memory;type:int4" json:"memory"`
	Disk             int              `gorm:"column:disk;type:int4" json:"disk"`
	Comments         string           `gorm:"column:comments;type:text;not null" json:"comments"`
	ClusterID        int64            `gorm:"column:cluster_id;type:int8" json:"clusterID"`
	PlatformID       int64            `gorm:"column:platform_id;type:int8" json:"platformID"`
	PrimaryIp4ID     int64            `gorm:"column:primary_ip4_id;type:int8" json:"primaryIp4ID"`
	PrimaryIp6ID     int64            `gorm:"column:primary_ip6_id;type:int8" json:"primaryIp6ID"`
	RoleID           int64            `gorm:"column:role_id;type:int8" json:"roleID"`
	TenantID         int64            `gorm:"column:tenant_id;type:int8" json:"tenantID"`
	SiteID           int64            `gorm:"column:site_id;type:int8" json:"siteID"`
	DeviceID         int64            `gorm:"column:device_id;type:int8" json:"deviceID"`
	Description      string           `gorm:"column:description;type:varchar(200);not null" json:"description"`
	InterfaceCount   int64            `gorm:"column:interface_count;type:int8;not null" json:"interfaceCount"`
	ConfigTemplateID int64            `gorm:"column:config_template_id;type:int8" json:"configTemplateID"`
	VirtualDiskCount int64            `gorm:"column:virtual_disk_count;type:int8;not null" json:"virtualDiskCount"`
	Serial           string           `gorm:"column:serial;type:varchar(50);not null" json:"serial"`
}

// TableName table name
func (m *VirtualizationVirtualmachine) TableName() string {
	return "virtualization_virtualmachine"
}

// VirtualizationVirtualmachineColumnNames Whitelist for custom query fields to prevent sql injection attacks
var VirtualizationVirtualmachineColumnNames = map[string]bool{
	"created":            true,
	"last_updated":       true,
	"custom_field_data":  true,
	"id":                 true,
	"local_context_data": true,
	"name":               true,
	"status":             true,
	"vcpus":              true,
	"memory":             true,
	"disk":               true,
	"comments":           true,
	"cluster_id":         true,
	"platform_id":        true,
	"primary_ip4_id":     true,
	"primary_ip6_id":     true,
	"role_id":            true,
	"tenant_id":          true,
	"site_id":            true,
	"device_id":          true,
	"description":        true,
	"interface_count":    true,
	"config_template_id": true,
	"virtual_disk_count": true,
	"serial":             true,
}
