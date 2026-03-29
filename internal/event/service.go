package event

import (
	"github.com/dhruvpurohit2k/expressions-india-backend/internal/models"
	"github.com/dhruvpurohit2k/expressions-india-backend/internal/storage"
	"gorm.io/gorm"
)

type Service struct {
	db *gorm.DB
	s3 *storage.S3
}

func NewService(db *gorm.DB, s3 *storage.S3) *Service {

	return &Service{
		db: db,
		s3: s3,
	}
}

func (s *Service) GetAllEvents() ([]models.Event, error) {
	var events []models.Event
	err := s.db.Preload("Medias").Find(&events).Error

	return events, err
}
