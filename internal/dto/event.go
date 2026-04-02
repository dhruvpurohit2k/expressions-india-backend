package dto

import (
	"mime/multipart"
	"time"

	"github.com/dhruvpurohit2k/expressions-india-backend/internal/models"
)

type EventListItemDTO struct {
	ID        string     `json:"id"`
	Title     string     `json:"title"`
	IsOnline  bool       `json:"isOnline"`
	IsPaid    bool       `json:"isPaid"`
	StartDate time.Time  `json:"startDate"`
	EndDate   *time.Time `json:"endDate"`
}
type EventDTO struct {
	models.Event
	Audiences []string `json:"audiences"`
}
type EventCreateRequestDTO struct {
	Title            string                  `form:"title" binding:"required"`
	Description      string                  `form:"description"`
	Perks            string                  `form:"perks"`
	StartDate        time.Time               `form:"startDate" binding:"required"`
	EndDate          *time.Time              `form:"endDate"`
	StartTime        *string                 `form:"startTime"`
	EndTime          *string                 `form:"endTime"`
	Location         *string                 `form:"location"`
	RegistrationURL  *string                 `form:"registrationUrl" binding:"required"`
	IsOnline         *bool                   `form:"isOnline" binding:"required"`
	IsPaid           *bool                   `form:"isPaid" binding:"required"`
	Price            *int                    `form:"price"`
	PromotionalMedia []*multipart.FileHeader `form:"promotionalMedia"`
	Medias           []*multipart.FileHeader `form:"medias"`
	Documents        []*multipart.FileHeader `form:"documents"`
	VideoLinks       []string                `form:"videoLinks"`
	Audiences        []string                `form:"audiences" binding:"required"`
	Status           *string                 `form:"status" binding:"required"`
}

type EventUpdateRequestDTO struct {
	Title                      string                  `form:"title" binding:"required"`
	Description                string                  `form:"description"`
	Perks                      string                  `form:"perks"`
	StartDate                  time.Time               `form:"startDate" binding:"required"`
	EndDate                    *time.Time              `form:"endDate"`
	StartTime                  *string                 `form:"startTime"`
	EndTime                    *string                 `form:"endTime"`
	Location                   *string                 `form:"location"`
	RegistrationURL            *string                 `form:"registrationUrl" binding:"required"`
	IsOnline                   *bool                   `form:"isOnline" binding:"required"`
	IsPaid                     *bool                   `form:"isPaid" binding:"required"`
	Price                      *int                    `form:"price"`
	DeletedMediaIds            []string                `form:"deletedMediaIds"`
	DeletedDocumentIds         []string                `form:"deletedDocumentIds"`
	DeletedPromotionalMediaIds []string                `form:"deletedPromotionalMediaIds"`
	PromotionalMedia           []*multipart.FileHeader `form:"promotionalMedia"`
	Medias                     []*multipart.FileHeader `form:"medias"`
	Documents                  []*multipart.FileHeader `form:"documents"`
	VideoLinks                 []string                `form:"videoLinks"`
	Audiences                  []string                `form:"audiences" binding:"required"`
	Status                     *string                 `form:"status" binding:"required"`
}
