package main

import (
	"encoding/json"
	"log"
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
	if err := s.db.Where("name = ?", "all").First(&allAudience).Error; err != nil {
		return err
	}
	status := "upcoming"
	for _, d := range eventSeeds {
		eventID, _ := uuid.NewV7()
		perksBlob, _ := json.Marshal(d.Perks)

		var promotionalMedia []models.Media
		var thumbnail *models.Media
		for i, fileName := range d.Medias {
			location, id, err := s.s3.UploadLocal(path.Join("./data/events/media", fileName))
			if err != nil {
				return err
			}
			media := models.Media{
				ID:       id,
				URL:      location,
				FileType: "image/png",
			}
			promotionalMedia = append(promotionalMedia, media)
			if i == 0 {
				m := media
				thumbnail = &m
			}
		}

		startDate, _ := time.Parse("2006-01-02", d.StartDate)
		var endDate *time.Time
		if d.EndDate != nil {
			endDateParsed, _ := time.Parse("2006-01-02", *d.EndDate)
			endDate = &endDateParsed
		}

		event := &models.Event{
			ID:               eventID.String(),
			Title:            d.Title,
			Description:      d.Description,
			Perks:            datatypes.JSON(perksBlob),
			Location:         &d.Location,
			IsPaid:           d.IsPaid,
			Price:            d.Price,
			StartDate:        startDate,
			Status:           &status,
			EndDate:          endDate,
			StartTime:        d.StartTime,
			EndTime:          d.EndTime,
			PromotionalMedia: promotionalMedia,
			Audiences:        []models.Audience{allAudience},
		}
		if thumbnail != nil {
			event.Thumbnail = thumbnail
			event.ThumbnailID = &thumbnail.ID
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
		mediaPath := path.Join("./data/journal/media/", seed.Title)
		wholePaper := &models.Media{
			ID:       uuid.Must(uuid.NewV7()).String(),
			FileType: "application/pdf",
		}
		location, _, err := s.s3.UploadLocal(path.Join(mediaPath, "journal.pdf"))
		if err != nil {
			log.Print(err.Error())
		}
		wholePaper.URL = location
		s.db.Create(wholePaper)

		startTime, err := time.Parse("2006-01-02", seed.StartDate)
		if err != nil {
			return err
		}
		endTime, err := time.Parse("2006-01-02", seed.EndDate)
		if err != nil {
			return err
		}
		chapters := make([]models.JournalChapter, len(seed.Chapters)+1)
		prefaceMedia := &models.Media{
			ID:       uuid.Must(uuid.NewV7()).String(),
			FileType: "application/pdf",
		}

		location, _, err = s.s3.UploadLocal(path.Join(mediaPath, "prelimenry.pdf"))
		if err != nil {
			log.Print(err.Error())
		}
		prefaceMedia.URL = location
		chapters[0] = models.JournalChapter{
			Title:   "Preface",
			Authors: nil,
			Media:   *prefaceMedia,
		}
		for i, chapter := range seed.Chapters {
			authors := make([]models.Author, len(chapter.Authors))
			for j, author := range chapter.Authors {
				s.db.Where(models.Author{Name: author}).FirstOrCreate(&authors[j], models.Author{Name: author})
			}
			media := &models.Media{
				ID:       uuid.Must(uuid.NewV7()).String(),
				FileType: "application/pdf",
			}
			location, _, err = s.s3.UploadLocal(path.Join(mediaPath, chapter.File))
			if err != nil {
				log.Print(err.Error())
			}
			media.URL = location
			chapters[i+1] = models.JournalChapter{
				Title:   chapter.Name,
				Authors: authors,
				Media:   *media,
			}
		}
		journal := models.Journal{
			Title:      seed.Title,
			StartMonth: startTime.Month().String(),
			EndMonth:   endTime.Month().String(),
			Year:       startTime.Year(),
			Volume:     seed.Volume,
			Media:      *wholePaper,
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
