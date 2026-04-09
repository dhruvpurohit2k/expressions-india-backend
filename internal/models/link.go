package models

import (
	"time"

	"gorm.io/gorm"
)

type Link struct {
	ID        string         `gorm:"primaryKey:EventID;type:uuid" json:"id"`
	URL       string         `json:"url"`
	CreatedAt time.Time      `json:"-"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
