package enquiry

import (
	"github.com/dhruvpurohit2k/expressions-india-backend/internal/dto"
	"github.com/dhruvpurohit2k/expressions-india-backend/internal/models"
	"gorm.io/gorm"
)

type Service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

func (s *Service) GetEnquiryList() ([]dto.EnquiryListItemDTO, error) {
	var enquiries []models.Enquiry
	if err := s.db.Find(&enquiries).Error; err != nil {
		return nil, err
	}
	var result []dto.EnquiryListItemDTO
	for _, enquiry := range enquiries {
		result = append(result, dto.EnquiryListItemDTO{
			ID:        enquiry.ID,
			Subject:   enquiry.Subject,
			Name:      enquiry.Name,
			Email:     enquiry.Email,
			Phone:     enquiry.Phone,
			CreatedAt: enquiry.CreatedAt,
		})
	}
	return result, nil
}

func (s *Service) GetEnquiryById(id string) (*models.Enquiry, error) {
	var enquiry models.Enquiry
	if err := s.db.First(&enquiry, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &enquiry, nil
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
