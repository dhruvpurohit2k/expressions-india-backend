package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"time"

	"github.com/dhruvpurohit2k/expressions-india-backend/internal/models"
	"github.com/google/uuid"
	"gorm.io/datatypes"
)

func SeedDBWithEvent(s *Server, filepath string) error {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return err
	}
	var eventSeeds []struct {
		ID          string   `json:"id"`
		Title       string   `json:"title"`
		Description string   `json:"description"`
		Perks       []string `json:"perks"`
		StartDate   string   `json:"startDate"`
		EndDate     *string  `json:"endDate"`
		StartTime   *string  `json:"startTime"`
		EndTime     *string  `json:"endTime"`
		Location    string   `json:"location"`
		IsPaid      bool     `json:"isPaid"`
		Price       *int     `json:"price"`
		Medias      []string `json:"medias"`
	}
	err = json.Unmarshal(data, &eventSeeds)
	if err != nil {
		return err
	}
	var allAudience models.Audience
	if err := s.db.Where("name = ?", "All").First(&allAudience).Error; err != nil {
		return err
	}
	for _, d := range eventSeeds {
		eventID, _ := uuid.NewV7()
		perksBlob, _ := json.Marshal(d.Perks)
		var promotionalMedia []models.Media
		for _, fileName := range d.Medias {

			location, id, err := s.s3.UploadLocal(path.Join("./data/events/media", fileName))
			if err != nil {
				return err
			}
			promotionalMedia = append(promotionalMedia, models.Media{
				ID:       id,
				EventID:  eventID.String(),
				URL:      location,
				Key:      id,
				FileType: "image/png",
				Category: "PROMOTION",
			})
		}
		event := &models.Event{
			ID:               eventID.String(),
			Title:            d.Title,
			Description:      d.Description,
			Perks:            datatypes.JSON(perksBlob),
			Location:         &d.Location,
			IsPaid:           d.IsPaid,
			Price:            d.Price,
			PromotionalMedia: promotionalMedia,
			Audiences:        []models.Audience{allAudience},
		}
		s.db.Create(event)
	}
	return nil
}

func SeedJournal(s *Server, filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}
	var journalSeeds []JournalSeed
	if err := json.Unmarshal(data, &journalSeeds); err != nil {
		return err
	}

	for _, seed := range journalSeeds {
		startTime, err := time.Parse("2006-01-02", seed.StartDate)
		if err != nil {
			return err
		}
		endTime, err := time.Parse("2006-01-02", seed.EndDate)
		if err != nil {
			return err
		}
		chapters := make([]models.JournalChapter, len(seed.Chapters))
		for i, chapter := range seed.Chapters {
			authors := make([]models.Author, len(chapter.Authors))
			for j, author := range chapter.Authors {
				s.db.Where(models.Author{Name: author}).FirstOrCreate(&authors[j], models.Author{Name: author})
			}
			chapters[i] = models.JournalChapter{
				Title:   chapter.Name,
				Authors: authors,
			}
		}
		fmt.Println(chapters)
		journal := models.Journal{
			Title:      seed.Title,
			StartMonth: startTime.Month().String(),
			EndMonth:   endTime.Month().String(),
			Year:       startTime.Year(),
			Volume:     seed.Volume,
			Issue:      seed.Issue,
			Chapters:   chapters,
		}
		s.db.Create(&journal)
	}

	return nil
}

type JournalSeed struct {
	Title     string        `json:"title"`
	StartDate string        `json:"startDate"`
	EndDate   string        `json:"endDate"`
	Volume    int           `json:"volumeNumber"`
	Issue     int           `json:"issueNumber"`
	Chapters  []ChapterSeed `json:"chapters"`
}

type ChapterSeed struct {
	Name    string   `json:"name"`
	Authors []string `json:"authors"`
	File    string   `json:"file"`
}
