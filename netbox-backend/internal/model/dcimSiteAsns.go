package model

type DcimSiteAsns struct {
	ID     uint64 `gorm:"column:id;type:int8;primary_key" json:"id"`
	SiteID int64  `gorm:"column:site_id;type:int8;not null" json:"siteID"`
	AsnID  int64  `gorm:"column:asn_id;type:int8;not null" json:"asnID"`
}

// DcimSiteAsnsColumnNames Whitelist for custom query fields to prevent sql injection attacks
var DcimSiteAsnsColumnNames = map[string]bool{
	"id":      true,
	"site_id": true,
	"asn_id":  true,
}
