package audience

import (
	"github.com/dhruvpurohit2k/expressions-india-backend/internal/dto"
	"github.com/dhruvpurohit2k/expressions-india-backend/internal/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

func (s *Service) GetAudienceByName(name string) (*dto.AudieceListItemDTO, error) {
	var audience models.Audience
	if err := s.db.Where("name = ?", name).First(&audience).Error; err != nil {
		return nil, err
	}
	return &dto.AudieceListItemDTO{
		ID:          audience.ID,
		Name:        audience.Name,
		Description: audience.Description,
	}, nil
}

func (s *Service) UpdateAudienceDescription(id string, description string) error {
	var audience models.Audience
	if err := s.db.First(&audience, id).Error; err != nil {
		return err
	}
	return s.db.Model(&audience).Update("description", description).Error
}

func (s *Service) GetAudience(ctx *gin.Context) ([]dto.AudieceListItemDTO, error) {
	var audiences []models.Audience
	if err := s.db.Where("Name NOT LIKE ?", "all").Find(&audiences).Error; err != nil {
		return nil, err
	}
	var result []dto.AudieceListItemDTO
	for _, audience := range audiences {
		result = append(result, dto.AudieceListItemDTO{
			ID:          audience.ID,
			Name:        audience.Name,
			Description: audience.Description,
		})
	}
	return result, nil
}
