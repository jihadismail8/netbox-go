package model

import (
	"gorm.io/datatypes"
	"time"
)

type CoreJob struct {
	ID           uint64          `gorm:"column:id;type:int8;primary_key" json:"id"`
	ObjectID     int64           `gorm:"column:object_id;type:int8" json:"objectID"`
	Name         string          `gorm:"column:name;type:varchar(200);not null" json:"name"`
	Created      *time.Time      `gorm:"column:created;type:timestamptz;not null" json:"created"`
	Scheduled    *time.Time      `gorm:"column:scheduled;type:timestamptz" json:"scheduled"`
	Interval     int             `gorm:"column:interval;type:int4" json:"interval"`
	Started      *time.Time      `gorm:"column:started;type:timestamptz" json:"started"`
	Completed    *time.Time      `gorm:"column:completed;type:timestamptz" json:"completed"`
	Status       string          `gorm:"column:status;type:varchar(30);not null" json:"status"`
	Data         *datatypes.JSON `gorm:"column:data;type:jsonb" json:"data"`
	JobID        string          `gorm:"column:job_id;type:uuid;not null" json:"jobID"`
	ObjectTypeID int             `gorm:"column:object_type_id;type:int4" json:"objectTypeID"`
	UserID       int64           `gorm:"column:user_id;type:int8" json:"userID"`
	Error        string          `gorm:"column:error;type:text;not null" json:"error"`
	LogEntries   string          `gorm:"column:log_entries;type:_jsonb;not null" json:"logEntries"`
}

// TableName table name
func (m *CoreJob) TableName() string {
	return "core_job"
}

// CoreJobColumnNames Whitelist for custom query fields to prevent sql injection attacks
var CoreJobColumnNames = map[string]bool{
	"id":             true,
	"object_id":      true,
	"name":           true,
	"created":        true,
	"scheduled":      true,
	"interval":       true,
	"started":        true,
	"completed":      true,
	"status":         true,
	"data":           true,
	"job_id":         true,
	"object_type_id": true,
	"user_id":        true,
	"error":          true,
	"log_entries":    true,
}
