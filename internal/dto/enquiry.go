package dto

import "time"

type EnquiryListItemDTO struct {
	ID        string    `json:"id"`
	Subject   string    `json:"subject"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Phone     string    `json:"phone"`
	CreatedAt time.Time `json:"createdAt"`
}

type EnquiryDetailDTO struct {
	ID        string    `json:"id"`
	Subject   string    `json:"subject"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Phone     string    `json:"phone"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"createdAt"`
}

type EnquiryCreateDTO struct {
	Subject string `form:"subject" json:"subject" binding:"required"`
	Name    string `form:"name" json:"name" binding:"required"`
	Email   string `form:"email" json:"email" binding:"required,email"`
	Phone   string `form:"phone" json:"phone" binding:"required"`
	Message string `form:"message" json:"message" binding:"required"`
}
