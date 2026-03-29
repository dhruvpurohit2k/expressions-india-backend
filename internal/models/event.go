package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Event struct {
	ID          string         `gorm:"primaryKey;type:uuid" json:"id"`
	Title       string         `gorm:"type:varchar(255);not null" json:"title"`
	Description string         `gorm:"type:text" json:"description"`
	Perks       datatypes.JSON `json:"perks"`
	StartDate   time.Time      `gorm:"not null" json:"startDate"`
	EndDate     *time.Time     `json:"endDate"`
	StartTime   *string        `json:"startTime"`
	EndTime     *string        `json:"endTime"`
	Location    string         `gorm:"type:varchar(255)" json:"location"`
	IsPaid      bool           `gorm:"default:false" json:"isPaid"`
	Price       *int           `json:"price"` // Pointer allows NULL if it's free
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
	Medias      []Media        `gorm:"foreignKey:EventID" json:"medias"`
}

func (e *Event) BeforeCreate(tx *gorm.DB) error {
	newID, err := uuid.NewV7()
	if err != nil {
		return err
	}
	e.ID = newID.String()
	return nil
}
