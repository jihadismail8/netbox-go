package model

type ExtrasSavedfilterObjectTypes struct {
	ID            uint64 `gorm:"column:id;type:int8;primary_key" json:"id"`
	SavedfilterID int64  `gorm:"column:savedfilter_id;type:int8;not null" json:"savedfilterID"`
	ContenttypeID int    `gorm:"column:contenttype_id;type:int4;not null" json:"contenttypeID"`
}

// ExtrasSavedfilterObjectTypesColumnNames Whitelist for custom query fields to prevent sql injection attacks
var ExtrasSavedfilterObjectTypesColumnNames = map[string]bool{
	"id":             true,
	"savedfilter_id": true,
	"contenttype_id": true,
}
