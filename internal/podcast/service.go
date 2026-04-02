package podcast

import (
	"github.com/dhruvpurohit2k/expressions-india-backend/internal/dto"
	"github.com/dhruvpurohit2k/expressions-india-backend/internal/models"
	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Service struct {
	db *gorm.DB
}

func (s *Service) CreatePodcast(o *dto.PodcastCreateDTO) any {
	podcast := models.Podcast{
		ID:          uuid.Must(uuid.NewV7()).String(),
		Title:       o.Title,
		Link:        o.Link,
		Description: &o.Description,
		Tags:        datatypes.JSON(o.Tags),
	}
	for _, audience := range o.Audiences {
		var a models.Audience
		if err := s.db.Where("name = ?", audience).First(&a).Error; err != nil {
			return err
		}
		podcast.Audiences = append(podcast.Audiences, a)
	}
	if err := s.db.Create(&podcast).Error; err != nil {
		return err
	}
	return nil
}

func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

func (s *Service) GetPodcasts() ([]dto.PodcastListItemDTO, error) {
	var podcasts []models.Podcast
	if err := s.db.Find(&podcasts).Error; err != nil {
		return nil, err
	}
	// var podcastDTOs []dto.PodcastDTO
	podcastDTOs := make([]dto.PodcastListItemDTO, 0, len(podcasts))
	for _, podcast := range podcasts {
		data := dto.PodcastListItemDTO{
			ID:        podcast.ID,
			Title:     podcast.Title,
			CreatedAt: podcast.CreatedAt,
		}
		podcastDTOs = append(podcastDTOs, data)
	}
	return podcastDTOs, nil
}

func (s *Service) DeletePodcast(id string) error {
	var podcast models.Podcast
	if err := s.db.First(&podcast, "id = ?", id).Error; err != nil {
		return err
	}
	if err := s.db.Model(&podcast).Association("Audiences").Clear(); err != nil {
		return err
	}
	return s.db.Delete(&podcast).Error
}

func (s *Service) GetPodcastById(id string) (*dto.PodcastDTO, error) {
	var podcast models.Podcast
	if err := s.db.Where("id = ?", id).Preload("Audiences").First(&podcast).Error; err != nil {
		return nil, err
	}
	data := &dto.PodcastDTO{
		ID:          podcast.ID,
		Title:       podcast.Title,
		Link:        podcast.Link,
		Description: podcast.Description,
		Transcript:  podcast.Transcript,
		Tags:        string(podcast.Tags),
		Audiences:   []string{},
	}
	for _, audience := range podcast.Audiences {
		data.Audiences = append(data.Audiences, audience.Name)
	}
	return data, nil
}
