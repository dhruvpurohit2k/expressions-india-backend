package event

import (
	"errors"
	"log"
	"mime/multipart"

	"github.com/dhruvpurohit2k/expressions-india-backend/internal/dto"
	"github.com/dhruvpurohit2k/expressions-india-backend/internal/models"
	"github.com/dhruvpurohit2k/expressions-india-backend/internal/pkg/utils"
	"github.com/dhruvpurohit2k/expressions-india-backend/internal/storage"
	"github.com/google/uuid"
	"gorm.io/datatypes"
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
	err := s.db.Preload("Medias").
		Preload("Documents").
		Preload("VideoLinks").
		Preload("PromotionalMedia").
		Preload("Audiences").
		Find(&events).Error

	return events, err
}

func (s *Service) GetEventById(id string) (*models.Event, error) {
	var event models.Event
	err := s.db.Where("id = ?", id).Preload("Medias", "category = ?", "MEDIA").
		Preload("Documents", "category = ?", "DOCUMENT").
		Preload("VideoLinks").
		Preload("PromotionalMedia", "category = ?", "PROMOTION").
		Preload("Audiences").
		First(&event).Error

	return &event, err
}

func (s *Service) CreateEvent(data *dto.EventCreateRequestDTO) error {
	var newEvent models.Event
	EventID := uuid.Must(uuid.NewV7()).String()
	newEvent.ID = EventID
	newEvent.Title = data.Title
	newEvent.Description = data.Description
	newEvent.Perks = datatypes.JSON(data.Perks)
	newEvent.StartDate = data.StartDate
	newEvent.EndDate = data.EndDate
	newEvent.StartTime = data.StartTime
	newEvent.EndTime = data.EndTime
	newEvent.Location = data.Location
	if data.IsOnline != nil {
		newEvent.IsOnline = *data.IsOnline
	} else {
		newEvent.IsOnline = false
	}
	if data.IsPaid != nil {
		newEvent.IsPaid = *data.IsPaid
		newEvent.Price = data.Price
	} else {
		newEvent.IsPaid = false
		newEvent.Price = nil
	}
	newEvent.RegistrationURL = *data.RegistrationURL

	if len(data.Audiences) > 0 {
		audienceRows, err := s.getAudience(data.Audiences)
		if err != nil {
			return err
		}
		newEvent.Audiences = audienceRows
	}
	if len(data.PromotionalMedia) > 0 {
		if err := s.appendPromotionalMedia(&newEvent, data.PromotionalMedia); err != nil {
			return err
		}
	}
	if len(data.Documents) > 0 {
		if err := s.appendDocument(&newEvent, data.Documents); err != nil {
			return err
		}
	}
	if len(data.Medias) > 0 {
		if err := s.appendMedia(&newEvent, data.Medias); err != nil {
			return err
		}
	}
	if len(data.VideoLinks) > 0 {
		videoLinks, err := s.getLink(newEvent.ID, data.VideoLinks)
		if err != nil {
			return err
		}
		newEvent.VideoLinks = videoLinks
	}
	if err := s.db.Create(&newEvent).Error; err != nil {
		return err
	}
	return nil
}

func (s *Service) GetEventList(eventFilter utils.Filter) ([]dto.EventListItemDTO, error) {
	var events []models.Event

	query := utils.ApplyFilters(s.db.Model(&models.Event{}), eventFilter)

	err := query.Find(&events).Error

	var eventList []dto.EventListItemDTO

	for _, event := range events {
		eventList = append(eventList, dto.EventListItemDTO{
			ID:        event.ID,
			Title:     event.Title,
			IsOnline:  event.IsOnline,
			IsPaid:    event.IsPaid,
			Price:     event.Price,
			StartDate: event.StartDate,
			EndDate:   event.EndDate,
		})
	}

	return eventList, err
}

