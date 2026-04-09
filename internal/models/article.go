package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Article struct {
	ID        string     `gorm:"primaryKey;type:uuid" json:"id"`
	Title     string     `json:"title"`
	Content   string     `json:"content"`
	Category  string     `json:"category"`
	Audience []Audience `gorm:"many2many:article_audience;" json:"audience"`
	Medias   []Media    `gorm:"many2many:article_media;" json:"medias"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
}

func (a *Article) BeforeCreate(tx *gorm.DB) error {
	if a.ID == "" {
		a.ID = uuid.Must(uuid.NewV7()).String()
	}
	return nil
}
