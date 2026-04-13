package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func (t *Team) BeforeCreate(tx *gorm.DB) error {
	if t.ID == "" {
		t.ID = uuid.Must(uuid.NewV7()).String()
	}
	return nil
}

type Team struct {
	ID          string    `gorm:"primaryKey;type:uuid" json:"id"`
	Description string    `gorm:"type:text" json:"description"`
	Members     []Member  `gorm:"many2many:team_member" json:"members"`
	CreatedAt   time.Time `gorm:"not null" json:"createdAt"`
	UpdatedAt   time.Time `gorm:"not null" json:"updatedAt"`
}

type Member struct {
	ID       string   `gorm:"primaryKey;type:uuid" json:"id"`
	Position string   `json:"position"`
	Holders  []string `gorm:"serializer:json" json:"holders"`
}

func (m *Member) BeforeCreate(tx *gorm.DB) error {
	if m.ID == "" {
		m.ID = uuid.Must(uuid.NewV7()).String()
	}
	return nil
}
