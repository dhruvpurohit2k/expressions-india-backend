package dto

type AudieceListItemDTO struct {
	ID          uint   `json:"ID"`
	Name        string `json:"name"`
	Description string `json:"introduction"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
}

type AudienceUpdateDescriptionDTO struct {
	Description string `json:"description" binding:"required"`
}
