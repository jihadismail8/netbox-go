package model

type CoreAutosyncrecord struct {
	ID           uint64 `gorm:"column:id;type:int8;primary_key" json:"id"`
	ObjectID     int64  `gorm:"column:object_id;type:int8;not null" json:"objectID"`
	DatafileID   int64  `gorm:"column:datafile_id;type:int8;not null" json:"datafileID"`
	ObjectTypeID int    `gorm:"column:object_type_id;type:int4;not null" json:"objectTypeID"`
}

// TableName table name
func (m *CoreAutosyncrecord) TableName() string {
	return "core_autosyncrecord"
}

// CoreAutosyncrecordColumnNames Whitelist for custom query fields to prevent sql injection attacks
var CoreAutosyncrecordColumnNames = map[string]bool{
	"id":             true,
	"object_id":      true,
	"datafile_id":    true,
	"object_type_id": true,
}
