package main

import (
	"encoding/json"
	"os"
	"path"

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
	for _, d := range eventSeeds {
		eventID, _ := uuid.NewV7()

		perksBlob, _ := json.Marshal(d.Perks)

		var mediaModels []models.Media
		for _, fileName := range d.Medias {

			location, id, err := s.s3.UploadLocal(path.Join("./data/events/media", fileName))
			if err != nil {
				return err
			}
			mediaModels = append(mediaModels, models.Media{
				ID:       id,
				EventID:  eventID.String(),
				URL:      location,
				Key:      id,
				FileType: "image/png",
			})
		}
		event := &models.Event{
			ID:          eventID.String(),
			Title:       d.Title,
			Description: d.Description,
			Perks:       datatypes.JSON(perksBlob),
			Location:    d.Location,
			IsPaid:      d.IsPaid,
			Price:       d.Price,
			Medias:      mediaModels, // GORM handles the batch insert
		}
		s.db.Create(event)
	}
	return nil
}
