package model

type ExtrasConfigcontextPlatforms struct {
	ID              uint64 `gorm:"column:id;type:int8;primary_key" json:"id"`
	ConfigcontextID int64  `gorm:"column:configcontext_id;type:int8;not null" json:"configcontextID"`
	PlatformID      int64  `gorm:"column:platform_id;type:int8;not null" json:"platformID"`
}

// ExtrasConfigcontextPlatformsColumnNames Whitelist for custom query fields to prevent sql injection attacks
var ExtrasConfigcontextPlatformsColumnNames = map[string]bool{
	"id":               true,
	"configcontext_id": true,
	"platform_id":      true,
}
