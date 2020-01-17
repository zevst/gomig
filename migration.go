package main

import (
	"os"
	"time"
)

type Entity struct {
	Value     string     `gorm:"column:value;size:255;NOT NULL;unique_index" json:"value"`
	AppliedAt *time.Time `gorm:"column:applied_at;NOT NULL;default:CURRENT_TIMESTAMP" json:"applied_at"`
}

func (m *Entity) TableName() string {
	return os.Getenv("GOMIG_MIGRATION_TABLE_PREFIX") + "migrations"
}
