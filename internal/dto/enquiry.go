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

type EnquiryCreateDTO struct {
	Subject string `form:"subject" json:"subject"`
	Name    string `form:"name" json:"name"`
	Email   string `form:"email" json:"email"`
	Phone   string `form:"phone" json:"phone"`
	Message string `form:"message" json:"message"`
}
