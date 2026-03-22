package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func seedDBWithEvents(s *Server) error {
	tx, err := s.db.Beginx()
	if err != nil {
		return err
	}
	data, err := os.ReadFile("./data/events/events.json")
	if err != nil {
		return err
	}
	eventsSeeds := []EventSeed{}
	if err := json.Unmarshal(data, &eventsSeeds); err != nil {
		return err
	}

	if os.Getenv("CLEAR_WHILE_SEEDING") == "true" {
		tx.Exec(`TRUNCATE TABLE event, event_media RESTART IDENTITY CASCADE`)
	}

	insertEventQuery := `
		INSERT INTO event (title,start_date,end_date) VALUES ($1,$2,$3) ON CONFLICT DO NOTHING RETURNING id;
	`
	insertMediaQuery := `
	   INSERT INTO media (title, media_type, description, url, thumbnail_url,s3_key)
	   VALUES ($1, $2, $3, $4, $5, $6)
	   ON CONFLICT (url) DO UPDATE SET url = EXCLUDED.url
	   RETURNING id;
	`
	insertEventMedia := `
		INSERT INTO event_media VALUES ($1,$2) ON CONFLICT DO NOTHING;	
	`
	mediaFolderPath := "./data/events/media"
	for _, eventSeed := range eventsSeeds {
		currentEventMediaFolder := filepath.Join(mediaFolderPath, eventSeed.Title)
		files, err := os.ReadDir(currentEventMediaFolder)
		if err != nil {
			fmt.Println("NO MEDIA FOLDER")
			return err
		}
		for _, file := range files {
			eventSeed.Medias = append(eventSeed.Medias, Media{
				Title:       nil,
				MediaType:   "s3",
				Description: nil,
				Url:         file.Name(),
			})
		}
		var eventId string
		err = tx.QueryRow(insertEventQuery, eventSeed.Title, eventSeed.StartDate, eventSeed.EndDate).Scan(&eventId)
		if err != nil {
			return err
		}
		for _, media := range eventSeed.Medias {
			if media.MediaType == "s3" {
				location, s3Key, err := s.uploadTos3(filepath.Join(currentEventMediaFolder, media.Url), media.Url, "events")
				if err != nil {
					return err
				}
				media.Url = location
				media.S3Key = s3Key
			}
			var mediaId string
			err = tx.QueryRow(insertMediaQuery, media.Title, media.MediaType, media.Description, media.Url, media.ThumbnailUrl, media.S3Key).Scan(&mediaId)
			if err != nil {
				return err
			}
			tx.Exec(insertEventMedia, eventId, mediaId)
		}
	}
	return tx.Commit()
}

func seedDBWithActivities(s *Server) error {
	data, err := os.ReadFile("./data/activities/activities.json")
	if err != nil {
		return err
	}
	tx, err := s.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	activitySeeds := []ActivitySeed{}
	if err := json.Unmarshal(data, &activitySeeds); err != nil {
		fmt.Println("Error unmarshalling activities", err)
		return err
	}
	if os.Getenv("CLEAR_WHILE_SEEDING") == "true" {
		tx.Exec(`TRUNCATE TABLE activity, activity_media RESTART IDENTITY CASCADE`)
	}

	insertActivityQuery := `
		INSERT INTO activity (title,start_date,end_date) VALUES ($1,$2,$3) ON CONFLICT DO NOTHING RETURNING id;
	`

	insertMediaQuery := `
    INSERT INTO media (title, media_type, description, url, thumbnail_url,s3_key) 
    VALUES ($1, $2, $3, $4, $5,$6) 
    ON CONFLICT (url) DO UPDATE SET url = EXCLUDED.url 
    RETURNING id;
	`
	insertActivityMediaQuery := `
		INSERT INTO activity_media VALUES($1,$2)`

	basePath := "./data/activities"
	for _, activitySeed := range activitySeeds {
		var activityId string
		err := tx.QueryRow(insertActivityQuery, activitySeed.Title, activitySeed.StartDate, activitySeed.EndDate).Scan(&activityId)
		if err != nil {
			return err
		}
		for _, media := range activitySeed.Medias {
			if media.MediaType == "s3" {
				location, s3Key, err := s.uploadTos3(filepath.Join(basePath, "media", media.Url), media.Url, "activities")
				if err != nil {
					return err
				}
				media.Url = location
				media.S3Key = s3Key
			}
			var mediaId string
			err = tx.QueryRow(insertMediaQuery, media.Title, media.MediaType, media.Description, media.Url, media.ThumbnailUrl, media.S3Key).Scan(&mediaId)

			if err != nil {
				return err
			}
			_, err := tx.Exec(insertActivityMediaQuery, activityId, mediaId)
			if err != nil {
				return err
			}
		}

	}
	return tx.Commit()
}

