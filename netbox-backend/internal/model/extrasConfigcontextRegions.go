package model

type ExtrasConfigcontextRegions struct {
	ID              uint64 `gorm:"column:id;type:int8;primary_key" json:"id"`
	ConfigcontextID int64  `gorm:"column:configcontext_id;type:int8;not null" json:"configcontextID"`
	RegionID        int64  `gorm:"column:region_id;type:int8;not null" json:"regionID"`
}

// ExtrasConfigcontextRegionsColumnNames Whitelist for custom query fields to prevent sql injection attacks
var ExtrasConfigcontextRegionsColumnNames = map[string]bool{
	"id":               true,
	"configcontext_id": true,
	"region_id":        true,
}
