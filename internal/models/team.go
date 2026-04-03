package models

type Team struct {
	ID          uint   `gorm:"primaryKey"`
	Description string `gorm:"not null"`
}
