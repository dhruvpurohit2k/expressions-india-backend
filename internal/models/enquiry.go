package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func (e *Enquiry) BeforeCreate(tx *gorm.DB) error {
	if e.ID == "" {
		e.ID = uuid.Must(uuid.NewV7()).String()
	}
	return nil
}

type Enquiry struct {
	ID        string    `gorm:"primaryKey;type:uuid" json:"id"`
	Subject   string    `gorm:"not null" json:"subject"`
	Name      string    `gorm:"not null" json:"name"`
	Email     string    `gorm:"not null" json:"email"`
	Message   string    `gorm:"not null" json:"message"`
	Phone     string    `gorm:"not null" json:"phone"`
	CreatedAt time.Time `gorm:"not null" json:"created_at"`
}
