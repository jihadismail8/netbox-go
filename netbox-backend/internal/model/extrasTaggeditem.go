package model

type ExtrasTaggeditem struct {
	ObjectID      int    `gorm:"column:object_id;type:int4;not null" json:"objectID"`
	ID            uint64 `gorm:"column:id;type:int8;primary_key" json:"id"`
	ContentTypeID int    `gorm:"column:content_type_id;type:int4;not null" json:"contentTypeID"`
	TagID         int64  `gorm:"column:tag_id;type:int8;not null" json:"tagID"`
}

// TableName table name
func (m *ExtrasTaggeditem) TableName() string {
	return "extras_taggeditem"
}

// ExtrasTaggeditemColumnNames Whitelist for custom query fields to prevent sql injection attacks
var ExtrasTaggeditemColumnNames = map[string]bool{
	"object_id":       true,
	"id":              true,
	"content_type_id": true,
	"tag_id":          true,
}
