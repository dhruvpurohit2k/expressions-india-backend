package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Media struct {
	ID        string         `gorm:"primaryKey:EventID;type:uuid" json:"id"`
	Name      string         `json:"name"`
	URL       string         `json:"url"`
	Key       string         `json:"-"`
	FileType  string         `json:"fileType"`
	Category  string         `gorm:"not null"json:"category"`
	EventID   string         `gorm:"type:uuid" json:"-"`
	CreatedAt time.Time      `json:"-"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (m *Media) BeforeCreate(tx *gorm.DB) error {
	id, err := uuid.NewV7()
	if err != nil {
		return err
	}
	m.ID = id.String()
	return nil
}
