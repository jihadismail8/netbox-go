package model

type ExtrasConfigcontextTags struct {
	ID              uint64 `gorm:"column:id;type:int8;primary_key" json:"id"`
	ConfigcontextID int64  `gorm:"column:configcontext_id;type:int8;not null" json:"configcontextID"`
	TagID           int64  `gorm:"column:tag_id;type:int8;not null" json:"tagID"`
}

// ExtrasConfigcontextTagsColumnNames Whitelist for custom query fields to prevent sql injection attacks
var ExtrasConfigcontextTagsColumnNames = map[string]bool{
	"id":               true,
	"configcontext_id": true,
	"tag_id":           true,
}
