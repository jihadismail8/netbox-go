package model

type ExtrasConfigcontextTenantGroups struct {
	ID              uint64 `gorm:"column:id;type:int8;primary_key" json:"id"`
	ConfigcontextID int64  `gorm:"column:configcontext_id;type:int8;not null" json:"configcontextID"`
	TenantgroupID   int64  `gorm:"column:tenantgroup_id;type:int8;not null" json:"tenantgroupID"`
}

// ExtrasConfigcontextTenantGroupsColumnNames Whitelist for custom query fields to prevent sql injection attacks
var ExtrasConfigcontextTenantGroupsColumnNames = map[string]bool{
	"id":               true,
	"configcontext_id": true,
	"tenantgroup_id":   true,
}
