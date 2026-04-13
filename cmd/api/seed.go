package main

import (
	"encoding/json"
	"log"
	"os"
	"path"
	"time"

	"github.com/dhruvpurohit2k/expressions-india-backend/internal/models"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
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
		Status      *string  `json:"status"`
	}
	err = json.Unmarshal(data, &eventSeeds)
	if err != nil {
		return err
	}
	var allAudience models.Audience
	if err := s.db.Where("name = ?", "all").First(&allAudience).Error; err != nil {
		return err
	}
	defaultStatus := "upcoming"
	for _, d := range eventSeeds {
		status := defaultStatus
		if d.Status != nil {
			status = *d.Status
		}
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

// SeedAdminUser creates the first admin user from ADMIN_EMAIL and ADMIN_PASSWORD
// env vars. It is a no-op if those vars are unset or if an admin with that email
// already exists.
func SeedAdminUser(s *Server) error {
	email := os.Getenv("ADMIN_EMAIL")
	password := os.Getenv("ADMIN_PASSWORD")
	if email == "" || password == "" {
		log.Println("ADMIN_EMAIL or ADMIN_PASSWORD not set; skipping admin seed")
		return nil
	}
	var count int64
	s.db.Model(&models.User{}).Where("email = ?", email).Count(&count)
	if count > 0 {
		return nil
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user := models.User{
		Email:    email,
		Password: string(hash),
		IsAdmin:  true,
	}
	return s.db.Create(&user).Error
}

func SeedPodcasts(s *Server, filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}
	var podcastSeeds []struct {
		Title       string  `json:"title"`
		Link        string  `json:"link"`
		Description *string `json:"description"`
		Tags        string  `json:"tags"`
		Transcript  *string `json:"transcript"`
	}
	err = json.Unmarshal(data, &podcastSeeds)
	if err != nil {
		return err
	}

	var allAudience models.Audience
	if err := s.db.Where("name = ?", "all").First(&allAudience).Error; err != nil {
		return err
	}

	for _, d := range podcastSeeds {
		podcastID, _ := uuid.NewV7()
		var tagsBlob datatypes.JSON
		if d.Tags != "" {
			tagsBlob = datatypes.JSON([]byte(d.Tags))
		}

		podcast := &models.Podcast{
			ID:          podcastID.String(),
			Title:       d.Title,
			Link:        d.Link,
			Description: d.Description,
			Tags:        tagsBlob,
			Transcript:  d.Transcript,
			Audiences:   []models.Audience{allAudience},
		}
		s.db.Create(podcast)
	}
	return nil
}

func SeedArticles(s *Server, filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}
	var articleSeeds []struct {
		Title    string   `json:"title"`
		Content  string   `json:"content"`
		Category string   `json:"category"`
		Medias   []string `json:"medias"`
	}
	err = json.Unmarshal(data, &articleSeeds)
	if err != nil {
		return err
	}

	var allAudience models.Audience
	if err := s.db.Where("name = ?", "all").First(&allAudience).Error; err != nil {
		return err
	}

	for _, d := range articleSeeds {
		articleID, _ := uuid.NewV7()

		var articleMedias []models.Media
		for _, fileName := range d.Medias {
			location, id, err := s.s3.UploadLocal(path.Join("./data/articles/media", fileName))
			if err != nil {
				// Try to find the file in events/media as fallback
				location, id, err = s.s3.UploadLocal(path.Join("./data/events/media", fileName))
				if err != nil {
					log.Print(err.Error())
					continue
				}
			}
			mediaType := "image/png"
			if fileName == "demo_document.pdf" {
				mediaType = "application/pdf"
			}
			media := models.Media{
				ID:       id,
				URL:      location,
				FileType: mediaType,
			}
			articleMedias = append(articleMedias, media)
		}

		article := &models.Article{
			ID:       articleID.String(),
			Title:    d.Title,
			Content:  d.Content,
			Category: d.Category,
			Audience: []models.Audience{allAudience},
			Medias:   articleMedias,
		}
		s.db.Create(article)
	}
	return nil
}

func SeedCourses(s *Server, filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}
	var courseSeeds []struct {
		Title                string `json:"title"`
		Description          string `json:"description"`
		IntroductionVideoUrl string `json:"introductionVideoUrl"`
		Thumbnail            string `json:"thumbnail"`
	}
	err = json.Unmarshal(data, &courseSeeds)
	if err != nil {
		return err
	}

	var allAudience models.Audience
	if err := s.db.Where("name = ?", "all").First(&allAudience).Error; err != nil {
		return err
	}

	for _, d := range courseSeeds {
		courseID, _ := uuid.NewV7()

		// Create thumbnail
		var thumbnail *models.Media
		if d.Thumbnail != "" {
			location, id, err := s.s3.UploadLocal(path.Join("./data/events/media", d.Thumbnail))
			if err != nil {
				log.Print(err.Error())
			} else {
				thumbnail = &models.Media{
					ID:       id,
					URL:      location,
					FileType: "image/png",
				}
			}
		}

		// Create intro video link if provided
		var introVideoLink *models.Link
		if d.IntroductionVideoUrl != "" {
			introVideoLink = &models.Link{
				ID:  uuid.Must(uuid.NewV7()).String(),
				URL: d.IntroductionVideoUrl,
			}
			s.db.Create(introVideoLink)
		}

		course := &models.Course{
			ID:          courseID.String(),
			Title:       d.Title,
			Description: d.Description,
			Audiences:   []models.Audience{allAudience},
		}
		if thumbnail != nil {
			course.Thumbnail = *thumbnail
			course.ThumbnailID = thumbnail.ID
		}
		if introVideoLink != nil {
			course.IntroductionVideo = *introVideoLink
			course.IntroductionVideoID = introVideoLink.ID
		}
		s.db.Create(course)

		// Create 10 chapters; the first 3 are free previews.
		type chapterDef struct {
			title       string
			description string
			isFree      bool
		}
		chapterDefs := []chapterDef{
			{"Introduction to Life Skills", "An overview of essential life skills and their importance in everyday student life.", true},
			{"Self-Awareness and Identity", "Understanding your values, strengths, and areas for growth to build a strong personal foundation.", true},
			{"Emotional Intelligence", "Recognising, understanding, and managing emotions to improve relationships and decision-making.", true},
			{"Communication Skills", "Developing effective verbal and non-verbal communication techniques for academic and social settings.", false},
			{"Critical Thinking and Problem Solving", "Frameworks for analysing situations, evaluating options, and making sound, confident decisions.", false},
			{"Stress Management", "Identifying personal stressors and building healthy, sustainable coping strategies.", false},
			{"Healthy Relationships", "Building and maintaining positive relationships in personal, academic, and professional life.", false},
			{"Goal Setting and Time Management", "Setting meaningful short- and long-term goals and managing time to achieve them effectively.", false},
			{"Financial Wellbeing", "Fundamentals of budgeting, saving, and responsible financial planning tailored for students.", false},
			{"Resilience and Growth Mindset", "Cultivating resilience to bounce back from setbacks and embrace challenges as opportunities for growth.", false},
		}
		for _, cd := range chapterDefs {
			chapter := models.CourseChapter{
				CourseID:    course.ID,
				Title:       cd.title,
				Description: cd.description,
				IsFree:      cd.isFree,
			}
			if err := s.db.Create(&chapter).Error; err != nil {
				log.Printf("failed to seed chapter %q for course %q: %v", cd.title, course.Title, err)
			}
		}
	}
	return nil
}
