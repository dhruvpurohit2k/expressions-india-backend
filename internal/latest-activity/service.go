package latestactivity

import (
	"github.com/dhruvpurohit2k/expressions-india-backend/internal/dto"
	"gorm.io/gorm"
)

type Service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

func (s *Service) GetLatestActivity() ([]dto.LatestActivity, error) {
	query := `SELECT id, title, start_date, end_date, 'event' AS type, created_at FROM events ORDER BY created_at DESC LIMIT 5`
	var activities []dto.LatestActivity
	err := s.db.Raw(query).Scan(&activities).Error
	return activities, err
}
