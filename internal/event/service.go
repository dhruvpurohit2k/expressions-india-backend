package event

import (
	"errors"
	"log"
	"mime/multipart"
	"time"

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

func (s Service) GetUpcomingEvents(limit int, offset int) ([]dto.EventListItemDTO, int64, error) {
	var events []models.Event
	var total int64

	base := s.db.Model(&models.Event{}).
		Where("status = ?", "upcoming").
		Where("start_date >= ?", time.Now())

	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := base.Preload("Thumbnail").Limit(limit).Offset(offset).Find(&events).Error; err != nil {
		return nil, 0, err
	}

	var result []dto.EventListItemDTO
	for _, event := range events {
		item := dto.EventListItemDTO{
			ID:        event.ID,
			Title:     event.Title,
			StartDate: event.StartDate,
			EndDate:   event.EndDate,
		}
		if event.Thumbnail != nil {
			item.ThumbnailURL = &event.Thumbnail.URL
		}
		result = append(result, item)
	}
	return result, total, nil
}

func (s Service) GetPastEvents(limit int, offset int) ([]dto.EventListItemDTO, int64, error) {
	var events []models.Event
	var total int64

	base := s.db.Model(&models.Event{}).
		Where("status = ?", "completed")

	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := base.Preload("Thumbnail").Order("end_date DESC").Limit(limit).Offset(offset).Find(&events).Error; err != nil {
		return nil, 0, err
	}

	var result []dto.EventListItemDTO
	for _, event := range events {
		item := dto.EventListItemDTO{
			ID:        event.ID,
			Title:     event.Title,
			StartDate: event.StartDate,
			EndDate:   event.EndDate,
		}
		if event.Thumbnail != nil {
			item.ThumbnailURL = &event.Thumbnail.URL
		}
		result = append(result, item)
	}
	return result, total, nil
}

func NewService(db *gorm.DB, s3 *storage.S3) *Service {

	return &Service{
		db: db,
		s3: s3,
	}
}

func (s *Service) GetAllEvents() ([]models.Event, error) {
	var events []models.Event
	err := s.db.Preload("Thumbnail").
		Preload("Medias").
		Preload("Documents").
		Preload("VideoLinks").
		Preload("PromotionalMedia").
		Preload("Audiences").
		Find(&events).Error

	return events, err
}

func (s *Service) GetUpcomingEventsByAudience(audience string, limit int, offset int) ([]dto.EventListItemDTO, int64, error) {
	var events []models.Event
	var total int64

	base := s.db.Model(&models.Event{}).
		Where("status = ?", "upcoming").
		Where(
			"events.id IN (SELECT ea.event_id FROM event_audience ea JOIN audiences a ON a.id = ea.audience_id WHERE a.name = ? OR a.name = 'all') OR events.id NOT IN (SELECT DISTINCT ea.event_id FROM event_audience ea)",
			audience,
		)

	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := base.Preload("Thumbnail").Limit(limit).Offset(offset).Find(&events).Error; err != nil {
		return nil, 0, err
	}

	result := make([]dto.EventListItemDTO, 0, len(events))
	for _, e := range events {
		item := dto.EventListItemDTO{
			ID:        e.ID,
			Title:     e.Title,
			IsOnline:  e.IsOnline,
			IsPaid:    e.IsPaid,
			StartDate: e.StartDate,
			EndDate:   e.EndDate,
		}
		if e.Thumbnail != nil {
			item.ThumbnailURL = &e.Thumbnail.URL
		}
		result = append(result, item)
	}
	return result, total, nil
}

func (s *Service) GetEventById(id string) (*dto.EventDTO, error) {
	var event models.Event
	err := s.db.Where("id = ?", id).
		Preload("Thumbnail").
		Preload("Medias").
		Preload("Documents").
		Preload("VideoLinks").
		Preload("PromotionalVideoLinks").
		Preload("PromotionalMedia").
		Preload("Audiences").
		First(&event).Error

	var audiences []string
	for _, audience := range event.Audiences {
		audiences = append(audiences, audience.Name)
	}
	result := &dto.EventDTO{
		Event:     event,
		Audiences: audiences,
	}

	return result, err
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
	newEvent.Status = data.Status
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

	if data.Thumbnail != nil {
		if err := s.uploadThumbnail(&newEvent, data.Thumbnail); err != nil {
			return err
		}
	}

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

func (s *Service) GetEventList(eventFilter utils.Filter) ([]dto.EventListItemDTO, int64, error) {
	var events []models.Event
	var total int64

	base := utils.ApplyEventListFilters(s.db.Model(&models.Event{}), eventFilter)

	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	query := utils.ApplyEventListFilters(s.db.Model(&models.Event{}).Preload("Thumbnail"), eventFilter)
	if err := query.Find(&events).Error; err != nil {
		return nil, 0, err
	}

	var eventList []dto.EventListItemDTO
	for _, event := range events {
		status := ""
		if event.Status != nil {
			status = *event.Status
		}
		item := dto.EventListItemDTO{
			ID:        event.ID,
			Title:     event.Title,
			Status:    status,
			IsOnline:  event.IsOnline,
			IsPaid:    event.IsPaid,
			StartDate: event.StartDate,
			EndDate:   event.EndDate,
		}
		if event.Thumbnail != nil {
			item.ThumbnailURL = &event.Thumbnail.URL
		}
		eventList = append(eventList, item)
	}

	return eventList, total, nil
}

func (s *Service) UpdateEvent(id string, newData *dto.EventUpdateRequestDTO) error {
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
	event.Status = newData.Status
	if newData.IsOnline != nil {
		event.IsOnline = *newData.IsOnline
	} else {
		event.IsOnline = false
	}
	if newData.IsPaid != nil {
		event.IsPaid = *newData.IsPaid
		if *newData.IsPaid {
			if newData.Price != nil {
				event.Price = newData.Price
			} else {
				return errors.New("price is required for paid events")
			}
		} else {
			event.Price = nil
		}
	} else {
		event.IsPaid = false
		event.Price = nil
	}
	event.RegistrationURL = *newData.RegistrationURL

	if newData.DeletedThumbnailId != nil && *newData.DeletedThumbnailId != "" {
		if err := s.db.Delete(&models.Media{}, "id = ?", *newData.DeletedThumbnailId).Error; err != nil {
			return err
		}
		if err := s.s3.Delete(*newData.DeletedThumbnailId); err != nil {
			return err
		}
		event.ThumbnailID = nil
		event.Thumbnail = nil
	}
	if newData.Thumbnail != nil {
		if err := s.uploadThumbnail(&event, newData.Thumbnail); err != nil {
			return err
		}
	}

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
	if len(newData.VideoLinks) > 0 {
		newLinks, err := s.getLink(event.ID, newData.VideoLinks)
		if err != nil {
			return err
		}
		if err := s.db.Model(&event).Association("VideoLinks").Append(newLinks); err != nil {
			return err
		}
	}
	if len(newData.PromotionalVideoLinks) > 0 {
		newLinks, err := s.getLink(event.ID, newData.PromotionalVideoLinks)
		if err != nil {
			return err
		}
		if err := s.db.Model(&event).Association("PromotionalVideoLinks").Append(newLinks); err != nil {
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
		location, key, contentType, err := s.s3.UploadNetwork(f)
		if err != nil {
			return err
		}
		event.PromotionalMedia = append(event.PromotionalMedia, models.Media{ID: key, URL: location, FileType: contentType})
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
		location, key, contentType, err := s.s3.UploadNetwork(f)
		if err != nil {
			return err
		}
		event.Documents = append(event.Documents, models.Media{ID: key, URL: location, FileType: contentType})
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
		location, key, contentType, err := s.s3.UploadNetwork(f)
		if err != nil {
			return err
		}
		event.Medias = append(event.Medias, models.Media{ID: key, URL: location, FileType: contentType})
		f.Close()
	}

	return nil
}

func (s *Service) uploadThumbnail(event *models.Event, file *multipart.FileHeader) error {
	f, err := file.Open()
	if err != nil {
		return err
	}
	defer f.Close()
	location, key, contentType, err := s.s3.UploadNetwork(f)
	if err != nil {
		return err
	}
	media := models.Media{ID: key, URL: location, FileType: contentType}
	if err := s.db.Create(&media).Error; err != nil {
		return err
	}
	event.ThumbnailID = &media.ID
	event.Thumbnail = &media
	return nil
}

func (s *Service) DeleteEvent(id string) error {
	var event models.Event
	if err := s.db.Preload("Thumbnail").Preload("PromotionalMedia").Preload("Medias").Preload("Documents").Preload("VideoLinks").Preload("Audiences").First(&event, "id = ?", id).Error; err != nil {
		return err
	}

	allMedia := append(append(event.PromotionalMedia, event.Medias...), event.Documents...)
	videoLinks := event.VideoLinks

	if err := s.db.Model(&event).Association("PromotionalMedia").Clear(); err != nil {
		return err
	}
	if err := s.db.Model(&event).Association("Medias").Clear(); err != nil {
		return err
	}
	if err := s.db.Model(&event).Association("Documents").Clear(); err != nil {
		return err
	}
	if err := s.db.Model(&event).Association("VideoLinks").Clear(); err != nil {
		return err
	}
	if err := s.db.Model(&event).Association("Audiences").Clear(); err != nil {
		return err
	}

	if event.Thumbnail != nil {
		thumbnailID := event.Thumbnail.ID
		event.ThumbnailID = nil
		if err := s.db.Save(&event).Error; err != nil {
			return err
		}
		if err := s.db.Delete(&models.Media{}, "id = ?", thumbnailID).Error; err != nil {
			return err
		}
		if err := s.s3.Delete(thumbnailID); err != nil {
			return err
		}
	}

	for _, media := range allMedia {
		if err := s.db.Delete(&models.Media{}, "id = ?", media.ID).Error; err != nil {
			return err
		}
		if err := s.s3.Delete(media.ID); err != nil {
			return err
		}
	}

	for _, link := range videoLinks {
		if err := s.db.Delete(&models.Link{}, "id = ?", link.ID).Error; err != nil {
			return err
		}
	}

	return s.db.Delete(&event).Error
}

func (s *Service) getLink(eventID string, links []string) ([]models.Link, error) {
	var videoLinks []models.Link
	for _, link := range links {
		link := models.Link{
			ID:  uuid.Must(uuid.NewV7()).String(),
			URL: link,
		}
		videoLinks = append(videoLinks, link)
	}
	return videoLinks, nil
}
