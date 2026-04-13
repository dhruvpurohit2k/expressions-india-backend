package enquiry

import (
	"github.com/dhruvpurohit2k/expressions-india-backend/internal/dto"
	"github.com/dhruvpurohit2k/expressions-india-backend/internal/models"
	"github.com/dhruvpurohit2k/expressions-india-backend/internal/pkg/utils"
	"gorm.io/gorm"
)

type Service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

func (s *Service) GetEnquiryListFiltered(filter utils.EnquiryFilter) ([]dto.EnquiryListItemDTO, int64, error) {
	base := s.db.Model(&models.Enquiry{})
	if filter.Name != "" {
		base = base.Where("LOWER(name) LIKE LOWER(?)", "%"+filter.Name+"%")
	}
	if filter.Email != "" {
		base = base.Where("LOWER(email) LIKE LOWER(?)", "%"+filter.Email+"%")
	}
	if filter.Phone != "" {
		base = base.Where("phone LIKE ?", "%"+filter.Phone+"%")
	}

	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var enquiries []models.Enquiry
	if err := base.Order("created_at DESC").Limit(filter.Limit).Offset(filter.Offset).Find(&enquiries).Error; err != nil {
		return nil, 0, err
	}

	result := make([]dto.EnquiryListItemDTO, 0, len(enquiries))
	for _, e := range enquiries {
		result = append(result, dto.EnquiryListItemDTO{
			ID:        e.ID,
			Subject:   e.Subject,
			Name:      e.Name,
			Email:     e.Email,
			Phone:     e.Phone,
			CreatedAt: e.CreatedAt,
		})
	}
	return result, total, nil
}

func (s *Service) GetEnquiryById(id string) (*dto.EnquiryDetailDTO, error) {
	var enquiry models.Enquiry
	if err := s.db.First(&enquiry, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &dto.EnquiryDetailDTO{
		ID:        enquiry.ID,
		Subject:   enquiry.Subject,
		Name:      enquiry.Name,
		Email:     enquiry.Email,
		Phone:     enquiry.Phone,
		Message:   enquiry.Message,
		CreatedAt: enquiry.CreatedAt,
	}, nil
}

func (s *Service) DeleteEnquiry(id string) error {
	return s.db.Delete(&models.Enquiry{}, "id = ?", id).Error
}

func (s *Service) CreateEnquiry(enquiry *dto.EnquiryCreateDTO) error {
	return s.db.Create(&models.Enquiry{
		Subject: enquiry.Subject,
		Name:    enquiry.Name,
		Email:   enquiry.Email,
		Phone:   enquiry.Phone,
		Message: enquiry.Message,
	}).Error
}
