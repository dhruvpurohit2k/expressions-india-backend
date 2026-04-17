package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Brochure struct {
	ID          string    `gorm:"primaryKey;type:uuid" json:"id"`
	Title       string    `gorm:"not null" json:"title"`
	Description string    `json:"description"`
	PDFID       *string   `gorm:"type:uuid" json:"pdfId"`
	PDF         *Media    `gorm:"foreignKey:PDFID" json:"pdf"`
	ThumbnailID *string   `gorm:"type:uuid" json:"thumbnailId"`
	Thumbnail   *Media    `gorm:"foreignKey:ThumbnailID" json:"thumbnail"`
	CreatedAt   time.Time `json:"createdAt"`
}

func (b *Brochure) BeforeCreate(tx *gorm.DB) error {
	if b.ID == "" {
		b.ID = uuid.Must(uuid.NewV7()).String()
	}
	return nil
}
