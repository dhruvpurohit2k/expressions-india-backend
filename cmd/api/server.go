package main

import (
	"fmt"
	"log"
	"os"

	"github.com/dhruvpurohit2k/expressions-india-backend/internal/article"
	"github.com/dhruvpurohit2k/expressions-india-backend/internal/audience"
	"github.com/dhruvpurohit2k/expressions-india-backend/internal/enquiry"
	"github.com/dhruvpurohit2k/expressions-india-backend/internal/event"
	"github.com/dhruvpurohit2k/expressions-india-backend/internal/journal"
	"github.com/dhruvpurohit2k/expressions-india-backend/internal/models"
	"github.com/dhruvpurohit2k/expressions-india-backend/internal/pkg/utils"
	"github.com/dhruvpurohit2k/expressions-india-backend/internal/podcast"
	"github.com/dhruvpurohit2k/expressions-india-backend/internal/promotion"
	"github.com/dhruvpurohit2k/expressions-india-backend/internal/storage"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Server struct {
	r                   *gin.Engine
	db                  *gorm.DB
	s3                  *storage.S3
	eventController     *event.Controller
	promotionController *promotion.Controller
	journalController   *journal.Controller
	podcastController   *podcast.Controller
	enquiryController   *enquiry.Controller
	articleController   *article.Controller
	audienceController  *audience.Controller
}

func initServer() *Server {
	fmt.Println("INITDB STARTING")
	db := storage.InitDB()
	s3 := storage.InitS3()

	if os.Getenv("DB_FRESH_START") == "true" {
		db.Migrator().DropTable(
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
		)
	}
	err := db.AutoMigrate(
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
	)
	if err != nil {
		log.Fatal(err.Error())
	}

	models.SeedAudience(db)
	// models.SeedPromotions(db, s3)

	eventService := *event.NewService(db, s3)
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

	r := gin.Default()
	r.Use(cors.Default())
	return &Server{
		r:                   r,
		db:                  db,
		s3:                  s3,
		eventController:     eventController,
		promotionController: promotionController,
		journalController:   journalsController,
		podcastController:   podcastController,
		enquiryController:   enquiryController,
		articleController:   articleController,
		audienceController:  audienceController,
	}
}

func (s *Server) SetupRoutes() {
	s.r.GET("/hello", func(ctx *gin.Context) {
		utils.OK(ctx, gin.H{
			"message": "Success",
		})
	})
	groupAdmin := s.r.Group("/admin")
	{
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

		groupAdmin.GET("/podcast", s.podcastController.Get)
		groupAdmin.GET("/podcast/:id", s.podcastController.GetById)
		groupAdmin.POST("/podcast", s.podcastController.Create)
		groupAdmin.DELETE("/podcast/:id", s.podcastController.Delete)

		groupAdmin.GET("/enquiry", s.enquiryController.Get)
		groupAdmin.GET("/enquiry/:id", s.enquiryController.GetById)
		groupAdmin.DELETE("/enquiry/:id", s.enquiryController.Delete)

		groupAdmin.GET("/audience", s.audienceController.GetAudience)
		// groupAdmin.GET("/audience/:id", s.audienceController.GetById)
		// groupAdmin.POST("/audience", s.audienceController.Create)
		// groupAdmin.DELETE("/audience/:id", s.audienceController.Delete)

		groupAdmin.GET("/article", s.articleController.GetArticleList)
		groupAdmin.GET("/article/:id", s.articleController.GetArticleById)
		groupAdmin.POST("/article", s.articleController.Create)
		groupAdmin.PUT("/article/:id", s.articleController.Update)
		groupAdmin.DELETE("/article/:id", s.articleController.Delete)
	}
	groupApi := s.r.Group("/api")
	{
		groupApi.GET("/event/upcoming", s.eventController.GetUpcomingEvents)
		groupApi.GET("/event/past", s.eventController.GetPastEvents)
		groupApi.GET("/event/:id", s.eventController.GetEventById)
		groupApi.GET("/podcast", s.podcastController.GetPodcastList)
		groupApi.GET("/podcast/:id", s.podcastController.GetById)
		groupApi.GET("/journal", s.journalController.GetJournalList)
		groupApi.GET("/journal/:id", s.journalController.GetById)
		groupApi.POST("/enquiry", s.enquiryController.CreateEnquiry)
		groupApi.GET("/article", s.articleController.GetArticleListPaginated)
		groupApi.GET("/article/audience/:audience", s.articleController.GetArticlesByAudience)
		groupApi.GET("/article/:id", s.articleController.GetArticleById)
		groupApi.GET("/podcast/audience/:audience", s.podcastController.GetPodcastsByAudience)
		groupApi.GET("/event/audience/:audience", s.eventController.GetUpcomingEventsByAudience)
		groupApi.GET("/audience/:name", s.audienceController.GetAudienceByName)
	}

}
