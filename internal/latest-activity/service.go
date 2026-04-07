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
	query := `
	SELECT * FROM (
    SELECT * FROM ( SELECT id, title, start_date, end_date, 'event' AS type FROM events ORDER BY created_at DESC LIMIT 5 ) AS e
    UNION ALL
    SELECT * FROM ( SELECT id, title, created_at AS start_date, NULL AS end_date, 'podcast' AS type FROM podcasts ORDER BY created_at DESC LIMIT 5 ) AS p
    UNION ALL
    SELECT * FROM ( SELECT id, title, created_at AS start_date, NULL AS end_date, 'article' AS type FROM articles ORDER BY created_at DESC LIMIT 5 ) AS a
    UNION ALL
    SELECT * FROM ( SELECT id, title, created_at AS start_date, NULL AS end_date, 'journal' AS type FROM journals ORDER BY created_at DESC LIMIT 5 ) AS j
) AS latest_activity
ORDER BY start_date DESC
LIMIT 5;
	`
	var activities []dto.LatestActivity
	err := s.db.Raw(query).Scan(&activities).Error
	return activities, err
}
