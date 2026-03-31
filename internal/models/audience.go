package models

import "gorm.io/gorm"

type Audience struct {
	ID   uint   `gorm:"primaryKey" json:"-"`
	Name string `gorm:"uniqueIndex" json:"name"`
}

func SeedAudience(db *gorm.DB) {
	options := []Audience{
		{Name: "All"},
		{Name: "Student"},
		{Name: "Teacher"},
		{Name: "Head Of Department"},
		{Name: "Parent"},
		{Name: "Counselor"},
		{Name: "Mental Health Professional"},
	}
	for _, opt := range options {
		db.FirstOrCreate(&opt, Audience{Name: opt.Name})
	}
}
