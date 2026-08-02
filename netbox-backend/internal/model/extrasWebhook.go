package model

import (
	"github.com/go-dev-frame/sponge/pkg/sgorm"
	"gorm.io/datatypes"
	"time"
)

type ExtrasWebhook struct {
	ID                uint64          `gorm:"column:id;type:int8;primary_key" json:"id"`
	Name              string          `gorm:"column:name;type:varchar(150);not null" json:"name"`
	PayloadURL        string          `gorm:"column:payload_url;type:varchar(500);not null" json:"payloadURL"`
	HttpMethod        string          `gorm:"column:http_method;type:varchar(30);not null" json:"httpMethod"`
	HttpContentType   string          `gorm:"column:http_content_type;type:varchar(100);not null" json:"httpContentType"`
	AdditionalHeaders string          `gorm:"column:additional_headers;type:text;not null" json:"additionalHeaders"`
	BodyTemplate      string          `gorm:"column:body_template;type:text;not null" json:"bodyTemplate"`
	Secret            string          `gorm:"column:secret;type:varchar(255);not null" json:"secret"`
	SslVerification   *sgorm.Bool     `gorm:"column:ssl_verification;type:bool;not null" json:"sslVerification"`
	CaFilePath        string          `gorm:"column:ca_file_path;type:varchar(4096)" json:"caFilePath"`
	Created           *time.Time      `gorm:"column:created;type:timestamptz" json:"created"`
	LastUpdated       *time.Time      `gorm:"column:last_updated;type:timestamptz" json:"lastUpdated"`
	CustomFieldData   *datatypes.JSON `gorm:"column:custom_field_data;type:jsonb;not null" json:"customFieldData"`
	Description       string          `gorm:"column:description;type:varchar(200);not null" json:"description"`
}

// TableName table name
func (m *ExtrasWebhook) TableName() string {
	return "extras_webhook"
}

// ExtrasWebhookColumnNames Whitelist for custom query fields to prevent sql injection attacks
var ExtrasWebhookColumnNames = map[string]bool{
	"id":                 true,
	"name":               true,
	"payload_url":        true,
	"http_method":        true,
	"http_content_type":  true,
	"additional_headers": true,
	"body_template":      true,
	"secret":             true,
	"ssl_verification":   true,
	"ca_file_path":       true,
	"created":            true,
	"last_updated":       true,
	"custom_field_data":  true,
	"description":        true,
}
