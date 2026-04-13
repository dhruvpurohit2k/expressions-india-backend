package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Link struct {
	ID        string         `gorm:"primaryKey;type:uuid" json:"id"`
	URL       string         `json:"url"`
	CreatedAt time.Time      `json:"-"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (l *Link) BeforeCreate(tx *gorm.DB) error {
	if l.ID == "" {
		l.ID = uuid.Must(uuid.NewV7()).String()
	}
	return nil
}
