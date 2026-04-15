package storage

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func InitDB() *gorm.DB {
	env := os.Getenv("APP_ENV")

	const maxAttempts = 5
	var db *gorm.DB
	var err error

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		if attempt > 1 {
			wait := time.Duration(attempt*2) * time.Second
			log.Printf("DB connection failed, retrying in %s (attempt %d/%d)...", wait, attempt, maxAttempts)
			time.Sleep(wait)
		}

		if env == "production" {
			dns := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable&search_path=public",
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

		if err == nil {
			break
		}
	}

	if err != nil {
		log.Fatalf("Could not open DB after %d attempts: %v", maxAttempts, err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("failed to get underlying sql.DB: %v", err)
	}
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(30)
	sqlDB.SetConnMaxLifetime(time.Hour)
	return db
}
