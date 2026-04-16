package dto

import "time"

type EnquiryListItemDTO struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Subject     string    `json:"subject"`
	Designation string    `json:"designation"`
	Email       string    `json:"email"`
	Phone       string    `json:"phone"`
	CreatedAt   time.Time `json:"createdAt"`
}

type EnquiryDetailDTO struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Subject     string    `json:"subject"`
	Designation string    `json:"designation"`
	Email       string    `json:"email"`
	Phone       string    `json:"phone"`
	Message     string    `json:"message"`
	CreatedAt   time.Time `json:"createdAt"`
}

type EnquiryCreateDTO struct {
	Name        string `form:"name" json:"name" binding:"required"`
	Subject     string `form:"subject" json:"subject" binding:"required"`
	Designation string `form:"designation" json:"designation" binding:"required"`
	Email       string `form:"email" json:"email" binding:"required,email"`
	Phone       string `form:"phone" json:"phone" binding:"required"`
	Message     string `form:"message" json:"message" binding:"required"`
}
