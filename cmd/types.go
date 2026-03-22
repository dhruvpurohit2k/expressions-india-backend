package main

type Activity struct {
	Id        string  `json:"id" db:"id"`
	Title     string  `json:"title" db:"title"`
	StartDate string  `json:"start_date" db:"start_date"`
	EndDate   *string `json:"end_date" db:"end_date"`
}

type Enquiry struct {
	Id            string `json:"id" db:"id"`
	Name          string `json:"name" db:"name"`
	Designation   string `json:"designation" db:"designation"`
	EmailId       string `json:"emailId" db:"email_id"`
	ContactNumber string `json:"contactNumber" db:"contact_number"`
	Body          string `json:"body" db:"body"`
	CreatedAt     string `json:"createdAt" db:"created_at"`
	UpdatedAt     string `json:"updatedAt" db:"updated_at"`
}

type Event struct {
	Id        string  `json:"id" db:"id"`
	Title     string  `json:"title" db:"title"`
	StartDate string  `json:"start_date" db:"start_date"`
	EndDate   *string `json:"end_date" db:"end_date"`
}

type Workshop struct {
	Id           string  `json:"id" db:"id"`
	Name         string  `json:"name" db:"name"`
	Description  string  `json:"description" db:"description"`
	StartDate    string  `json:"start_date" db:"start_date"`
	EndDate      *string `json:"end_date" db:"end_date"`
	WorkshopType string  `json:"workshopType"`
}
