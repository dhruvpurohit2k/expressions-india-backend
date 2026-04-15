package dto

import "time"

type LatestActivity struct {
	Type      string     `gorm:"column:type" json:"type"`
	ID        string     `gorm:"column:id" json:"id"`
	Title     string     `gorm:"column:title" json:"title"`
	StartDate *time.Time `gorm:"column:start_date" json:"start"`
	EndDate   *time.Time `gorm:"column:end_date" json:"end"`
	CreatedAt time.Time  `gorm:"column:created_at" json:"createdAt"`
}
