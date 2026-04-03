package journal

import (
	"github.com/dhruvpurohit2k/expressions-india-backend/internal/dto"
	"github.com/dhruvpurohit2k/expressions-india-backend/internal/models"
	"github.com/dhruvpurohit2k/expressions-india-backend/internal/storage"
	"gorm.io/gorm"
)

type Service struct {
	db *gorm.DB
	s3 *storage.S3
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

func (s *Service) GetJournalList(limit int, offset int) ([]dto.JournalListItemDTO, int64, error) {
	var journals []models.Journal
	var total int64

	base := s.db.Model(&models.Journal{})

	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := base.Order("year DESC, issue DESC").Limit(limit).Offset(offset).Find(&journals).Error; err != nil {
		return nil, 0, err
	}

	result := make([]dto.JournalListItemDTO, 0, len(journals))
	for _, journal := range journals {
		result = append(result, dto.JournalListItemDTO{
			ID:         journal.ID,
			Title:      journal.Title,
			Volume:     journal.Volume,
			Issue:      journal.Issue,
			StartMonth: journal.StartMonth,
			EndMonth:   journal.EndMonth,
			Year:       journal.Year,
		})
	}
	return result, total, nil
}

func (s *Service) GetJournalById(id string) (models.Journal, error) {
	journal := models.Journal{}
	if err := s.db.Where("id = ?", id).Preload("Chapters").Preload("Chapters.Authors").Preload("Media").Preload("Chapters.Media").First(&journal).Error; err != nil {
		return models.Journal{}, err
	}
	return journal, nil
}

func (s *Service) DeleteJournal(id string) error {
	var journal models.Journal
	if err := s.db.Preload("Media").Preload("Chapters").Preload("Chapters.Media").Preload("Chapters.Authors").First(&journal, "id = ?", id).Error; err != nil {
		return err
	}

	for _, chapter := range journal.Chapters {
		if err := s.db.Model(&chapter).Association("Authors").Clear(); err != nil {
			return err
		}
		if chapter.MediaId != nil {
			if err := s.db.Delete(&models.Media{}, "id = ?", *chapter.MediaId).Error; err != nil {
				return err
			}
			if err := s.s3.Delete(*chapter.MediaId); err != nil {
				return err
			}
		}
		if err := s.db.Delete(&chapter).Error; err != nil {
			return err
		}
	}

	if journal.MediaId != nil {
		if err := s.db.Delete(&models.Media{}, "id = ?", *journal.MediaId).Error; err != nil {
			return err
		}
		if err := s.s3.Delete(*journal.MediaId); err != nil {
			return err
		}
	}

	return s.db.Delete(&journal).Error
}

func NewService(db *gorm.DB, s3 *storage.S3) *Service {
	return &Service{db: db, s3: s3}
}
