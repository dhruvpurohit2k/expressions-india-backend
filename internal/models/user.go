package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	ProviderPassword = "password"
	ProviderGoogle   = "google"
	ProviderApple    = "apple"
)

type User struct {
	ID                 string     `gorm:"primaryKey" json:"id"`
	Email              string     `gorm:"index;not null" json:"email"`
	EmailVerified      bool       `gorm:"default:false" json:"emailVerified"`
	Password           *string    `json:"-"` // null for OIDC-only accounts
	Name               string     `gorm:"type:text" json:"name"`
	Phone              string     `gorm:"type:text" json:"phone"`
	IsAdmin            bool       `gorm:"default:false" json:"isAdmin"`
	Provider           string     `gorm:"index:idx_user_provider_sub,unique;not null" json:"provider"`
	ProviderSub        string     `gorm:"index:idx_user_provider_sub,unique;not null" json:"-"`
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
