package model

type DjangoContentType struct {
	ID       uint64 `gorm:"column:id;type:int4;primary_key" json:"id"`
	AppLabel string `gorm:"column:app_label;type:varchar(100);not null;uniqueIndex:django_content_type_app_label_model" json:"appLabel"`
	Model    string `gorm:"column:model;type:varchar(100);not null;uniqueIndex:django_content_type_app_label_model" json:"model"`
}

// TableName table name
func (m *DjangoContentType) TableName() string {
	return "django_content_type"
}

// DjangoContentTypeColumnNames Whitelist for custom query fields to prevent sql injection attacks
var DjangoContentTypeColumnNames = map[string]bool{
	"id":        true,
	"app_label": true,
	"model":     true,
}
