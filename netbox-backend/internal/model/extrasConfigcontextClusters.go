package model

type ExtrasConfigcontextClusters struct {
	ID              uint64 `gorm:"column:id;type:int8;primary_key" json:"id"`
	ConfigcontextID int64  `gorm:"column:configcontext_id;type:int8;not null" json:"configcontextID"`
	ClusterID       int64  `gorm:"column:cluster_id;type:int8;not null" json:"clusterID"`
}

// ExtrasConfigcontextClustersColumnNames Whitelist for custom query fields to prevent sql injection attacks
var ExtrasConfigcontextClustersColumnNames = map[string]bool{
	"id":               true,
	"configcontext_id": true,
	"cluster_id":       true,
}
