package model

type ExtrasCustomlinkObjectTypes struct {
	ID            uint64 `gorm:"column:id;type:int8;primary_key" json:"id"`
	CustomlinkID  int64  `gorm:"column:customlink_id;type:int8;not null" json:"customlinkID"`
	ContenttypeID int    `gorm:"column:contenttype_id;type:int4;not null" json:"contenttypeID"`
}

// ExtrasCustomlinkObjectTypesColumnNames Whitelist for custom query fields to prevent sql injection attacks
var ExtrasCustomlinkObjectTypesColumnNames = map[string]bool{
	"id":             true,
	"customlink_id":  true,
	"contenttype_id": true,
}
