package dto

type EnquiryListItemDTO struct {
	Id        string `json:"id"`
	Subject   string `json:"subject"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
	CreatedAt string `json:"created_at"`
}

type EnquiryCreateDTO struct {
	Subject string `form:"subject" json:"subject"`
	Name    string `form:"name" json:"name"`
	Email   string `form:"email" json:"email"`
	Phone   string `form:"phone" json:"phone"`
	Message string `form:"message" json:"message"`
}
