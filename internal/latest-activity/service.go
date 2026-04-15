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
	// UNION ALL across all content types. ORDER BY and LIMIT are at the outermost
	// level so the query is valid on both SQLite and PostgreSQL.
	// events uses soft-delete, so deleted_at IS NULL is required there.
	query := `
		SELECT id, title, start_date, end_date, 'event'   AS type, created_at FROM events   WHERE deleted_at IS NULL
		UNION ALL
		SELECT id, title, NULL,       NULL,     'article' AS type, created_at FROM articles
		UNION ALL
		SELECT id, title, NULL,       NULL,     'journal' AS type, created_at FROM journals
		UNION ALL
		SELECT id, title, NULL,       NULL,     'podcast' AS type, created_at FROM podcasts
		UNION ALL
		SELECT id, title, NULL,       NULL,     'course'  AS type, created_at FROM courses
		ORDER BY created_at DESC
		LIMIT 5
	`
	var activities []dto.LatestActivity
	err := s.db.Raw(query).Scan(&activities).Error
	return activities, err
}
