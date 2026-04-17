package brochure

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

func (s *Service) GetList(limit, offset int) ([]dto.BrochureListItemDTO, int64, error) {
	var brochures []models.Brochure
	var total int64

	base := s.db.Model(&models.Brochure{})
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := base.Order("created_at DESC").Limit(limit).Offset(offset).
		Preload("PDF").Preload("Thumbnail").Find(&brochures).Error; err != nil {
		return nil, 0, err
	}

	result := make([]dto.BrochureListItemDTO, 0, len(brochures))
	for _, b := range brochures {
		item := dto.BrochureListItemDTO{
			ID:          b.ID,
			Title:       b.Title,
			Description: b.Description,
			CreatedAt:   b.CreatedAt,
		}
		if b.Thumbnail != nil {
			item.ThumbnailURL = &b.Thumbnail.URL
		}
		if b.PDF != nil {
			item.PDFURL = &b.PDF.URL
		}
		result = append(result, item)
	}
	return result, total, nil
}

func (s *Service) GetById(id string) (*dto.BrochureDTO, error) {
	var b models.Brochure
	if err := s.db.Where("id = ?", id).
		Preload("PDF").Preload("Thumbnail").
		First(&b).Error; err != nil {
		return nil, err
	}
	return &dto.BrochureDTO{
		ID:          b.ID,
		Title:       b.Title,
		Description: b.Description,
		PDF:         b.PDF,
		Thumbnail:   b.Thumbnail,
		CreatedAt:   b.CreatedAt,
	}, nil
}

func (s *Service) Create(data *dto.BrochureCreateDTO) error {
	b := models.Brochure{
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
		b.ThumbnailID = &media.ID
		b.Thumbnail = &media
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
		b.PDFID = &media.ID
		b.PDF = &media
	}

	return s.db.Create(&b).Error
}

func (s *Service) Update(id string, data *dto.BrochureUpdateDTO) error {
	var b models.Brochure
	if err := s.db.Where("id = ?", id).First(&b).Error; err != nil {
		return err
	}
	b.Title = data.Title
	b.Description = data.Description

	if data.DeletedThumbnailId != nil && *data.DeletedThumbnailId != "" {
		if err := s.db.Delete(&models.Media{}, "id = ?", *data.DeletedThumbnailId).Error; err != nil {
			return err
		}
		if err := s.s3.Delete(*data.DeletedThumbnailId); err != nil {
			log.Printf("S3 cleanup failed for thumbnail %s: %v", *data.DeletedThumbnailId, err)
		}
		b.ThumbnailID = nil
		b.Thumbnail = nil
	}

	if data.DeletedPDFId != nil && *data.DeletedPDFId != "" {
		if err := s.db.Delete(&models.Media{}, "id = ?", *data.DeletedPDFId).Error; err != nil {
			return err
		}
		if err := s.s3.Delete(*data.DeletedPDFId); err != nil {
			log.Printf("S3 cleanup failed for PDF %s: %v", *data.DeletedPDFId, err)
		}
		b.PDFID = nil
		b.PDF = nil
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
		b.ThumbnailID = &media.ID
		b.Thumbnail = &media
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
		b.PDFID = &media.ID
		b.PDF = &media
	}

	return s.db.Save(&b).Error
}

func (s *Service) Delete(id string) error {
	var b models.Brochure
	if err := s.db.Where("id = ?", id).Preload("PDF").Preload("Thumbnail").First(&b).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		return err
	}

	if b.Thumbnail != nil {
		thumbID := b.Thumbnail.ID
		b.ThumbnailID = nil
		if err := s.db.Save(&b).Error; err != nil {
			return err
		}
		if err := s.db.Delete(&models.Media{}, "id = ?", thumbID).Error; err != nil {
			return err
		}
		if err := s.s3.Delete(thumbID); err != nil {
			log.Printf("S3 cleanup failed for thumbnail %s: %v", thumbID, err)
		}
	}

	if b.PDF != nil {
		pdfID := b.PDF.ID
		b.PDFID = nil
		if err := s.db.Save(&b).Error; err != nil {
			return err
		}
		if err := s.db.Delete(&models.Media{}, "id = ?", pdfID).Error; err != nil {
			return err
		}
		if err := s.s3.Delete(pdfID); err != nil {
			log.Printf("S3 cleanup failed for PDF %s: %v", pdfID, err)
		}
	}

	return s.db.Delete(&b).Error
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
