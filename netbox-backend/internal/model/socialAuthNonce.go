package model

type SocialAuthNonce struct {
	ID        uint64 `gorm:"column:id;type:int8;primary_key" json:"id"`
	ServerURL string `gorm:"column:server_url;type:varchar(255);not null" json:"serverURL"`
	Timestamp int    `gorm:"column:timestamp;type:int4;not null" json:"timestamp"`
	Salt      string `gorm:"column:salt;type:varchar(65);not null" json:"salt"`
}

// TableName table name
func (m *SocialAuthNonce) TableName() string {
	return "social_auth_nonce"
}

// SocialAuthNonceColumnNames Whitelist for custom query fields to prevent sql injection attacks
var SocialAuthNonceColumnNames = map[string]bool{
	"id":         true,
	"server_url": true,
	"timestamp":  true,
	"salt":       true,
}
