package model

import (
	"github.com/go-dev-frame/sponge/pkg/sgorm"
	"gorm.io/datatypes"
	"time"
)

type IpamIprange struct {
	Created         *time.Time      `gorm:"column:created;type:timestamptz" json:"created"`
	LastUpdated     *time.Time      `gorm:"column:last_updated;type:timestamptz" json:"lastUpdated"`
	CustomFieldData *datatypes.JSON `gorm:"column:custom_field_data;type:jsonb;not null" json:"customFieldData"`
	ID              uint64          `gorm:"column:id;type:int8;primary_key" json:"id"`
	StartAddress    string          `gorm:"column:start_address;type:inet;not null" json:"startAddress"`
	EndAddress      string          `gorm:"column:end_address;type:inet;not null" json:"endAddress"`
	Size            int             `gorm:"column:size;type:int4;not null" json:"size"`
	Status          string          `gorm:"column:status;type:varchar(50);not null" json:"status"`
	Description     string          `gorm:"column:description;type:varchar(200);not null" json:"description"`
	RoleID          int64           `gorm:"column:role_id;type:int8" json:"roleID"`
	TenantID        int64           `gorm:"column:tenant_id;type:int8" json:"tenantID"`
	VrfID           int64           `gorm:"column:vrf_id;type:int8" json:"vrfID"`
	Comments        string          `gorm:"column:comments;type:text;not null" json:"comments"`
	MarkUtilized    *sgorm.Bool     `gorm:"column:mark_utilized;type:bool;not null" json:"markUtilized"`
	MarkPopulated   *sgorm.Bool     `gorm:"column:mark_populated;type:bool;not null" json:"markPopulated"`
}

// TableName table name
func (m *IpamIprange) TableName() string {
	return "ipam_iprange"
}

// IpamIprangeColumnNames Whitelist for custom query fields to prevent sql injection attacks
var IpamIprangeColumnNames = map[string]bool{
	"created":           true,
	"last_updated":      true,
	"custom_field_data": true,
	"id":                true,
	"start_address":     true,
	"end_address":       true,
	"size":              true,
	"status":            true,
	"description":       true,
	"role_id":           true,
	"tenant_id":         true,
	"vrf_id":            true,
	"comments":          true,
	"mark_utilized":     true,
	"mark_populated":    true,
}
