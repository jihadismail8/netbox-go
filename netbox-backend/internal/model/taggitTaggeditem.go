package model

type TaggitTaggeditem struct {
	ID            uint64 `gorm:"column:id;type:int4;primary_key" json:"id"`
	ObjectID      int    `gorm:"column:object_id;type:int4;not null" json:"objectID"`
	ContentTypeID int    `gorm:"column:content_type_id;type:int4;not null" json:"contentTypeID"`
	TagID         int    `gorm:"column:tag_id;type:int4;not null" json:"tagID"`
}

// TableName table name
func (m *TaggitTaggeditem) TableName() string {
	return "taggit_taggeditem"
}

// TaggitTaggeditemColumnNames Whitelist for custom query fields to prevent sql injection attacks
var TaggitTaggeditemColumnNames = map[string]bool{
	"id":              true,
	"object_id":       true,
	"content_type_id": true,
	"tag_id":          true,
}
