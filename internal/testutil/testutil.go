// Package testutil provides shared helpers for unit tests across all domain packages.
// It is only imported from *_test.go files.
package testutil

import (
	"os"
	"testing"

	"github.com/dhruvpurohit2k/expressions-india-backend/internal/auth"
	"github.com/dhruvpurohit2k/expressions-india-backend/internal/models"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const TestJWTSecret = "test-jwt-secret-do-not-use-in-production"

func init() {
	os.Setenv("JWT_SECRET", TestJWTSecret)
}

// NewTestDB opens a fresh in-memory SQLite database, runs AutoMigrate for all
// models, and seeds the audience table. Each call returns an isolated DB.
func NewTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	// Ensure JWT_SECRET is set for any auth operations triggered during the test.
	os.Setenv("JWT_SECRET", TestJWTSecret)

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("open in-memory DB: %v", err)
	}

	if err := db.AutoMigrate(
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
	); err != nil {
		t.Fatalf("migrate test DB: %v", err)
	}

	models.SeedAudience(db)
	return db
}

// AdminToken returns a signed JWT for an admin user suitable for Bearer auth headers.
func AdminToken(t *testing.T) string {
	t.Helper()
	os.Setenv("JWT_SECRET", TestJWTSecret)
	tok, err := auth.GenerateAccessToken("admin-test-id", "admin@test.com", true)
	if err != nil {
		t.Fatalf("generate admin token: %v", err)
	}
	return tok
}

// UserToken returns a signed JWT for a regular (non-admin) user.
func UserToken(t *testing.T) string {
	t.Helper()
	os.Setenv("JWT_SECRET", TestJWTSecret)
	tok, err := auth.GenerateAccessToken("user-test-id", "user@test.com", false)
	if err != nil {
		t.Fatalf("generate user token: %v", err)
	}
	return tok
}
