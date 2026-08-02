package model

type ExtrasExporttemplateObjectTypes struct {
	ID               uint64 `gorm:"column:id;type:int8;primary_key" json:"id"`
	ExporttemplateID int64  `gorm:"column:exporttemplate_id;type:int8;not null" json:"exporttemplateID"`
	ContenttypeID    int    `gorm:"column:contenttype_id;type:int4;not null" json:"contenttypeID"`
}

// ExtrasExporttemplateObjectTypesColumnNames Whitelist for custom query fields to prevent sql injection attacks
var ExtrasExporttemplateObjectTypesColumnNames = map[string]bool{
	"id":                true,
	"exporttemplate_id": true,
	"contenttype_id":    true,
}
