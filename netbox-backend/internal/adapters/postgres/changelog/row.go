package changelog

import (
	"time"

	"gorm.io/datatypes"

	identityrow "netbox-go/internal/adapters/postgres/identity"
)

// ChangeRow is the private, append-only PostgreSQL representation of an
// object change. Before/after remain JSON because they are immutable
// historical snapshots across the closed first-profile object set.
type ChangeRow struct {
	ID         int64  `gorm:"primaryKey;autoIncrement"`
	ActorID    int64  `gorm:"not null;index"`
	Action     string `gorm:"size:16;not null;index"`
	Kind       string `gorm:"size:40;not null;index:idx_go_change_object,priority:1"`
	ObjectID   int64  `gorm:"not null;index:idx_go_change_object,priority:2"`
	BeforeData datatypes.JSON
	AfterData  datatypes.JSON
	OccurredAt time.Time            `gorm:"not null;index"`
	Actor      *identityrow.UserRow `gorm:"foreignKey:ActorID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT" json:"-"`
}

func (ChangeRow) TableName() string { return "go_object_changes" }
