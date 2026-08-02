package model

type ExtrasConfigcontextSites struct {
	ID              uint64 `gorm:"column:id;type:int8;primary_key" json:"id"`
	ConfigcontextID int64  `gorm:"column:configcontext_id;type:int8;not null" json:"configcontextID"`
	SiteID          int64  `gorm:"column:site_id;type:int8;not null" json:"siteID"`
}

// ExtrasConfigcontextSitesColumnNames Whitelist for custom query fields to prevent sql injection attacks
var ExtrasConfigcontextSitesColumnNames = map[string]bool{
	"id":               true,
	"configcontext_id": true,
	"site_id":          true,
}
