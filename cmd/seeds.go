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
	fmt.Println(eventsSeeds)
	s.db.Exec(`TRUNCATE TABLE event CASCADE;`)
	insertEventQuery := `
	INSERT INTO event 
	(title, description, perks, start_date, end_date, start_time, end_time, location, is_paid, price) 
	VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
	ON CONFLICT DO NOTHING
	RETURNING id;
	`
	insertMediaQuery := `
	   INSERT INTO media (media_type,url,s3_key)
	   VALUES ($1, $2, $3)
	   ON CONFLICT (url) DO UPDATE SET url = EXCLUDED.url
	   RETURNING id;
	`
	insertEventMedia := `
		INSERT INTO event_media VALUES ($1,$2) ON CONFLICT DO NOTHING;	
	`
	mediaFolderPath := "./data/events/media"
	for _, eventSeed := range eventsSeeds {

		perksJSON, err := json.Marshal(eventSeed.Perks)
		if err != nil {
			return err
		}
		var eventId string
		err = tx.QueryRow(
			insertEventQuery,
			eventSeed.Title,
			eventSeed.Description,
			perksJSON,
			eventSeed.StartDate,
			eventSeed.EndDate,
			eventSeed.StartTime,
			eventSeed.EndTime,
			eventSeed.Location,
			eventSeed.IsPaid,
			eventSeed.Price,
		).Scan(&eventId)

		if err != nil {
			return err
		}

		for _, media := range eventSeed.Medias {
			fmt.Println(media)

			location, s3Key, err := s.uploadTos3(
				filepath.Join(mediaFolderPath, media),
				media,
				"events",
			)
			if err != nil {
				return err
			}
			var mediaId string
			err = tx.QueryRow(
				insertMediaQuery,
				"image",
				location,
				s3Key,
			).Scan(&mediaId)

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
	insertWorkshop := `
	INSERT INTO workshop 
	(title, description, perks, start_date, end_date, start_time, end_time, location, is_paid, price,workshop_type) 
	VALUES ($1,$2,$3,$4,NULLIF($5,'')::DATE,$6,$7,$8,$9,$10,$11::INT)
	ON CONFLICT DO NOTHING
	RETURNING id;
	`
	insertMedia := `INSERT INTO media (media_type,url,s3_key) VALUES($1,$2,$3) ON CONFLICT (url) DO UPDATE SET url = EXCLUDED.url RETURNING id`
	insertWorkshopMedia := `INSERT INTO workshop_media (workshop_id,media_id) VALUES ($1,$2) `
	for _, workshopSeed := range workshopSeeds {
		fmt.Println(workshopSeed)
		var workshopId string
		perksJSON, err := json.Marshal(workshopSeed.Perks)
		if err != nil {
			return err
		}
		err = tx.QueryRow(
			insertWorkshop,
			workshopSeed.Title,
			workshopSeed.Description,
			perksJSON,
			workshopSeed.StartDate,
			workshopSeed.EndDate,
			workshopSeed.StartTime,
			workshopSeed.EndTime,
			workshopSeed.Location,
			workshopSeed.IsPaid,
			workshopSeed.Price,
			workshopSeed.WorkshopType,
		).Scan(&workshopId)
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
		if workshopSeed.Picture != nil {
			location, s3Key, err := s.uploadTos3(filepath.Join(basPath, "media", *workshopSeed.Picture), *workshopSeed.Picture, "workshop")
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

func addHomePageImages(s *Server) error {
	insertImage := `INSERT INTO home_page_image (url,s3_key) VALUES ($1,$2);`
	defaultImages := []string{
		"./data/events/media/1.png",
		"./data/events/media/2.png",
		"./data/events/media/3.png",
	}
	s.db.Exec(`TRUNCATE TABLE home_page_image CASCADE;`)
	tx, err := s.db.Beginx()
	if err != nil {
		fmt.Println(err)
		return err
	}

	defer tx.Rollback()
	for _, image := range defaultImages {

		location, s3Key, err := s.uploadTos3(image, image, "homepage")
		if err != nil {
			fmt.Println(err)
			return err
		}
		_, err = tx.Exec(insertImage, location, s3Key)
		if err != nil {
			return err
		}
		fmt.Println("ADDED", image)

	}
	return tx.Commit()
}

func seedDB(s *Server) error {
	fmt.Println("[SEEDING STARTED]")

	fmt.Println("\t[SEEDING EVENTS]")
	if err := seedDBWithEvents(s); err != nil {
		return err
	}
	fmt.Println("\t[SEEDED EVENTS]")
	// fmt.Println("\t[SEEDING JOURNAL]")
	// if err := seedDBWithJournal(s); err != nil {
	// 	return err
	// }
	// fmt.Println("\t[SEEDED JOURNAL]")
	fmt.Println("\t[SEEDING WORKSHOP]")
	if err := seedDBWithWorkshop(s); err != nil {
		return err
	}
	fmt.Println("\t[SEEDED WORKSHOP]")
	fmt.Println("\t[SEEDING HOMEPAGEIMAGE]")
	if err := addHomePageImages(s); err != nil {
		return err
	}
	fmt.Println("\t[SEEDED HOMEPAGEIMAGE]")
	fmt.Println("[SEEDING DONE]")

	return nil
}
