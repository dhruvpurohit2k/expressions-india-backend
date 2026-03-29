package main

import (
	"fmt"
	"log"

	"github.com/dhruvpurohit2k/expressions-india-backend/internal/event"
	"github.com/dhruvpurohit2k/expressions-india-backend/internal/models"
	"github.com/dhruvpurohit2k/expressions-india-backend/internal/pkg/utils"
	"github.com/dhruvpurohit2k/expressions-india-backend/internal/storage"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Server struct {
	r               *gin.Engine
	db              *gorm.DB
	s3              *storage.S3
	eventController *event.Controller
}

func initServer() *Server {
	fmt.Println("INITDB STARTING")
	db := storage.InitDB()
	s3 := storage.InitS3()

	err := db.AutoMigrate(&models.Event{}, &models.Media{})

	eventService := *event.NewService(db, s3)
	eventController := event.NewController(eventService)

	if err != nil {
		log.Fatal(err.Error())
	}
	r := gin.Default()

	return &Server{
		r:               r,
		db:              db,
		s3:              s3,
		eventController: eventController,
	}
}

func (s *Server) SetupRoutes() {
	s.r.GET("/hello", func(ctx *gin.Context) {
		utils.OK(ctx, gin.H{
			"message": "Success",
		})
	})
	s.r.GET("/event", s.eventController.GetAll)
}
