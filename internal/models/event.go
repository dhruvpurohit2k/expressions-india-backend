package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Event struct {
	ID                    string         `gorm:"primaryKey;type:uuid" json:"id"`
	Title                 string         `gorm:"type:varchar(255);not null" json:"title"`
	Description           string         `gorm:"type:text" json:"description"`
	Perks                 datatypes.JSON `json:"perks"`
	StartDate             time.Time      `gorm:"not null" json:"startDate"`
	EndDate               *time.Time     `json:"endDate"`
	StartTime             *string        `json:"startTime"`
	EndTime               *string        `json:"endTime"`
	Location              *string        `gorm:"type:varchar(255)" json:"location"`
	IsOnline              bool           `gorm:"default:false" json:"isOnline"`
	IsPaid                bool           `gorm:"default:false" json:"isPaid"`
	Price                 *int           `json:"price"`
	Status                *string        `gorm:"type:varchar(255)" json:"status"`
	CreatedAt             time.Time      `json:"createdAt"`
	UpdatedAt             time.Time      `json:"updatedAt"`
	DeletedAt             gorm.DeletedAt `gorm:"index" json:"-"`
	RegistrationURL       string         `gorm:"type:text;default:''" json:"registrationUrl"`
	ThumbnailID           *string        `gorm:"type:uuid" json:"thumbnailId"`
	Thumbnail             *Media         `gorm:"foreignKey:ThumbnailID" json:"thumbnail"`
	PromotionalMedia      []Media        `gorm:"many2many:event_promotional_media;" json:"promotionalMedia"`
	Medias                []Media        `gorm:"many2many:event_media;" json:"medias"`
	Documents             []Media        `gorm:"many2many:event_documents;" json:"documents"`
	PromotionalVideoLinks []Link         `gorm:"many2many:event_promotional_video_links;" json:"promotionalVideoLinks"`
	VideoLinks            []Link         `gorm:"many2many:event_video_links;" json:"videoLinks"`
	Audiences             []Audience     `gorm:"many2many:event_audience;" json:"audiences"`
}

func (e *Event) BeforeCreate(tx *gorm.DB) error {
	if e.ID == "" {
		newID, err := uuid.NewV7()
		if err != nil {
			return err
		}
		e.ID = newID.String()
	}
	return nil
}