func seedDBWithJournal(s *Server) error {
	tx, err := s.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if os.Getenv("CLEAR_WHILE_SEEDING") == "true" {
		tx.Exec(`TRUNCATE TABLE journal, journal_chapter, author RESTART IDENTITY CASCADE`)
	}
	basePath := "./data/journal"
	data, err := os.ReadFile(filepath.Join(basePath, "journals.json"))
	if err != nil {
		return err
	}

	journalSeeds := []JournalSeed{}
	if err := json.Unmarshal(data, &journalSeeds); err != nil {
		return err
	}
	insertJournal := `INSERT INTO journal (title,start_date,end_date,volume_number,issue_number) VALUES($1,$2,$3,$4,$5) ON CONFLICT DO NOTHING RETURNING id;`
	insertJournalChapter := `INSERT INTO journal_chapter(title,journal_id) VALUES ($1,$2) ON CONFLICT DO NOTHING RETURNING id;`
	insertAuthor := `INSERT INTO author (name) VALUES ($1)
	ON CONFLICT (name) DO 
	UPDATE SET name=EXCLUDED.name RETURNING id;`
	insertJournalChapterAuthor := `INSERT INTO journal_chapter_author VALUES ($1,$2) ON CONFLICT DO NOTHING;`
	insertJournalMedia := `INSERT INTO journal_media VALUES($1,$2) ON CONFLICT DO NOTHING;`
	insertMedia := `INSERT INTO media (media_type,url,s3_key) VALUES($1,$2,$3) ON CONFLICT (url) DO UPDATE SET url = EXCLUDED.url RETURNING id`
	insertJournalChapterMedia := `INSERT INTO journal_chapter_media VALUES($1,$2) ON CONFLICT DO NOTHING;`
	mediaFolder := filepath.Join(basePath, "media")
	for _, journalSeed := range journalSeeds {

		var journalId string
		err := tx.QueryRow(
			insertJournal,
			journalSeed.Title,
			journalSeed.StartDate,
			journalSeed.EndDate,
			journalSeed.VolumeNumber,
			journalSeed.IssueNumebr,
		).Scan(&journalId)

		if err != nil {
			return err
		}

		location, s3Key, err := s.uploadTos3(filepath.Join(mediaFolder, journalSeed.Title, "journal.pdf"), journalSeed.Title, "journal")

		var journalMediaId string

		err = tx.QueryRow(insertMedia, "s3", location, s3Key).Scan(&journalMediaId)
		tx.Exec(insertJournalMedia, journalId, journalMediaId)

		location, s3Key, err = s.uploadTos3(filepath.Join(mediaFolder, journalSeed.Title, "prelimenry.pdf"), journalSeed.Title+"_prelimenry", "journal")

		var journalMediaPreId string

		err = tx.QueryRow(insertMedia, "s3", location, s3Key).Scan(&journalMediaPreId)
		tx.Exec(insertJournalMedia, journalId, journalMediaPreId)

		for _, chapter := range journalSeed.Chapters {
			var chapterId string
			err := tx.QueryRow(insertJournalChapter, chapter.Name, journalId).Scan(&chapterId)
			if err != nil {
				return err
			}
			location, s3Key, err := s.uploadTos3(filepath.Join(mediaFolder, journalSeed.Title, fmt.Sprintf("%s.pdf", chapter.Name)), journalSeed.Title+chapter.Name, "journalchapter")
			if err != nil {
				return err
			}
			var journalChapMediaId string
			err = tx.QueryRow(insertMedia, "s3", location, s3Key).Scan(&journalChapMediaId)

			tx.Exec(insertJournalChapterMedia, chapterId, journalChapMediaId)

			for _, author := range chapter.Authors {
				var authorId string
				err := tx.QueryRow(insertAuthor, author).Scan(&authorId)
				if err != nil {
					return err
				}
				tx.Exec(insertJournalChapterAuthor, chapterId, authorId)
			}
		}

	}
	return tx.Commit()
}

