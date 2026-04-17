package dto

import (
	"time"

	"github.com/dhruvpurohit2k/expressions-india-backend/internal/models"
)

type AlmanacListItemDTO struct {
	ID           string    `json:"id"`
	Title        string    `json:"title"`
	Description  string    `json:"description"`
	ThumbnailURL *string   `json:"thumbnailUrl"`
	PDFURL       *string   `json:"pdfUrl"`
	CreatedAt    time.Time `json:"createdAt"`
}

type AlmanacDTO struct {
	ID          string        `json:"id"`
	Title       string        `json:"title"`
	Description string        `json:"description"`
	PDF         *models.Media `json:"pdf"`
	Thumbnail   *models.Media `json:"thumbnail"`
	CreatedAt   time.Time     `json:"createdAt"`
}

type AlmanacCreateDTO struct {
	Title           string `form:"title" binding:"required"`
	Description     string `form:"description"`
	ThumbnailUpload string `form:"thumbnailUpload"`
	PDFUpload       string `form:"pdfUpload"`
}

type AlmanacUpdateDTO struct {
	Title              string  `form:"title" binding:"required"`
	Description        string  `form:"description"`
	ThumbnailUpload    string  `form:"thumbnailUpload"`
	PDFUpload          string  `form:"pdfUpload"`
	DeletedThumbnailId *string `form:"deletedThumbnailId"`
	DeletedPDFId       *string `form:"deletedPDFId"`
}
