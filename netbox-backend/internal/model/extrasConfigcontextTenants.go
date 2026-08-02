package model

type ExtrasConfigcontextTenants struct {
	ID              uint64 `gorm:"column:id;type:int8;primary_key" json:"id"`
	ConfigcontextID int64  `gorm:"column:configcontext_id;type:int8;not null" json:"configcontextID"`
	TenantID        int64  `gorm:"column:tenant_id;type:int8;not null" json:"tenantID"`
}

// ExtrasConfigcontextTenantsColumnNames Whitelist for custom query fields to prevent sql injection attacks
var ExtrasConfigcontextTenantsColumnNames = map[string]bool{
	"id":               true,
	"configcontext_id": true,
	"tenant_id":        true,
}
