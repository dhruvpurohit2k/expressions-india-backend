package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Journal struct {
	Id          string           `gorm:"primaryKey;type:uuid" json:"id"`
	Title       string           `gorm:"not null" json:"title"`
	Description *string          `json:"description"`
	StartMonth  string           `gorm:"not null" json:"startMonth"`
	EndMonth    string           `gorm:"not null" json:"endMonth"`
	Year        int              `gorm:"not null" json:"year"`
	Volume      int              `gorm:"not null" json:"volume"`
	Issue       int              `gorm:"not null" json:"issue"`
	Media       Media            `gorm:"foreignKey:JournalId" json:"media"`
	Chapters    []JournalChapter `gorm:"foreignKey:JournalId" json:"chapters"`
}

type JournalChapter struct {
	Id          string   `gorm:"primaryKey;type:uuid" json:"id"`
	JournalId   string   `gorm:"not null" json:"journalId"`
	Title       string   `gorm:"not null" json:"title"`
	Media       Media    `gorm:"foreignKey:JournalChapterId" json:"media"`
	Description *string  `json:"description"`
	Authors     []Author `gorm:"many2many:journal_chapter_authors" json:"authors"`
}

type Author struct {
	Id   string `gorm:"primaryKey;type:uuid" json:"id"`
	Name string `gorm:"not null unique" json:"name"`
}

func (j *Journal) BeforeCreate(tx *gorm.DB) error {
	if j.Id == "" {
		j.Id = uuid.Must(uuid.NewV7()).String()
	}
	for i := range j.Chapters {
		j.Chapters[i].JournalId = j.Id
	}
	return nil
}

func (jc *JournalChapter) BeforeCreate(tx *gorm.DB) error {
	if jc.Id == "" {
		jc.Id = uuid.Must(uuid.NewV7()).String()
	}
	return nil
}

func (a *Author) BeforeCreate(tx *gorm.DB) error {
	if a.Id == "" {
		a.Id = uuid.Must(uuid.NewV7()).String()
	}
	return nil
}
