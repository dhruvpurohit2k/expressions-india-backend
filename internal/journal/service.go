package journal

import (
	"github.com/dhruvpurohit2k/expressions-india-backend/internal/dto"
	"github.com/dhruvpurohit2k/expressions-india-backend/internal/models"
	"gorm.io/gorm"
)

type Service struct {
	db *gorm.DB
}

func (s *Service) GetAllJournals() ([]models.Journal, error) {
	journals := []models.Journal{}
	err := s.db.Preload("Chapters").Preload("Chapters.Authors").Preload("Media").Preload("Chapters.Media").Find(&journals).Error
	return journals, err
}

func (s *Service) Get() ([]dto.JournalListItemDTO, error) {
	journals := []models.Journal{}
	if err := s.db.Find(&journals).Error; err != nil {
		return nil, err
	}
	var journaldtos []dto.JournalListItemDTO
	for _, journal := range journals {
		journaldtos = append(journaldtos, dto.JournalListItemDTO{
			ID:         journal.ID,
			Title:      journal.Title,
			Volume:     journal.Volume,
			Issue:      journal.Issue,
			StartMonth: journal.StartMonth,
			EndMonth:   journal.EndMonth,
			Year:       journal.Year,
		})
	}
	return journaldtos, nil
}

func (s *Service) GetJournalById(id string) (models.Journal, error) {
	journal := models.Journal{}
	if err := s.db.Where("id = ?", id).Preload("Chapters").Preload("Chapters.Authors").Preload("Media").Preload("Chapters.Media").First(&journal).Error; err != nil {
		return models.Journal{}, err
	}
	return journal, nil
}

func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}
