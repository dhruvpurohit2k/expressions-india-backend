package journal

import (
	"github.com/dhruvpurohit2k/expressions-india-backend/internal/models"
	"gorm.io/gorm"
)

type Service struct {
	db *gorm.DB
}

func (s *Service) GetJournals() ([]models.Journal, error) {
	journals := []models.Journal{}
	err := s.db.Preload("Chapters").Preload("Chapters.Authors").Find(&journals).Error
	return journals, err
}

func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}
