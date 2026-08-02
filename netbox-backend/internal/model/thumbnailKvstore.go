package model

type ThumbnailKvstore struct {
	Key   string `gorm:"column:key;type:varchar(200);primary_key" json:"key"`
	Value string `gorm:"column:value;type:text;not null" json:"value"`
}

// TableName table name
func (m *ThumbnailKvstore) TableName() string {
	return "thumbnail_kvstore"
}

// ThumbnailKvstoreColumnNames Whitelist for custom query fields to prevent sql injection attacks
var ThumbnailKvstoreColumnNames = map[string]bool{
	"key":   true,
	"value": true,
}
