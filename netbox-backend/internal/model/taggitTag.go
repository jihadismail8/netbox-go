package model

type TaggitTag struct {
	ID   uint64 `gorm:"column:id;type:int4;primary_key" json:"id"`
	Name string `gorm:"column:name;type:varchar(100);not null" json:"name"`
	Slug string `gorm:"column:slug;type:varchar(100);not null" json:"slug"`
}

// TableName table name
func (m *TaggitTag) TableName() string {
	return "taggit_tag"
}

// TaggitTagColumnNames Whitelist for custom query fields to prevent sql injection attacks
var TaggitTagColumnNames = map[string]bool{
	"id":   true,
	"name": true,
	"slug": true,
}
