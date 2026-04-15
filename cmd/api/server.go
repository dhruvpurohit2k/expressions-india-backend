package main

import (
	"fmt"
	"log"
	"os"

	"strings"
	"time"

	"github.com/dhruvpurohit2k/expressions-india-backend/internal/article"
	"github.com/dhruvpurohit2k/expressions-india-backend/internal/audience"
	"github.com/dhruvpurohit2k/expressions-india-backend/internal/auth"
	"github.com/dhruvpurohit2k/expressions-india-backend/internal/course"
	"github.com/dhruvpurohit2k/expressions-india-backend/internal/enquiry"
	"github.com/dhruvpurohit2k/expressions-india-backend/internal/event"
	"github.com/dhruvpurohit2k/expressions-india-backend/internal/journal"
	latestactivity "github.com/dhruvpurohit2k/expressions-india-backend/internal/latest-activity"
	"github.com/dhruvpurohit2k/expressions-india-backend/internal/models"
	"github.com/dhruvpurohit2k/expressions-india-backend/internal/pkg/utils"
	"github.com/dhruvpurohit2k/expressions-india-backend/internal/podcast"
	"github.com/dhruvpurohit2k/expressions-india-backend/internal/promotion"
	"github.com/dhruvpurohit2k/expressions-india-backend/internal/storage"
	"github.com/dhruvpurohit2k/expressions-india-backend/internal/team"
	"github.com/dhruvpurohit2k/expressions-india-backend/internal/upload"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Server struct {
	r                        *gin.Engine
	db                       *gorm.DB
	s3                       *storage.S3
	eventController          *event.Controller
	promotionController      *promotion.Controller
	journalController        *journal.Controller
	podcastController        *podcast.Controller
	enquiryController        *enquiry.Controller
	articleController        *article.Controller
	audienceController       *audience.Controller
	latestActivityController *latestactivity.Controller
	courseController         *course.Controller
	authController           *auth.Controller
	teamController           *team.Controller
	uploadController         *upload.Controller
}

func initServer() *Server {
	fmt.Println("INITDB STARTING")
	db := storage.InitDB()
	s3 := storage.InitS3()

	if os.Getenv("DB_FRESH_START") == "true" {
		db.Migrator().DropTable(
			&models.User{},
			&models.Event{},
			&models.Media{},
			&models.Promotion{},
			&models.Link{},
			&models.Journal{},
			&models.JournalChapter{},
			&models.Author{},
			&models.Podcast{},
			&models.Enquiry{},
			&models.Article{},
			&models.Course{},
			&models.CourseChapter{},
			&models.Team{},
			&models.Member{},
		)
	}
	db.Migrator().DropTable(&models.Audience{})
	err := db.AutoMigrate(
		&models.User{},
		&models.Event{},
		&models.Media{},
		&models.Audience{},
		&models.Promotion{},
		&models.Link{},
		&models.Journal{},
		&models.JournalChapter{},
		&models.Author{},
		&models.Podcast{},
		&models.Enquiry{},
		&models.Article{},
		&models.Course{},
		&models.CourseChapter{},
		&models.Team{},
		&models.Member{},
	)
	if err != nil {
		log.Fatal(err.Error())
	}

	models.SeedAudience(db)
	// models.SeedPromotions(db, s3)

	eventService := event.NewService(db, s3)
	eventController := event.NewController(eventService)

	promotionService := promotion.NewService(db)
	promotionController := promotion.NewController(promotionService)

	journalsService := journal.NewService(db, s3)
	journalsController := journal.NewController(journalsService)

	podcastService := podcast.NewService(db)
	podcastController := podcast.NewController(podcastService)

	enquiryService := enquiry.NewService(db)
	enquiryController := enquiry.NewController(enquiryService)

	articleService := article.NewService(db, s3)
	articleController := article.NewController(articleService)

	audienceService := audience.NewService(db)
	audienceController := audience.NewController(audienceService)

	latestActivityService := latestactivity.NewService(db)
	latestActivityController := latestactivity.NewController(latestActivityService)

	courseService := course.NewService(db, s3)
	courseController := course.NewController(courseService)

	authService := auth.NewService(db)
	authController := auth.NewController(authService)

	teamService := team.NewService(db)
	teamController := team.NewController(teamService)

	uploadController := upload.NewController(s3)

	r := gin.Default()
	corsConfig := cors.Config{
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		MaxAge:           12 * time.Hour,
		AllowCredentials: true,
	}
	if allowedOrigins := os.Getenv("ALLOWED_ORIGINS"); allowedOrigins != "" {
		corsConfig.AllowOrigins = strings.Split(allowedOrigins, ",")
	} else {
		// Dev fallback: allow common local dev servers.
		corsConfig.AllowOrigins = []string{"http://localhost:5173", "http://localhost:3000"}
	}
	r.Use(cors.New(corsConfig))
	return &Server{
		r:                        r,
		db:                       db,
		s3:                       s3,
		eventController:          eventController,
		promotionController:      promotionController,
		journalController:        journalsController,
		podcastController:        podcastController,
		enquiryController:        enquiryController,
		articleController:        articleController,
		audienceController:       audienceController,
		latestActivityController: latestActivityController,
		courseController:         courseController,
		authController:           authController,
		teamController:           teamController,
		uploadController:         uploadController,
	}
}

func (s *Server) SetupRoutes() {
	s.r.GET("/hello", func(ctx *gin.Context) {
		utils.OK(ctx, gin.H{
			"message": "Success",
		})
	})

	groupAuth := s.r.Group("/auth")
	{
		groupAuth.POST("/login", s.authController.Login)
		groupAuth.POST("/refresh", s.authController.Refresh)
		groupAuth.POST("/logout", s.authController.Logout)
		// Signup is public — anyone can create a non-admin account.
		groupAuth.POST("/signup", s.authController.Signup)
		// Register is admin-only: only an existing admin can create new users.
		groupAuth.POST("/register", auth.RequireAdmin(), s.authController.Register)
	}

	groupAdmin := s.r.Group("/admin", auth.RequireAdmin())
	{
		groupAdmin.POST("/upload/presign", s.uploadController.Presign)

		groupAdmin.GET("/allEvents", s.eventController.GetAll)
		groupAdmin.GET("/event", s.eventController.GetEventList)
		groupAdmin.GET("/event/:id", s.eventController.GetEventById)
		groupAdmin.POST("/event", s.eventController.Create)
		groupAdmin.PUT("/event/:id", s.eventController.Update)
		groupAdmin.DELETE("/event/:id", s.eventController.Delete)

		groupAdmin.GET("/journal", s.journalController.GetList)
		groupAdmin.GET("/journal/:id", s.journalController.GetById)
		groupAdmin.DELETE("/journal/:id", s.journalController.Delete)
		groupAdmin.GET("/promotion", s.promotionController.Get)
		groupAdmin.GET("/promotion/:id", s.promotionController.GetById)

		groupAdmin.GET("/podcast", s.podcastController.GetPodcastList)
		groupAdmin.GET("/podcast/:id", s.podcastController.GetById)
		groupAdmin.POST("/podcast", s.podcastController.Create)
		groupAdmin.PUT("/podcast/:id", s.podcastController.Update)
		groupAdmin.DELETE("/podcast/:id", s.podcastController.Delete)

		groupAdmin.GET("/enquiry", s.enquiryController.GetList)
		groupAdmin.GET("/enquiry/:id", s.enquiryController.GetById)
		groupAdmin.DELETE("/enquiry/:id", s.enquiryController.Delete)

		groupAdmin.GET("/audience", s.audienceController.GetAudience)
		groupAdmin.PUT("/audience/:id", s.audienceController.UpdateDescription)
		// groupAdmin.GET("/audience/:id", s.audienceController.GetById)
		// groupAdmin.POST("/audience", s.audienceController.Create)
		// groupAdmin.DELETE("/audience/:id", s.audienceController.Delete)

		groupAdmin.GET("/article", s.articleController.GetArticleList)
		groupAdmin.GET("/article/:id", s.articleController.GetArticleById)
		groupAdmin.POST("/article", s.articleController.Create)
		groupAdmin.PUT("/article/:id", s.articleController.Update)
		groupAdmin.DELETE("/article/:id", s.articleController.Delete)

		groupAdmin.GET("/course", s.courseController.GetCoursesListAdmin)
		groupAdmin.GET("/course/:id", s.courseController.GetCourseById)
		groupAdmin.POST("/course", s.courseController.Create)
		groupAdmin.PUT("/course/:id", s.courseController.Update)
		groupAdmin.DELETE("/course/:id", s.courseController.Delete)
		groupAdmin.GET("/course/:id/enrolled", s.courseController.GetEnrolledUsers)
		groupAdmin.DELETE("/course/:id/enrolled/:userId", s.courseController.RevokeAccess)
		groupAdmin.GET("/course/:id/not-enrolled", s.courseController.GetNotEnrolledUsers)
		groupAdmin.POST("/course/:id/enroll", s.courseController.EnrollUser)

		groupAdmin.GET("/team", s.teamController.GetList)
		groupAdmin.GET("/team/:id", s.teamController.GetById)
		groupAdmin.POST("/team", s.teamController.Create)
		groupAdmin.PUT("/team/:id", s.teamController.Update)
		groupAdmin.DELETE("/team/:id", s.teamController.Delete)
	}
	groupApi := s.r.Group("/api")
	{
		groupApi.GET("/home/images", s.eventController.GetHomePageImages)
		groupApi.GET("/home/upcoming-images", s.eventController.GetUpcomingCarouselImages)
		groupApi.GET("/home/completed-images", s.eventController.GetCompletedCarouselImages)
		groupApi.GET("/event/upcoming", s.eventController.GetUpcomingEvents)
		groupApi.GET("/event/past", s.eventController.GetPastEvents)
		groupApi.GET("/event/:id", s.eventController.GetEventById)
		groupApi.GET("/podcast", s.podcastController.GetPodcastList)
		groupApi.GET("/podcast/:id", s.podcastController.GetById)
		groupApi.GET("/journal", s.journalController.GetList)
		groupApi.GET("/journal/:id", s.journalController.GetById)
		groupApi.POST("/enquiry", s.enquiryController.CreateEnquiry)
		groupApi.GET("/article", s.articleController.GetArticleListPaginated)
		groupApi.GET("/article/audience/:audience", s.articleController.GetArticlesByAudience)
		groupApi.GET("/article/:id", s.articleController.GetArticleById)
		groupApi.GET("/podcast/audience/:audience", s.podcastController.GetPodcastsByAudience)
		groupApi.GET("/event/audience/:audience", s.eventController.GetUpcomingEventsByAudience)
		groupApi.GET("/audience/:name", s.audienceController.GetAudienceByName)
		groupApi.GET("/latest-activity", s.latestActivityController.GetLatestActivity)
		groupApi.GET("/course", s.courseController.GetCoursesList)
		groupApi.GET("/course/audience/:audience", s.courseController.GetCoursesByAudience)
		groupApi.GET("/course/:id", s.courseController.GetCourseById)
		groupApi.GET("/course/:id/chapter/:chapterId", auth.TryExtractClaims(), s.courseController.GetChapterById)
		groupApi.GET("/team", s.teamController.GetList)
		groupApi.GET("/team/:id", s.teamController.GetById)
	}

}
