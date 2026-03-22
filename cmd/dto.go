package main

type EventDTO struct {
	Id               string   `json:"id" db:"id"`
	Title            string   `json:"title" db:"title"`
	Description      string   `json:"description" db:"description"`
	StartDate        string   `json:"startDate" db:"start_date"`
	EndDate          *string  `json:"endDate" db:"end_date"`
	MediaLink        []string `json:"mediaLink" db:"-"`
	RegistrationLink *string  `json:"registrationLink" db:"registration_link"`
}

type EventListItem struct {
	Id        string  `json:"id" db:"id"`
	Title     string  `json:"title" db:"title"`
	StartDate string  `json:"startDate" db:"start_date"`
	EndDate   *string `json:"endDate" db:"end_date"`
}
