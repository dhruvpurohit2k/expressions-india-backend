package main

type ActivitySeed struct {
	Title     string  `json:"title" db:"title"`
	StartDate string  `json:"startDate" db:"start_date"`
	EndDate   *string `json:"endDate" db:"end_date"`
	Medias    []Media `json:"medias"`
}

type Media struct {
	Title        *string `json:"title" db:"title"`
	MediaType    string  `json:"mediaType" db:"media_type"`
	Description  *string `json:"description" db:"description"`
	Url          string  `json:"url" db:"url"`
	S3Key        string  `json:"s3Key" db:"s3_key"`
	ThumbnailUrl *string `json:"thumbnailUrl" db:"thumbnail_url"`
}

type EventSeed struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Perks       []string `json:"perks"`
	StartDate   string   `json:"start_date"`
	EndDate     *string  `json:"end_date"`
	StartTime   *string  `json:"start_time"`
	EndTime     *string  `json:"end_time"`
	Location    string   `json:"location"`
	IsPaid      bool     `json:"is_paid"`
	Price       *int     `json:"price"`
	Medias      []string `json:"medias"`
}

type JournalSeed struct {
	Title        string               `json:"title" db:"title"`
	StartDate    string               `json:"startDate" db:"start_date"`
	EndDate      string               `json:"endDate" db:"end_date"`
	VolumeNumber int                  `json:"volumeNumber" db:"volume_number"`
	IssueNumebr  int                  `json:"issueNumebr" db:"issue_number"`
	Chapters     []JournalChapterSeed `json:"chapters"`
}

type JournalChapterSeed struct {
	Name    string   `json:"name"`
	Authors []string `json:"authors"`
}

type WorkshopSeed struct {
	Title        string   `json:"title"`
	Description  string   `json:"description"`
	Perks        []string `json:"perks"`
	StartDate    string   `json:"startDate"`
	EndDate      *string  `json:"endDate"`
	StartTime    *string  `json:"startTime"`
	EndTime      *string  `json:"endTime"`
	Location     string   `json:"location"`
	IsPaid       *bool    `json:"is_paid"`
	Price        *int     `json:"price"`
	Medias       []string `json:"medias"`
	WorkshopType *int     `json:"type"`
	Link         *string  `json:"link"`
	Picture      *string  `json:"picture"`
}
