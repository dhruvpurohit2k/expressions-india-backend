package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID                 string     `gorm:"primaryKey" json:"id"`
	Email              string     `gorm:"uniqueIndex;not null" json:"email"`
	Password           string     `gorm:"not null" json:"-"`
	Name               string     `gorm:"type:text" json:"name"`
	Phone              string     `gorm:"type:text" json:"phone"`
	IsAdmin            bool       `gorm:"default:false" json:"isAdmin"`
	RefreshTokenHash   string     `json:"-"`
	RefreshTokenExpiry *time.Time `json:"-"`
	CreatedAt          time.Time  `json:"createdAt"`
	UpdatedAt          time.Time  `json:"updatedAt"`
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == "" {
		id, err := uuid.NewV7()
		if err != nil {
			return err
		}
		u.ID = id.String()
	}
	return nil
}
