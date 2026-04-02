package models

import "time"

type Enquiry struct {
	ID        string    `gorm:"primaryKey;type:uuid" json:"id"`
	Subject   string    `gorm:"not null" json:"subject"`
	Name      string    `gorm:"not null" json:"name"`
	Email     string    `gorm:"not null" json:"email"`
	Message   string    `gorm:"not null" json:"message"`
	Phone     string    `gorm:"not null" json:"phone"`
	CreatedAt time.Time `gorm:"not null" json:"created_at"`
}
