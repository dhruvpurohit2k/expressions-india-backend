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
	Subject string `form:"subject"`
	Name    string `form:"name"`
	Email   string `form:"email"`
	Phone   string `form:"phone"`
	Message string `form:"message"`
}
