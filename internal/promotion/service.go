package promotion

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
func (s *Service) Get() ([]models.Promotion, error) {
	var promotions []models.Promotion
	err := s.db.Find(&promotions).Error
	if err != nil {
		return nil, err
	}
	return promotions, nil
}
