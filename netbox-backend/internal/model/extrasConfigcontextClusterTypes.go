package model

type ExtrasConfigcontextClusterTypes struct {
	ID              uint64 `gorm:"column:id;type:int8;primary_key" json:"id"`
	ConfigcontextID int64  `gorm:"column:configcontext_id;type:int8;not null" json:"configcontextID"`
	ClustertypeID   int64  `gorm:"column:clustertype_id;type:int8;not null" json:"clustertypeID"`
}

// ExtrasConfigcontextClusterTypesColumnNames Whitelist for custom query fields to prevent sql injection attacks
var ExtrasConfigcontextClusterTypesColumnNames = map[string]bool{
	"id":               true,
	"configcontext_id": true,
	"clustertype_id":   true,
}
