package team

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

func toMemberModels(inputs []dto.MemberInputDTO) []models.Member {
	members := make([]models.Member, 0, len(inputs))
	for _, m := range inputs {
		holders := m.Holders
		if holders == nil {
			holders = []string{}
		}
		members = append(members, models.Member{
			Position: m.Position,
			Holders:  holders,
		})
	}
	return members
}

func toMemberDTOs(members []models.Member) []dto.MemberDTO {
	result := make([]dto.MemberDTO, 0, len(members))
	for _, m := range members {
		holders := m.Holders
		if holders == nil {
			holders = []string{}
		}
		result = append(result, dto.MemberDTO{
			ID:       m.ID,
			Position: m.Position,
			Holders:  holders,
		})
	}
	return result
}

func (s *Service) GetList() ([]dto.TeamDTO, error) {
	var teams []models.Team
	if err := s.db.Preload("Members").Order("created_at ASC").Find(&teams).Error; err != nil {
		return nil, err
	}
	result := make([]dto.TeamDTO, 0, len(teams))
	for _, t := range teams {
		result = append(result, dto.TeamDTO{
			ID:          t.ID,
			Description: t.Description,
			Members:     toMemberDTOs(t.Members),
			CreatedAt:   t.CreatedAt,
			UpdatedAt:   t.UpdatedAt,
		})
	}
	return result, nil
}

func (s *Service) GetById(id string) (*dto.TeamDTO, error) {
	var t models.Team
	if err := s.db.Preload("Members").First(&t, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &dto.TeamDTO{
		ID:          t.ID,
		Description: t.Description,
		Members:     toMemberDTOs(t.Members),
		CreatedAt:   t.CreatedAt,
		UpdatedAt:   t.UpdatedAt,
	}, nil
}

func (s *Service) Create(req *dto.TeamCreateDTO) error {
	t := models.Team{
		Description: req.Description,
		Members:     toMemberModels(req.Members),
	}
	return s.db.Create(&t).Error
}

func (s *Service) Update(id string, req *dto.TeamUpdateDTO) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var t models.Team
		if err := tx.Preload("Members").First(&t, "id = ?", id).Error; err != nil {
			return err
		}

		// Collect old member IDs before clearing the association
		oldIDs := make([]string, len(t.Members))
		for i, m := range t.Members {
			oldIDs[i] = m.ID
		}

		if err := tx.Model(&t).Association("Members").Clear(); err != nil {
			return err
		}

		if len(oldIDs) > 0 {
			if err := tx.Delete(&models.Member{}, "id IN ?", oldIDs).Error; err != nil {
				return err
			}
		}

		t.Description = req.Description
		t.Members = toMemberModels(req.Members)
		return tx.Save(&t).Error
	})
}

func (s *Service) Delete(id string) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var t models.Team
		if err := tx.Preload("Members").First(&t, "id = ?", id).Error; err != nil {
			return err
		}

		memberIDs := make([]string, len(t.Members))
		for i, m := range t.Members {
			memberIDs[i] = m.ID
		}

		if err := tx.Model(&t).Association("Members").Clear(); err != nil {
			return err
		}

		if len(memberIDs) > 0 {
			if err := tx.Delete(&models.Member{}, "id IN ?", memberIDs).Error; err != nil {
				return err
			}
		}

		return tx.Delete(&t).Error
	})
}
