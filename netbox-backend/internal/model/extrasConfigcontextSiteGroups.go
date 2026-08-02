package model

type ExtrasConfigcontextSiteGroups struct {
	ID              uint64 `gorm:"column:id;type:int8;primary_key" json:"id"`
	ConfigcontextID int64  `gorm:"column:configcontext_id;type:int8;not null" json:"configcontextID"`
	SitegroupID     int64  `gorm:"column:sitegroup_id;type:int8;not null" json:"sitegroupID"`
}

// ExtrasConfigcontextSiteGroupsColumnNames Whitelist for custom query fields to prevent sql injection attacks
var ExtrasConfigcontextSiteGroupsColumnNames = map[string]bool{
	"id":               true,
	"configcontext_id": true,
	"sitegroup_id":     true,
}
