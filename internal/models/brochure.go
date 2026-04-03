package models

type Brochure struct {
	ID          uint   `gorm:"primaryKey"`
	Description string `gorm:"not null"`
	PDF         Media  `gorm:"not null"`
	Thumbnail   Media  `gorm:"not null"`
}
