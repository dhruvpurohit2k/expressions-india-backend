package dto

import (
	"mime/multipart"
	"time"

	"github.com/dhruvpurohit2k/expressions-india-backend/internal/models"
)

type ArticleListItemDTO struct {
	ID           string    `json:"id"`
	Title        string    `json:"title"`
	Category     string    `json:"category"`
	ThumbnailURL *string   `json:"thumbnailUrl"`
	PublishedAt  time.Time `json:"publishedAt"`
}

type ArticleDTO struct {
	ID        string          `json:"id"`
	Title     string          `json:"title"`
	Content   string          `json:"content"`
	Category  string          `json:"category"`
	Audience  []models.Audience `json:"audience"`
	Medias    []models.Media    `json:"medias"`
	CreatedAt time.Time         `json:"createdAt"`
	UpdatedAt time.Time         `json:"updatedAt"`
}

type ArticleCreateRequestDTO struct {
	Title     string                  `form:"title" binding:"required"`
	Content   string                  `form:"content" binding:"required"`
	Category  string                  `form:"category" binding:"required"`
	Audiences []string                `form:"audiences"`
	Medias    []*multipart.FileHeader `form:"medias"`
}

type ArticleUpdateRequestDTO struct {
	Title          string                  `form:"title" binding:"required"`
	Content        string                  `form:"content" binding:"required"`
	Category       string                  `form:"category" binding:"required"`
	Audiences      []string                `form:"audiences"`
	Medias         []*multipart.FileHeader `form:"medias"`
	DeletedMediaIds []string               `form:"deletedMediaIds"`
}
