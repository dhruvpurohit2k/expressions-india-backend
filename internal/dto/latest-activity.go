package dto

type LatestActivity struct {
	Type      string  `gorm:"column:type" json:"type"`
	ID        string  `gorm:"column:id" json:"id"`
	Title     string  `gorm:"column:title" json:"title"`
	StartDate string  `gorm:"column:start_date" json:"start"`
	EndDate   *string `gorm:"column:end_date" json:"end"`
}
