package dto

import (
	"time"

	"github.com/dhruvpurohit2k/expressions-india-backend/internal/models"
)

type EventListItemDTO struct {
	ID           string     `json:"id"`
	Title        string     `json:"title"`
	Status       string     `json:"status"`
	IsOnline     bool       `json:"isOnline"`
	IsPaid       bool       `json:"isPaid"`
	StartDate    time.Time  `json:"startDate"`
	EndDate      *time.Time `json:"endDate"`
	ThumbnailURL *string    `json:"thumbnailUrl"`
}
type EventDTO struct {
	models.Event
	Audiences []string `json:"audiences"`
}

type EventCreateRequestDTO struct {
	Title                 string     `form:"title" binding:"required"`
	Description           string     `form:"description"`
	Perks                 string     `form:"perks"`
	StartDate             time.Time  `form:"startDate" binding:"required"`
	EndDate               *time.Time `form:"endDate"`
	StartTime             *string    `form:"startTime"`
	EndTime               *string    `form:"endTime"`
	Location              *string    `form:"location"`
	RegistrationURL       *string    `form:"registrationUrl"`
	IsOnline              *bool      `form:"isOnline" binding:"required"`
	IsPaid                *bool      `form:"isPaid" binding:"required"`
	Price                 *int       `form:"price"`
	// Pre-uploaded files: each element is a JSON-encoded UploadedMediaRef.
	ThumbnailUpload               string   `form:"thumbnailUpload"`
	PromotionalMediaUploads       []string `form:"promotionalMediaUploads"`
	PromotionalDocumentUploads    []string `form:"promotionalDocumentUploads"`
	MediaUploads                  []string `form:"mediaUploads"`
	DocumentUploads               []string `form:"documentUploads"`
	VideoLinks                    []string `form:"videoLinks"`
	PromotionalVideoLinks         []string `form:"promotionalVideoLinks"`
	Audiences                     []string `form:"audiences" binding:"required"`
	Status                        *string  `form:"status" binding:"required"`
}

type EventUpdateRequestDTO struct {
	Title                      string     `form:"title" binding:"required"`
	Description                string     `form:"description"`
	Perks                      string     `form:"perks"`
	StartDate                  time.Time  `form:"startDate" binding:"required"`
	EndDate                    *time.Time `form:"endDate"`
	StartTime                  *string    `form:"startTime"`
	EndTime                    *string    `form:"endTime"`
	Location                   *string    `form:"location"`
	RegistrationURL            *string    `form:"registrationUrl"`
	IsOnline                   *bool      `form:"isOnline" binding:"required"`
	IsPaid                     *bool      `form:"isPaid" binding:"required"`
	Price                      *int       `form:"price"`
	DeletedMediaIds                  []string   `form:"deletedMediaIds"`
	DeletedDocumentIds               []string   `form:"deletedDocumentIds"`
	DeletedPromotionalMediaIds       []string   `form:"deletedPromotionalMediaIds"`
	DeletedPromotionalDocumentIds    []string   `form:"deletedPromotionalDocumentIds"`
	DeletedThumbnailId               *string    `form:"deletedThumbnailId"`
	// Pre-uploaded files: each element is a JSON-encoded UploadedMediaRef.
	ThumbnailUpload               string   `form:"thumbnailUpload"`
	PromotionalMediaUploads       []string `form:"promotionalMediaUploads"`
	PromotionalDocumentUploads    []string `form:"promotionalDocumentUploads"`
	MediaUploads                  []string `form:"mediaUploads"`
	DocumentUploads               []string `form:"documentUploads"`
	VideoLinks                    []string `form:"videoLinks"`
	PromotionalVideoLinks         []string `form:"promotionalVideoLinks"`
	Audiences                     []string `form:"audiences" binding:"required"`
	Status                        *string  `form:"status" binding:"required"`
}
