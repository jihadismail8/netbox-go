package model

type ExtrasConfigcontextClusterGroups struct {
	ID              uint64 `gorm:"column:id;type:int8;primary_key" json:"id"`
	ConfigcontextID int64  `gorm:"column:configcontext_id;type:int8;not null" json:"configcontextID"`
	ClustergroupID  int64  `gorm:"column:clustergroup_id;type:int8;not null" json:"clustergroupID"`
}

// ExtrasConfigcontextClusterGroupsColumnNames Whitelist for custom query fields to prevent sql injection attacks
var ExtrasConfigcontextClusterGroupsColumnNames = map[string]bool{
	"id":               true,
	"configcontext_id": true,
	"clustergroup_id":  true,
}
