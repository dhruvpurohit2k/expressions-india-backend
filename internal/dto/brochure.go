package dto

import (
	"time"

	"github.com/dhruvpurohit2k/expressions-india-backend/internal/models"
)

type BrochureListItemDTO struct {
	ID           string    `json:"id"`
	Title        string    `json:"title"`
	Description  string    `json:"description"`
	ThumbnailURL *string   `json:"thumbnailUrl"`
	PDFURL       *string   `json:"pdfUrl"`
	CreatedAt    time.Time `json:"createdAt"`
}

type BrochureDTO struct {
	ID          string        `json:"id"`
	Title       string        `json:"title"`
	Description string        `json:"description"`
	PDF         *models.Media `json:"pdf"`
	Thumbnail   *models.Media `json:"thumbnail"`
	CreatedAt   time.Time     `json:"createdAt"`
}

type BrochureCreateDTO struct {
	Title           string `form:"title" binding:"required"`
	Description     string `form:"description"`
	ThumbnailUpload string `form:"thumbnailUpload"`
	PDFUpload       string `form:"pdfUpload"`
}

type BrochureUpdateDTO struct {
	Title              string  `form:"title" binding:"required"`
	Description        string  `form:"description"`
	ThumbnailUpload    string  `form:"thumbnailUpload"`
	PDFUpload          string  `form:"pdfUpload"`
	DeletedThumbnailId *string `form:"deletedThumbnailId"`
	DeletedPDFId       *string `form:"deletedPDFId"`
}
