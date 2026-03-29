package storage

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/dhruvpurohit2k/expressions-india-backend/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func InitDB() *gorm.DB {
	env := os.Getenv("APP_ENV")
	freshStart := os.Getenv("DB_FRESH_START") == "true"
	var db *gorm.DB
	var err error
	if env == "production" {
		dns := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
			os.Getenv("DB_USERNAME"),
			os.Getenv("DB_PASSWORD"),
			os.Getenv("DB_HOST"),
			os.Getenv("DB_PORT"),
			os.Getenv("DB_NAME"))

		db, err = gorm.Open(postgres.Open(dns), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Error),
		})
	} else {
		db, err = gorm.Open(sqlite.Open("dev.db"), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Info),
		})
	}
	if err != nil {
		log.Fatal("Could not open DB", err.Error())
	}

	if env == "development" && freshStart {
		db.Migrator().DropTable(&models.Event{}, &models.Media{})
	}
	sqlDB, _ := db.DB()
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(30)
	sqlDB.SetConnMaxLifetime(time.Hour)
	return db

}
