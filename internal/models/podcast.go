package models

import "gorm.io/datatypes"

type Podcast struct {
	ID          string         `gorm:"primaryKey;type:uuid" json:"id"`
	Title       string         `gorm:"not null" json:"title"`
	Link        string         `gorm:"not null" json:"link"`
	Description *string        `json:"description"`
	Tags        datatypes.JSON `json:"tags"`
	Audiences   []Audience     `gorm:"many2many:podcast_audience;" json:"audience"`
}
