package models

type Almanac struct {
	ID          uint   `gorm:"primaryKey"`
	Description string `gorm:"not null"`
	PDF         Media  `gorm:"not null"`
	Thumbnail   Media  `gorm:"not null"`
}
