package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

func (p *Podcast) BeforeCreate(tx *gorm.DB) error {
	if p.ID == "" {
		p.ID = uuid.Must(uuid.NewV7()).String()
	}
	return nil
}

type Podcast struct {
	ID          string         `gorm:"primaryKey;type:uuid" json:"id"`
	Title       string         `gorm:"not null" json:"title"`
	Link        string         `gorm:"not null" json:"link"`
	Description *string        `json:"description"`
	Tags        datatypes.JSON `json:"tags"`
	Transcript  *string        `json:"transcript"`
	Audiences   []Audience     `gorm:"many2many:podcast_audience;" json:"audience"`
	CreatedAt   time.Time      `gorm:"not null" json:"createdAt"`
	UpdatedAt   time.Time      `gorm:"not null" json:"updatedAt"`
}
