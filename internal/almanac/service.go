package almanac

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/dhruvpurohit2k/expressions-india-backend/internal/dto"
	"github.com/dhruvpurohit2k/expressions-india-backend/internal/models"
	"github.com/dhruvpurohit2k/expressions-india-backend/internal/storage"
	"gorm.io/gorm"
)

type Service struct {
	db *gorm.DB
	s3 *storage.S3
}

func NewService(db *gorm.DB, s3 *storage.S3) *Service {
	return &Service{db: db, s3: s3}
}

func (s *Service) GetList(limit, offset int) ([]dto.AlmanacListItemDTO, int64, error) {
	var almanacs []models.Almanac
	var total int64

	base := s.db.Model(&models.Almanac{})
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := base.Order("created_at DESC").Limit(limit).Offset(offset).
		Preload("PDF").Preload("Thumbnail").Find(&almanacs).Error; err != nil {
		return nil, 0, err
	}

	result := make([]dto.AlmanacListItemDTO, 0, len(almanacs))
	for _, a := range almanacs {
		item := dto.AlmanacListItemDTO{
			ID:          a.ID,
			Title:       a.Title,
			Description: a.Description,
			CreatedAt:   a.CreatedAt,
		}
		if a.Thumbnail != nil {
			item.ThumbnailURL = &a.Thumbnail.URL
		}
		if a.PDF != nil {
			item.PDFURL = &a.PDF.URL
		}
		result = append(result, item)
	}
	return result, total, nil
}

func (s *Service) GetById(id string) (*dto.AlmanacDTO, error) {
	var a models.Almanac
	if err := s.db.Where("id = ?", id).
		Preload("PDF").Preload("Thumbnail").
		First(&a).Error; err != nil {
		return nil, err
	}
	return &dto.AlmanacDTO{
		ID:          a.ID,
		Title:       a.Title,
		Description: a.Description,
		PDF:         a.PDF,
		Thumbnail:   a.Thumbnail,
		CreatedAt:   a.CreatedAt,
	}, nil
}

func (s *Service) Create(data *dto.AlmanacCreateDTO) error {
	a := models.Almanac{
		Title:       data.Title,
		Description: data.Description,
	}

	if data.ThumbnailUpload != "" {
		ref, err := parseRef(data.ThumbnailUpload)
		if err != nil {
			return fmt.Errorf("invalid thumbnailUpload: %w", err)
		}
		media := s.mediaFromRef(ref)
		if err := s.db.Create(&media).Error; err != nil {
			return err
		}
		a.ThumbnailID = &media.ID
		a.Thumbnail = &media
	}

	if data.PDFUpload != "" {
		ref, err := parseRef(data.PDFUpload)
		if err != nil {
			return fmt.Errorf("invalid pdfUpload: %w", err)
		}
		media := s.mediaFromRef(ref)
		if err := s.db.Create(&media).Error; err != nil {
			return err
		}
		a.PDFID = &media.ID
		a.PDF = &media
	}

	return s.db.Create(&a).Error
}

func (s *Service) Update(id string, data *dto.AlmanacUpdateDTO) error {
	var a models.Almanac
	if err := s.db.Where("id = ?", id).First(&a).Error; err != nil {
		return err
	}
	a.Title = data.Title
	a.Description = data.Description

	if data.DeletedThumbnailId != nil && *data.DeletedThumbnailId != "" {
		if err := s.db.Delete(&models.Media{}, "id = ?", *data.DeletedThumbnailId).Error; err != nil {
			return err
		}
		if err := s.s3.Delete(*data.DeletedThumbnailId); err != nil {
			log.Printf("S3 cleanup failed for thumbnail %s: %v", *data.DeletedThumbnailId, err)
		}
		a.ThumbnailID = nil
		a.Thumbnail = nil
	}

	if data.DeletedPDFId != nil && *data.DeletedPDFId != "" {
		if err := s.db.Delete(&models.Media{}, "id = ?", *data.DeletedPDFId).Error; err != nil {
			return err
		}
		if err := s.s3.Delete(*data.DeletedPDFId); err != nil {
			log.Printf("S3 cleanup failed for PDF %s: %v", *data.DeletedPDFId, err)
		}
		a.PDFID = nil
		a.PDF = nil
	}

	if data.ThumbnailUpload != "" {
		ref, err := parseRef(data.ThumbnailUpload)
		if err != nil {
			return fmt.Errorf("invalid thumbnailUpload: %w", err)
		}
		media := s.mediaFromRef(ref)
		if err := s.db.Create(&media).Error; err != nil {
			return err
		}
		a.ThumbnailID = &media.ID
		a.Thumbnail = &media
	}

	if data.PDFUpload != "" {
		ref, err := parseRef(data.PDFUpload)
		if err != nil {
			return fmt.Errorf("invalid pdfUpload: %w", err)
		}
		media := s.mediaFromRef(ref)
		if err := s.db.Create(&media).Error; err != nil {
			return err
		}
		a.PDFID = &media.ID
		a.PDF = &media
	}

	return s.db.Save(&a).Error
}

func (s *Service) Delete(id string) error {
	var a models.Almanac
	if err := s.db.Where("id = ?", id).Preload("PDF").Preload("Thumbnail").First(&a).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		return err
	}

	if a.Thumbnail != nil {
		thumbID := a.Thumbnail.ID
		a.ThumbnailID = nil
		if err := s.db.Save(&a).Error; err != nil {
			return err
		}
		if err := s.db.Delete(&models.Media{}, "id = ?", thumbID).Error; err != nil {
			return err
		}
		if err := s.s3.Delete(thumbID); err != nil {
			log.Printf("S3 cleanup failed for thumbnail %s: %v", thumbID, err)
		}
	}

	if a.PDF != nil {
		pdfID := a.PDF.ID
		a.PDFID = nil
		if err := s.db.Save(&a).Error; err != nil {
			return err
		}
		if err := s.db.Delete(&models.Media{}, "id = ?", pdfID).Error; err != nil {
			return err
		}
		if err := s.s3.Delete(pdfID); err != nil {
			log.Printf("S3 cleanup failed for PDF %s: %v", pdfID, err)
		}
	}

	return s.db.Delete(&a).Error
}

func (s *Service) mediaFromRef(ref dto.UploadedMediaRef) models.Media {
	return models.Media{
		ID:       ref.ID,
		Name:     ref.Name,
		URL:      s.s3.PublicURL(ref.ID),
		FileType: ref.FileType,
	}
}

func parseRef(jsonStr string) (dto.UploadedMediaRef, error) {
	var ref dto.UploadedMediaRef
	if err := json.Unmarshal([]byte(jsonStr), &ref); err != nil {
		return ref, fmt.Errorf("invalid media ref JSON: %w", err)
	}
	return ref, nil
}