func seedDBWithWorkshop(s *Server) error {
	tx, err := s.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if os.Getenv("CLEAR_WHILE_SEEDING") == "true" {
		tx.Exec(`TRUNCATE TABLE workshop, workshop_media RESTART IDENTITY CASCADE`)
	}
	basPath := "./data/workshop"
	data, err := os.ReadFile(filepath.Join(basPath, "workshops.json"))
	if err != nil {
		return err
	}
	workshopSeeds := []WorkshopSeed{}
	if err := json.Unmarshal(data, &workshopSeeds); err != nil {
		return err
	}
	insertWorkshop := `INSERT INTO workshop(title,start_date,end_date,description,workshop_type) VALUES ($1,$2,$3,$4,$5) RETURNING id`
	insertMedia := `INSERT INTO media (media_type,url,s3_key) VALUES($1,$2,$3) ON CONFLICT (url) DO UPDATE SET url = EXCLUDED.url RETURNING id`
	insertWorkshopMedia := `INSERT INTO workshop_media (workshop_id,media_id) VALUES ($1,$2) `
	for _, workshopSeed := range workshopSeeds {
		var workshopId string
		err := tx.QueryRow(
			insertWorkshop,
			workshopSeed.Title,
			workshopSeed.StartDate,
			workshopSeed.EndDate,
			workshopSeed.Description,
			workshopSeed.WorkshopType).Scan(&workshopId)
		if err != nil {
			return err
		}
		if workshopSeed.Link != nil {
			filePathParts := strings.Split(*workshopSeed.Link, "/")
			fileName := filePathParts[len(filePathParts)-1]
			location, s3Key, err := s.uploadTos3(filepath.Join(basPath, "media", fileName), fileName, "workshop")
			if err != nil {
				return err
			}
			var mediaId string
			err = tx.QueryRow(insertMedia, "s3", location, s3Key).Scan(&mediaId)
			if err != nil {
				return err
			}
			tx.Exec(insertWorkshopMedia, workshopId, mediaId)

		}
	}
	return tx.Commit()
}

func seedDB(s *Server) error {
	fmt.Println("[SEEDING STARTED]")
	// fmt.Println("\t[SEEDING ACTIVTIES]")
	// if err := seedDBWithActivities(s); err != nil {
	// 	return err
	// }
	// fmt.Println("\t[SEEDED ACTIVTIES]")
	fmt.Println("\t[SEEDING EVENTS]")
	if err := seedDBWithEvents(s); err != nil {
		return err
	}
	fmt.Println("\t[SEEDED EVENTS]")
	fmt.Println("\t[SEEDING JOURNAL]")
	if err := seedDBWithJournal(s); err != nil {
		return err
	}
	fmt.Println("\t[SEEDED JOURNAL]")
	fmt.Println("\t[SEEDING WORKSHOP]")
	if err := seedDBWithWorkshop(s); err != nil {
		return err
	}
	fmt.Println("\t[SEEDED WORKSHOP]")
	fmt.Println("[SEEDING DONE]")

	return nil
}
