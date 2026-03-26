package main

type EventDTO struct {
	ID            string          `db:"id" json:"id"`
	Title         string          `db:"title" json:"title"`
	Description   string          `db:"description" json:"description"`
	Perks         []string        `json:"perks"`
	StartDate     string          `db:"start_date" json:"startDate"`
	EndDate       *string         `db:"end_date" json:"endDate"`
	StartTime     *string         `db:"start_time" json:"startTime"`
	EndTime       *string         `db:"end_time" json:"endTime"`
	Location      string          `db:"location" json:"location"`
	IsPaid        bool            `db:"is_paid" json:"isPaid"`
	Price         *int            `db:"price" json:"price"`
	UploadedMedia []UploadedMedia `json:"uploadedMedia"`
}
type WorkshopDTO struct {
	ID            string          `db:"id" json:"id"`
	Title         string          `db:"title" json:"title"`
	Description   string          `db:"description" json:"description"`
	Perks         []string        `json:"perks"`
	StartDate     string          `db:"start_date" json:"startDate"`
	EndDate       *string         `db:"end_date" json:"endDate"`
	StartTime     *string         `db:"start_time" json:"startTime"`
	EndTime       *string         `db:"end_time" json:"endTime"`
	Location      string          `db:"location" json:"location"`
	IsPaid        bool            `db:"is_paid" json:"isPaid"`
	Price         *int            `db:"price" json:"price"`
	WorkshopType  int             `db:"workshop_type" json:"workshopType"`
	UploadedMedia []UploadedMedia `json:"uploadedMedia"`
}

type EventListItem struct {
	Id        string  `json:"id" db:"id"`
	Title     string  `json:"title" db:"title"`
	StartDate string  `json:"startDate" db:"start_date"`
	EndDate   *string `json:"endDate" db:"end_date"`
}

type UploadedMedia struct {
	Id  string `json:"id" db:"id"`
	Url string `json:"url" db:"url"`
}
type WorkshopListDTO struct {
	Id           string  `json:"id" db:"id"`
	Title        string  `json:"title" db:"title"`
	StartDate    string  `json:"startDate" db:"start_date"`
	EndDate      *string `json:"endDate" db:"end_date"`
	WorkshopType string  `json:"workshopType" db:"workshop_type"`
}

// type WorkshopDTO struct {
// 	Id               string          `json:"id" db:"id"`
// 	Title            string          `json:"title" db:"title"`
// 	RegistrationLink *string         `json:"registrationLink" db:"registration_link"`
// 	Description      *string         `json:"description" db:"description"`
// 	StartDate        string          `json:"startDate" db:"start_date"`
// 	EndDate          *string         `json:"endDate" db:"end_date"`
// 	WorkshopType     string          `json:"workshopType" db:"workshop_type"`
// 	UploadedMedia    []UploadedMedia `json:"uploadedMedia" db:"-"`
// }
