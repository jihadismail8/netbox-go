package model

import (
	"time"
)

type DjangoMigrations struct {
	ID      uint64     `gorm:"column:id;type:int8;primary_key" json:"id"`
	App     string     `gorm:"column:app;type:varchar(255);not null" json:"app"`
	Name    string     `gorm:"column:name;type:varchar(255);not null" json:"name"`
	Applied *time.Time `gorm:"column:applied;type:timestamptz;not null" json:"applied"`
}

// DjangoMigrationsColumnNames Whitelist for custom query fields to prevent sql injection attacks
var DjangoMigrationsColumnNames = map[string]bool{
	"id":      true,
	"app":     true,
	"name":    true,
	"applied": true,
}
