package dto

type JournalListItemDTO struct {
	ID         string `json:"id"`
	Title      string `json:"title"`
	Volume     int    `gorm:"not null" json:"volume"`
	Issue      int    `gorm:"not null" json:"issue"`
	StartMonth string `gorm:"not null" json:"startMonth"`
	EndMonth   string `gorm:"not null" json:"endMonth"`
	Year       int    `gorm:"not null" json:"year"`
}
