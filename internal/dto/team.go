package dto

import "time"

type MemberDTO struct {
	ID       string   `json:"id"`
	Position string   `json:"position"`
	Holders  []string `json:"holders"`
}

type TeamListItemDTO struct {
	ID          string    `json:"id"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"createdAt"`
}

type TeamDTO struct {
	ID          string      `json:"id"`
	Description string      `json:"description"`
	Members     []MemberDTO `json:"members"`
	CreatedAt   time.Time   `json:"createdAt"`
	UpdatedAt   time.Time   `json:"updatedAt"`
}

type MemberInputDTO struct {
	Position string   `json:"position" binding:"required"`
	Holders  []string `json:"holders"`
}

type TeamCreateDTO struct {
	Description string           `json:"description"`
	Members     []MemberInputDTO `json:"members"`
}

type TeamUpdateDTO struct {
	Description string           `json:"description"`
	Members     []MemberInputDTO `json:"members"`
}
