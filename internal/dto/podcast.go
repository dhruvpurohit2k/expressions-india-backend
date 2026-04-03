package dto

import "time"

type PodcastListItemDTO struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"createdAt"`
}

type PodcastDTO struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Link        string   `json:"link"`
	Description *string  `json:"description"`
	Transcript  *string  `json:"transcript"`
	Tags        string   `json:"tags"`
	Audiences   []string `json:"audiences"`
}

type PodcastCreateDTO struct {
	Title       string   `form:"title" binding:"required"`
	Link        string   `form:"link" binding:"required"`
	Description string   `form:"description"`
	Tags        string   `form:"tags"`
	Transcript  *string  `form:"transcript"`
	Audiences   []string `form:"audiences"`
}
