package podcast

import (
	"github.com/dhruvpurohit2k/expressions-india-backend/internal/models"
	"gorm.io/gorm"
)

type Service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

func (s *Service) GetPodcasts() ([]models.Podcast, error) {
	var podcasts []models.Podcast
	if err := s.db.Find(&podcasts).Error; err != nil {
		return nil, err
	}
	return podcasts, nil
}
