package podcast

import (
	"github.com/dhruvpurohit2k/expressions-india-backend/internal/dto"
	"github.com/dhruvpurohit2k/expressions-india-backend/internal/models"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Service struct {
	db *gorm.DB
}

func (s *Service) CreatePodcast(o *dto.PodcastCreateDTO) any {
	podcast := models.Podcast{
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

func (s *Service) GetPodcasts() ([]dto.PodcastDTO, error) {
	var podcasts []models.Podcast
	if err := s.db.Find(&podcasts).Error; err != nil {
		return nil, err
	}
	// var podcastDTOs []dto.PodcastDTO
	podcastDTOs := make([]dto.PodcastDTO, 0, len(podcasts))
	for _, podcast := range podcasts {
		data := dto.PodcastDTO{
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
		podcastDTOs = append(podcastDTOs, data)
	}
	return podcastDTOs, nil
}

func (s *Service) GetPodcastById(id string) (*dto.PodcastDTO, error) {
	var podcast models.Podcast
	if err := s.db.Where("id = ?", id).First(&podcast).Error; err != nil {
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
