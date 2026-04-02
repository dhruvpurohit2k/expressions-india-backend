package models

import "gorm.io/gorm"

type Audience struct {
	ID   uint   `gorm:"primaryKey" json:"-"`
	Name string `gorm:"uniqueIndex" json:"name"`
}

func SeedAudience(db *gorm.DB) {
	options := []Audience{
		{Name: "all"},
		{Name: "student"},
		{Name: "teacher"},
		{Name: "head_of_department"},
		{Name: "parent"},
		{Name: "counselor"},
		{Name: "mental_health_professional"},
	}
	for _, opt := range options {
		db.FirstOrCreate(&opt, Audience{Name: opt.Name})
	}
}
