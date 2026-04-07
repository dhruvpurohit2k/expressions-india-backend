package event

import (
	"context"
	"errors"
	"fmt"
	"log"
	"mime/multipart"
	"time"

	"github.com/dhruvpurohit2k/expressions-india-backend/internal/dto"
	"github.com/dhruvpurohit2k/expressions-india-backend/internal/models"
	"github.com/dhruvpurohit2k/expressions-india-backend/internal/pkg/utils"
	"github.com/dhruvpurohit2k/expressions-india-backend/internal/storage"
	"github.com/google/uuid"
	"golang.org/x/sync/errgroup"
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

func (s *Service) CreateEvent(data *dto.EventCreateRequestDTO) (retErr error) {
	var newEvent models.Event
	// ID is assigned by the BeforeCreate hook on models.Event
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
	}
	if data.IsPaid != nil {
		newEvent.IsPaid = *data.IsPaid
		newEvent.Price = data.Price
	}
	newEvent.RegistrationURL = *data.RegistrationURL

	// Track whether thumbnail was persisted to DB so we can clean it up on failure.
	var thumbnailPersistedToDB bool

	defer func() {
		if retErr != nil {
			s.cleanupEventUploads(&newEvent, thumbnailPersistedToDB)
		}
	}()

	if data.Thumbnail != nil {
		if err := s.uploadThumbnail(&newEvent, data.Thumbnail); err != nil {
			return err
		}
		thumbnailPersistedToDB = true
	}

	if len(data.Audiences) > 0 {
		audienceRows, err := s.getAudience(data.Audiences)
		if err != nil {
			return err
		}
		newEvent.Audiences = audienceRows
	}

	if len(data.PromotionalMedia) > 0 {
		if err := s.appendUploadedMedia(&newEvent.PromotionalMedia, data.PromotionalMedia); err != nil {
			return err
		}
	}
	if len(data.Documents) > 0 {
		if err := s.appendUploadedMedia(&newEvent.Documents, data.Documents); err != nil {
			return err
		}
	}
	if len(data.Medias) > 0 {
		if err := s.appendUploadedMedia(&newEvent.Medias, data.Medias); err != nil {
			return err
		}
	}
	if len(data.VideoLinks) > 0 {
		videoLinks, err := s.getLink(data.VideoLinks)
		if err != nil {
			return err
		}
		newEvent.VideoLinks = videoLinks
	}

	return s.db.Create(&newEvent).Error
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

func (s *Service) UpdateEvent(id string, newData *dto.EventUpdateRequestDTO) (retErr error) {
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

	// Track newly created resources so we can clean them up if db.Save fails at the end.
	var newThumbnailID string
	var newS3Keys []string

	defer func() {
		if retErr != nil {
			// Roll back new thumbnail (DB record + S3 file).
			if newThumbnailID != "" {
				if err := s.db.Delete(&models.Media{}, "id = ?", newThumbnailID).Error; err != nil {
					log.Printf("DB cleanup failed for thumbnail %s: %v", newThumbnailID, err)
				}
				if err := s.s3.Delete(newThumbnailID); err != nil {
					log.Printf("S3 cleanup failed for thumbnail %s: %v", newThumbnailID, err)
				}
			}
			// Roll back new media S3 files (not yet in DB, so S3 only).
			for _, key := range newS3Keys {
				if err := s.s3.Delete(key); err != nil {
					log.Printf("S3 cleanup failed for key %s: %v", key, err)
				}
			}
		}
	}()

	if newData.DeletedThumbnailId != nil && *newData.DeletedThumbnailId != "" {
		if err := s.db.Delete(&models.Media{}, "id = ?", *newData.DeletedThumbnailId).Error; err != nil {
			return err
		}
		// Best-effort: DB record is already gone; don't fail the request if S3 delete fails.
		if err := s.s3.Delete(*newData.DeletedThumbnailId); err != nil {
			log.Printf("S3 cleanup failed for deleted thumbnail %s: %v", *newData.DeletedThumbnailId, err)
		}
		event.ThumbnailID = nil
		event.Thumbnail = nil
	}
	if newData.Thumbnail != nil {
		if err := s.uploadThumbnail(&event, newData.Thumbnail); err != nil {
			return err
		}
		newThumbnailID = event.Thumbnail.ID
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

	for _, mediaID := range newData.DeletedPromotionalMediaIds {
		if err := s.db.Delete(&models.Media{}, "id = ?", mediaID).Error; err != nil {
			return err
		}
		// Best-effort: DB record is gone; log but don't fail if S3 delete fails.
		if err := s.s3.Delete(mediaID); err != nil {
			log.Printf("S3 cleanup failed for deleted media %s: %v", mediaID, err)
		}
	}

	if len(newData.PromotionalMedia) > 0 {
		before := len(event.PromotionalMedia)
		if err := s.appendUploadedMedia(&event.PromotionalMedia, newData.PromotionalMedia); err != nil {
			return err
		}
		for _, m := range event.PromotionalMedia[before:] {
			newS3Keys = append(newS3Keys, m.ID)
		}
	}
	if len(newData.Documents) > 0 {
		before := len(event.Documents)
		if err := s.appendUploadedMedia(&event.Documents, newData.Documents); err != nil {
			return err
		}
		for _, m := range event.Documents[before:] {
			newS3Keys = append(newS3Keys, m.ID)
		}
	}
	if len(newData.Medias) > 0 {
		before := len(event.Medias)
		if err := s.appendUploadedMedia(&event.Medias, newData.Medias); err != nil {
			return err
		}
		for _, m := range event.Medias[before:] {
			newS3Keys = append(newS3Keys, m.ID)
		}
	}

	if len(newData.VideoLinks) > 0 {
		newLinks, err := s.getLink(newData.VideoLinks)
		if err != nil {
			return err
		}
		if err := s.db.Model(&event).Association("VideoLinks").Append(newLinks); err != nil {
			return err
		}
	}
	if len(newData.PromotionalVideoLinks) > 0 {
		newLinks, err := s.getLink(newData.PromotionalVideoLinks)
		if err != nil {
			return err
		}
		if err := s.db.Model(&event).Association("PromotionalVideoLinks").Append(newLinks); err != nil {
			return err
		}
	}

	return s.db.Save(&event).Error
}

// getAudience fetches audience records matching the given names in a single query.
func (s *Service) getAudience(audiences []string) ([]models.Audience, error) {
	var audienceRows []models.Audience
	if err := s.db.Where("name IN ?", audiences).Find(&audienceRows).Error; err != nil {
		return nil, err
	}
	return audienceRows, nil
}

// appendUploadedMedia uploads files to S3 concurrently and appends the resulting Media records to dest.
// If any upload fails, already-uploaded files are cleaned up before returning the error.
func (s *Service) appendUploadedMedia(dest *[]models.Media, files []*multipart.FileHeader) error {
	medias := make([]models.Media, len(files))
	g, _ := errgroup.WithContext(context.Background())
	for i, file := range files {
		i, file := i, file
		g.Go(func() error {
			f, err := file.Open()
			if err != nil {
				return fmt.Errorf("failed to open %s: %w", file.Filename, err)
			}
			defer f.Close()
			location, key, contentType, err := s.s3.UploadNetwork(f)
			if err != nil {
				return err
			}
			medias[i] = models.Media{ID: key, URL: location, FileType: contentType}
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		for _, m := range medias {
			if m.ID != "" {
				if delErr := s.s3.Delete(m.ID); delErr != nil {
					log.Printf("S3 cleanup failed for key %s: %v", m.ID, delErr)
				}
			}
		}
		return err
	}
	*dest = append(*dest, medias...)
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
		s.s3.Delete(key)
		return err
	}
	event.ThumbnailID = &media.ID
	event.Thumbnail = &media
	return nil
}

// cleanupEventUploads removes any S3 files (and DB thumbnail record if already created) on failure.
func (s *Service) cleanupEventUploads(event *models.Event, thumbnailPersistedToDB bool) {
	if event.Thumbnail != nil {
		if thumbnailPersistedToDB {
			if err := s.db.Delete(&models.Media{}, "id = ?", event.Thumbnail.ID).Error; err != nil {
				log.Printf("DB cleanup failed for thumbnail %s: %v", event.Thumbnail.ID, err)
			}
		}
		if err := s.s3.Delete(event.Thumbnail.ID); err != nil {
			log.Printf("S3 cleanup failed for thumbnail %s: %v", event.Thumbnail.ID, err)
		}
	}
	for _, m := range event.PromotionalMedia {
		if m.ID != "" {
			if err := s.s3.Delete(m.ID); err != nil {
				log.Printf("S3 cleanup failed for key %s: %v", m.ID, err)
			}
		}
	}
	for _, m := range event.Documents {
		if m.ID != "" {
			if err := s.s3.Delete(m.ID); err != nil {
				log.Printf("S3 cleanup failed for key %s: %v", m.ID, err)
			}
		}
	}
	for _, m := range event.Medias {
		if m.ID != "" {
			if err := s.s3.Delete(m.ID); err != nil {
				log.Printf("S3 cleanup failed for key %s: %v", m.ID, err)
			}
		}
	}
}

func (s *Service) DeleteEvent(id string) error {
	var event models.Event
	if err := s.db.Preload("Thumbnail").Preload("PromotionalMedia").Preload("Medias").Preload("Documents").Preload("VideoLinks").Preload("Audiences").First(&event, "id = ?", id).Error; err != nil {
		return err
	}

	allMedia := append(append(event.PromotionalMedia, event.Medias...), event.Documents...)
	videoLinks := event.VideoLinks

	// Clear all junction-table associations first.
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
		// Best-effort: DB record is gone. Log but don't fail the delete if S3 is unavailable.
		if err := s.s3.Delete(thumbnailID); err != nil {
			log.Printf("S3 delete failed for thumbnail %s: %v", thumbnailID, err)
		}
	}

	for _, media := range allMedia {
		if err := s.db.Delete(&models.Media{}, "id = ?", media.ID).Error; err != nil {
			return err
		}
		// Best-effort S3 cleanup.
		if err := s.s3.Delete(media.ID); err != nil {
			log.Printf("S3 delete failed for media %s: %v", media.ID, err)
		}
	}

	for _, link := range videoLinks {
		if err := s.db.Delete(&models.Link{}, "id = ?", link.ID).Error; err != nil {
			return err
		}
	}

	return s.db.Delete(&event).Error
}

func (s *Service) GetHomePageImages() ([]string, error) {
	var upcomingEvents []models.Event
	var pastEvents []models.Event

	if err := s.db.Model(&models.Event{}).
		Where("status = ?", "upcoming").
		Preload("Thumbnail").
		Order("start_date ASC").
		Limit(3).
		Find(&upcomingEvents).Error; err != nil {
		return nil, err
	}

	if err := s.db.Model(&models.Event{}).
		Where("status = ?", "completed").
		Preload("Thumbnail").
		Order("end_date DESC").
		Limit(3).
		Find(&pastEvents).Error; err != nil {
		return nil, err
	}

	var urls []string
	for _, e := range upcomingEvents {
		if e.Thumbnail != nil && e.Thumbnail.URL != "" {
			urls = append(urls, e.Thumbnail.URL)
		}
	}
	for _, e := range pastEvents {
		if e.Thumbnail != nil && e.Thumbnail.URL != "" {
			urls = append(urls, e.Thumbnail.URL)
		}
	}
	return urls, nil
}

func (s *Service) getLink(links []string) ([]models.Link, error) {
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
