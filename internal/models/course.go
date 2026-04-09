package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func (c *Course) BeforeCreate(tx *gorm.DB) error {
	if c.ID == "" {
		c.ID = uuid.Must(uuid.NewV7()).String()
	}
	return nil
}

func (cc *CourseChapter) BeforeCreate(tx *gorm.DB) error {
	if cc.ID == "" {
		cc.ID = uuid.Must(uuid.NewV7()).String()
	}
	return nil
}

type Course struct {
	ID                  string     `gorm:"primaryKey;type:uuid" json:"id"`
	Title               string     `gorm:"type:text" json:"title"`
	Description         string     `gorm:"type:text" json:"description"`
	ThumbnailID         string     `gorm:"type:uuid" json:"thumbnail_id"`
	Thumbnail           Media      `gorm:"foreignKey:ThumbnailID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"thumbnail"`
	DownloadableContent []Media    `gorm:"many2many:downloadable_content_course;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"downloadable_content"`
	IntroductionVideoID string     `gorm:"type:uuid" json:"-"`
	IntroductionVideo   Link       `gorm:"foreignKey:IntroductionVideoID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"introductionVideo"`
	Audiences           []Audience      `gorm:"many2many:course_audience;" json:"audiences"`
	Users               []User          `gorm:"many2many:course_users;" json:"users"`
	Chapters            []CourseChapter `gorm:"foreignKey:CourseID" json:"chapters"`
	CreatedAt           time.Time       `gorm:"type:timestamp" json:"createdAt"`
	UpdatedAt           time.Time       `gorm:"type:timestamp" json:"updatedAt"`
}

type CourseChapter struct {
	ID                  string    `gorm:"primaryKey;type:uuid" json:"id"`
	CourseID            string    `gorm:"type:uuid;not null" json:"course_id"`
	Course              *Course   `gorm:"foreignKey:CourseID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
	Title               string    `gorm:"type:text" json:"title"`
	Description         string    `gorm:"type:text" json:"description"`
	DownloadableContent []Media   `gorm:"many2many:downloadable_content_course_chapter;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"downloadable_content"`
	VideoLinkID         string    `gorm:"type:uuid" json:"-"`
	VideoLink           Link      `gorm:"foreignKey:VideoLinkID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
	IsFree              bool      `gorm:"default:false;type:boolean" json:"isFree"`
	CreatedAt           time.Time `gorm:"type:timestamp" json:"createdAt"`
	UpdatedAt           time.Time `gorm:"type:timestamp" json:"updatedAt"`
}
