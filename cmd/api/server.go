package main

import (
	"fmt"
	"log"
	"os"

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
	)

	models.SeedAudience(db)
	models.SeedPromotions(db, s3)

	eventService := *event.NewService(db, s3)
	eventController := event.NewController(eventService)

	promotionService := promotion.NewService(db)
	promotionController := promotion.NewController(promotionService)

	journalsService := journal.NewService(db)
	journalsController := journal.NewController(journalsService)

	podcastService := podcast.NewService(db)
	podcastController := podcast.NewController(podcastService)

	if err != nil {
		log.Fatal(err.Error())
	}
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
		groupAdmin.GET("/journal", s.journalController.Get)
		groupAdmin.GET("/promotion", s.promotionController.Get)
		groupAdmin.GET("/podcast", s.podcastController.Get)

		// groupAdmin.POST("/event")
	}

}
