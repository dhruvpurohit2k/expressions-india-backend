package models

import (
	"log"

	"github.com/dhruvpurohit2k/expressions-india-backend/internal/storage"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Promotion struct {
	ID  string `gorm:"primaryKey;type:uuid" json:"id"`
	URL string `json:"url"`
}

func (p *Promotion) BeforeCreate(tx *gorm.DB) error {
	if p.ID == "" {
		id, err := uuid.NewV7()
		if err != nil {
			return err
		}
		p.ID = id.String()
	}
	return nil
}

func SeedPromotions(db *gorm.DB, s3 *storage.S3) {
	imageUrls := []string{
		"./data/events/media/1.png",
		"./data/events/media/2.png",
		"./data/events/media/3.png",
	}
	for _, img := range imageUrls {

		location, key, err := s3.UploadLocal(img)
		if err != nil {
			log.Printf("failed to upload image: %v", err.Error())
			continue
		}

		if err := db.Create(&Promotion{ID: key, URL: location}).Error; err != nil {
			log.Printf("failed to Save Promotion to db: %v", err.Error())
			continue
		}

	}
}
