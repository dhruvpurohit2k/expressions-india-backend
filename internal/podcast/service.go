package podcast

import (
	"github.com/dhruvpurohit2k/expressions-india-backend/internal/dto"
	"github.com/dhruvpurohit2k/expressions-india-backend/internal/models"
	"github.com/dhruvpurohit2k/expressions-india-backend/internal/pkg/utils"
	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Service struct {
	db *gorm.DB
}

func (s *Service) CreatePodcast(o *dto.PodcastCreateDTO) error {
	podcast := models.Podcast{
		ID:          uuid.Must(uuid.NewV7()).String(),
		Title:       o.Title,
		Link:        o.Link,
		Description: &o.Description,
		Tags:        datatypes.JSON(o.Tags),
		Transcript:  o.Transcript,
	}
	if len(o.Audiences) > 0 {
		var audiences []models.Audience
		if err := s.db.Where("name IN ?", o.Audiences).Find(&audiences).Error; err != nil {
			return err
		}
		podcast.Audiences = audiences
	}
	return s.db.Create(&podcast).Error
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
			Link:      podcast.Link,
			CreatedAt: podcast.CreatedAt,
		}
		podcastDTOs = append(podcastDTOs, data)
	}
	return podcastDTOs, nil
}

func (s *Service) GetPodcastList(filter utils.PodcastFilter) ([]dto.PodcastListItemDTO, int64, error) {
	var podcasts []models.Podcast
	var total int64

	base := s.db.Model(&models.Podcast{})
	if filter.Search != "" {
		base = base.Where("title LIKE ?", "%"+filter.Search+"%")
	}

	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	order := "created_at DESC"
	if filter.SortOrder == "asc" {
		order = "created_at ASC"
	}

	if err := base.Order(order).Limit(filter.Limit).Offset(filter.Offset).Find(&podcasts).Error; err != nil {
		return nil, 0, err
	}

	result := make([]dto.PodcastListItemDTO, 0, len(podcasts))
	for _, podcast := range podcasts {
		result = append(result, dto.PodcastListItemDTO{
			ID:        podcast.ID,
			Title:     podcast.Title,
			Link:      podcast.Link,
			CreatedAt: podcast.CreatedAt,
		})
	}
	return result, total, nil
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

func (s *Service) GetPodcastsByAudience(audience string, limit int, offset int) ([]dto.PodcastListItemDTO, int64, error) {
	var podcasts []models.Podcast
	var total int64

	base := s.db.Model(&models.Podcast{}).
		Where(
			"podcasts.id IN (SELECT pa.podcast_id FROM podcast_audience pa JOIN audiences a ON a.id = pa.audience_id WHERE a.name = ? OR a.name = 'all') OR podcasts.id NOT IN (SELECT DISTINCT pa.podcast_id FROM podcast_audience pa)",
			audience,
		)

	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := base.Order("podcasts.created_at DESC").Limit(limit).Offset(offset).Find(&podcasts).Error; err != nil {
		return nil, 0, err
	}

	result := make([]dto.PodcastListItemDTO, 0, len(podcasts))
	for _, p := range podcasts {
		result = append(result, dto.PodcastListItemDTO{
			ID:        p.ID,
			Title:     p.Title,
			Link:      p.Link,
			CreatedAt: p.CreatedAt,
		})
	}
	return result, total, nil
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
