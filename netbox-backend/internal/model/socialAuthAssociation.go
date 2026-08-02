package model

type SocialAuthAssociation struct {
	ID        uint64 `gorm:"column:id;type:int8;primary_key" json:"id"`
	ServerURL string `gorm:"column:server_url;type:varchar(255);not null" json:"serverURL"`
	Handle    string `gorm:"column:handle;type:varchar(255);not null" json:"handle"`
	Secret    string `gorm:"column:secret;type:varchar(255);not null" json:"secret"`
	Issued    int    `gorm:"column:issued;type:int4;not null" json:"issued"`
	Lifetime  int    `gorm:"column:lifetime;type:int4;not null" json:"lifetime"`
	AssocType string `gorm:"column:assoc_type;type:varchar(64);not null" json:"assocType"`
}

// TableName table name
func (m *SocialAuthAssociation) TableName() string {
	return "social_auth_association"
}

// SocialAuthAssociationColumnNames Whitelist for custom query fields to prevent sql injection attacks
var SocialAuthAssociationColumnNames = map[string]bool{
	"id":         true,
	"server_url": true,
	"handle":     true,
	"secret":     true,
	"issued":     true,
	"lifetime":   true,
	"assoc_type": true,
}
