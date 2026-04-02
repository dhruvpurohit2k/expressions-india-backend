package models

import (
	"time"

	"gorm.io/datatypes"
)

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
