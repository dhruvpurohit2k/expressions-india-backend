package main

import (
"encoding/json"
"fmt"
"os"
)

type Media struct {
	Title        *string `json:"title" db:"title"`
	MediaType    string  `json:"mediaType" db:"media_type"`
	Description  *string `json:"description" db:"description"`
	Url          string  `json:"url" db:"url"`
	ThumbnailUrl *string `json:"thumbnailUrl" db:"thumbnail_url"`
}

type ActivitySeed struct {
	Title     string  `json:"title" db:"title"`
	StartDate string  `json:"startDate" db:"start_date"`
	EndDate   *string `json:"endDate" db:"end_date"`
	Medias    []Media `json:"medias"`
}

func main() {
	data, err := os.ReadFile("./data/activities/activities.json")
	if err != nil {
		fmt.Println("Err read:", err)
		return
	}
	var s []ActivitySeed
	if err := json.Unmarshal(data, &s); err != nil {
		fmt.Println("Err unmarshal:", err)
		return
	}
	for _, a := range s {
		fmt.Printf("Activity: %s, Medias: %d\n", a.Title, len(a.Medias))
	}
}
