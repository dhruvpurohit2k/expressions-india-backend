package dto

import "time"

type PodcastListItemDTO struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Link      string    `json:"link"`
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

type PodcastUpdateDTO struct {
	Title       string   `json:"title" binding:"required"`
	Link        string   `json:"link" binding:"required"`
	Description string   `json:"description"`
	Tags        string   `json:"tags"`
	Transcript  *string  `json:"transcript"`
	Audiences   []string `json:"audiences"`
}
