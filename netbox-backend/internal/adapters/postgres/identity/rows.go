package identity

import (
	"time"

	"gorm.io/datatypes"
)

type UserRow struct {
	ID           int64          `gorm:"primaryKey;autoIncrement"`
	Username     string         `gorm:"size:150;not null;uniqueIndex"`
	Email        string         `gorm:"size:254;not null;default:''"`
	FirstName    string         `gorm:"size:150;not null;default:''"`
	LastName     string         `gorm:"size:150;not null;default:''"`
	PasswordHash string         `gorm:"size:128;not null"`
	IsStaff      bool           `gorm:"not null;default:false"`
	IsSuperuser  bool           `gorm:"not null;default:false"`
	IsActive     bool           `gorm:"not null;default:true"`
	Permissions  datatypes.JSON `gorm:"not null"`
	Created      time.Time      `gorm:"not null"`
	Updated      time.Time      `gorm:"not null"`
}

func (UserRow) TableName() string { return "go_identity_users" }

type GroupRow struct {
	ID   int64  `gorm:"primaryKey;autoIncrement"`
	Name string `gorm:"size:150;not null;uniqueIndex"`
}

func (GroupRow) TableName() string { return "go_identity_groups" }

type GroupMembershipRow struct {
	UserID  int64     `gorm:"primaryKey;autoIncrement:false"`
	GroupID int64     `gorm:"primaryKey;autoIncrement:false"`
	User    *UserRow  `gorm:"foreignKey:UserID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"-"`
	Group   *GroupRow `gorm:"foreignKey:GroupID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"-"`
}

func (GroupMembershipRow) TableName() string { return "go_identity_group_memberships" }

// PermissionGrantRow is deliberately independent of the frozen Django
// permission tables. App/action/model derive a NetBox-style codename and the
// nullable object ID represents an object-scoped grant.
type PermissionGrantRow struct {
	ID       int64  `gorm:"primaryKey;autoIncrement"`
	Name     string `gorm:"size:150;not null"`
	AppLabel string `gorm:"size:100;not null;index"`
	Action   string `gorm:"size:16;not null;index"`
	Model    string `gorm:"size:100;not null;index"`
	ObjectID *int64 `gorm:"index"`
}

func (PermissionGrantRow) TableName() string { return "go_identity_permission_grants" }

type UserPermissionGrantRow struct {
	UserID       int64               `gorm:"primaryKey;autoIncrement:false"`
	PermissionID int64               `gorm:"primaryKey;autoIncrement:false"`
	User         *UserRow            `gorm:"foreignKey:UserID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"-"`
	Permission   *PermissionGrantRow `gorm:"foreignKey:PermissionID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"-"`
}

func (UserPermissionGrantRow) TableName() string { return "go_identity_user_permission_grants" }

type GroupPermissionGrantRow struct {
	GroupID      int64               `gorm:"primaryKey;autoIncrement:false"`
	PermissionID int64               `gorm:"primaryKey;autoIncrement:false"`
	Group        *GroupRow           `gorm:"foreignKey:GroupID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"-"`
	Permission   *PermissionGrantRow `gorm:"foreignKey:PermissionID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"-"`
}

func (GroupPermissionGrantRow) TableName() string {
	return "go_identity_group_permission_grants"
}

type TokenRow struct {
	ID           int64          `gorm:"primaryKey;autoIncrement"`
	UserID       int64          `gorm:"not null;index"`
	Display      string         `gorm:"size:32;not null"`
	SecretHash   []byte         `gorm:"size:32;not null;uniqueIndex"`
	Description  string         `gorm:"size:200;not null;default:''"`
	WriteEnabled bool           `gorm:"not null;default:false"`
	AllowedIPs   datatypes.JSON `gorm:"not null"`
	Created      time.Time      `gorm:"not null"`
	Expires      *time.Time
	LastUsed     *time.Time
	RevokedAt    *time.Time
	User         *UserRow `gorm:"foreignKey:UserID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT" json:"-"`
}

func (TokenRow) TableName() string { return "go_identity_tokens" }

type SessionRow struct {
	SecretHash []byte    `gorm:"size:32;primaryKey"`
	CSRFHash   []byte    `gorm:"size:32;not null"`
	UserID     int64     `gorm:"not null;index"`
	Created    time.Time `gorm:"not null"`
	Expires    time.Time `gorm:"not null;index"`
	LastSeen   time.Time `gorm:"not null"`
	User       *UserRow  `gorm:"foreignKey:UserID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT" json:"-"`
}

func (SessionRow) TableName() string { return "go_identity_sessions" }

// Models returns the Go-owned identity schema in dependency order.
func Models() []any {
	return []any{
		&UserRow{},
		&GroupRow{},
		&PermissionGrantRow{},
		&GroupMembershipRow{},
		&UserPermissionGrantRow{},
		&GroupPermissionGrantRow{},
		&TokenRow{},
		&SessionRow{},
	}
}