func (s Service) UpdateEvent(id string, newData *dto.EventUpdateRequestDTO) error {
	var event models.Event
	if err := s.db.First(&event, "id = ?", id).Error; err != nil {
		return err
	}
	event.Title = newData.Title
	event.Description = newData.Description
	event.Perks = datatypes.JSON(newData.Perks)
	event.StartDate = newData.StartDate
	event.EndDate = newData.EndDate
	event.StartTime = newData.StartTime
	event.EndTime = newData.EndTime
	event.Location = newData.Location
	if newData.IsOnline != nil {
		event.IsOnline = *newData.IsOnline
	} else {
		event.IsOnline = false
	}
	if newData.IsPaid != nil {
		event.IsPaid = *newData.IsPaid
		if newData.Price != nil {
			event.Price = newData.Price
		} else {
			return errors.New("price is required for paid events")
		}
	} else {
		event.IsPaid = false
		event.Price = nil
	}
	event.RegistrationURL = *newData.RegistrationURL

	if len(newData.Audiences) > 0 {
		audienceRows, err := s.getAudience(newData.Audiences)
		if err != nil {
			return err
		}
		if err := s.db.Model(&event).Association("Audiences").Replace(audienceRows); err != nil {
			return err
		}
	}
	for _, id := range newData.DeletedPromotionalMediaIds {
		if err := s.db.Delete(&models.Media{}, "id = ?", id).Error; err != nil {
			return err
		}
		if err := s.s3.Delete(id); err != nil {
			return err
		}
	}
	if len(newData.PromotionalMedia) > 0 {
		if err := s.appendPromotionalMedia(&event, newData.PromotionalMedia); err != nil {
			return err
		}
	}
	if len(newData.Documents) > 0 {
		if err := s.appendDocument(&event, newData.Documents); err != nil {
			return err
		}
	}
	if len(newData.Medias) > 0 {
		if err := s.appendMedia(&event, newData.Medias); err != nil {
			return err
		}
	}

	if err := s.db.Save(&event).Error; err != nil {
		return err
	}
	return nil
}

func (s *Service) getAudience(audiences []string) ([]models.Audience, error) {
	var audienceRows []models.Audience
	for _, audience := range audiences {
		var audienceRow models.Audience
		if err := s.db.Where("name = ?", audience).First(&audienceRow).Error; err != nil {
			return nil, err
		}
		audienceRows = append(audienceRows, audienceRow)
	}
	return audienceRows, nil
}

func (s *Service) appendPromotionalMedia(event *models.Event, files []*multipart.FileHeader) error {
	for _, file := range files {
		f, err := file.Open()
		if err != nil {
			log.Println(err)
			continue
		}
		location, key, err := s.s3.UploadNetwork(f)
		if err != nil {
			return err
		}
		event.PromotionalMedia = append(event.PromotionalMedia, models.Media{ID: key, URL: location, Key: key, EventID: event.ID, Category: "PROMOTION"})
		f.Close()
	}

	return nil
}
func (s *Service) appendDocument(event *models.Event, files []*multipart.FileHeader) error {
	for _, file := range files {
		f, err := file.Open()
		if err != nil {
			log.Println(err)
			continue
		}
		location, key, err := s.s3.UploadNetwork(f)
		if err != nil {
			return err
		}
		event.Documents = append(event.Documents, models.Media{ID: key, URL: location, Key: key, EventID: event.ID, Category: "DOCUMENT"})
		f.Close()
	}

	return nil
}
func (s *Service) appendMedia(event *models.Event, files []*multipart.FileHeader) error {
	for _, file := range files {
		f, err := file.Open()
		if err != nil {
			log.Println(err)
			continue
		}
		location, key, err := s.s3.UploadNetwork(f)
		if err != nil {
			return err
		}
		event.Medias = append(event.Medias, models.Media{ID: key, URL: location, Key: key, EventID: event.ID, Category: "MEDIA"})
		f.Close()
	}

	return nil
}

func (s *Service) getLink(eventID string, links []string) ([]models.Link, error) {
	var videoLinks []models.Link
	for _, link := range links {
		link := models.Link{
			ID:      uuid.Must(uuid.NewV7()).String(),
			URL:     link,
			EventID: eventID,
		}
		videoLinks = append(videoLinks, link)
	}
	return videoLinks, nil
}
