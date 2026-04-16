package latestactivity_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	latestactivity "github.com/dhruvpurohit2k/expressions-india-backend/internal/latest-activity"
	"github.com/dhruvpurohit2k/expressions-india-backend/internal/models"
	"github.com/dhruvpurohit2k/expressions-india-backend/internal/testutil"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func setupRouter(t *testing.T) (*gin.Engine, *gorm.DB) {
	t.Helper()
	db := testutil.NewTestDB(t)
	svc := latestactivity.NewService(db)
	ctrl := latestactivity.NewController(svc)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/api/latest-activity", ctrl.GetLatestActivity)
	return r, db
}

func parseResp(t *testing.T, w *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var m map[string]any
	if err := json.NewDecoder(w.Body).Decode(&m); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	return m
}

func TestGetLatestActivity_Empty(t *testing.T) {
	r, _ := setupRouter(t)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/latest-activity", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	resp := parseResp(t, w)
	if !resp["success"].(bool) {
		t.Errorf("expected success=true")
	}
}

func TestGetLatestActivity_WithEvents(t *testing.T) {
	r, db := setupRouter(t)

	// Seed events directly to avoid S3 dependency.
	status := "upcoming"
	for i := 0; i < 3; i++ {
		ev := models.Event{
			Title:     "Event",
			StartDate: time.Now().Add(time.Duration(i+1) * 24 * time.Hour),
			Status:    &status,
		}
		if err := db.Create(&ev).Error; err != nil {
			t.Fatalf("seed event: %v", err)
		}
	}

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/latest-activity", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	resp := parseResp(t, w)
	if !resp["success"].(bool) {
		t.Errorf("expected success=true")
	}
	data, ok := resp["data"].([]any)
	if !ok {
		t.Fatalf("data should be an array, got %T", resp["data"])
	}
	if len(data) != 3 {
		t.Errorf("expected 3 activities, got %d", len(data))
	}
}

func TestGetLatestActivity_LimitedToFive(t *testing.T) {
	r, db := setupRouter(t)

	// Seed 7 events; the service limits to 5.
	status := "upcoming"
	for i := 0; i < 7; i++ {
		ev := models.Event{
			Title:     "Event",
			StartDate: time.Now().Add(time.Duration(i+1) * 24 * time.Hour),
			Status:    &status,
		}
		if err := db.Create(&ev).Error; err != nil {
			t.Fatalf("seed event: %v", err)
		}
	}

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/latest-activity", nil)
	r.ServeHTTP(w, req)

	resp := parseResp(t, w)
	data := resp["data"].([]any)
	if len(data) > 5 {
		t.Errorf("expected at most 5 activities (LIMIT 5), got %d", len(data))
	}
}

func TestGetLatestActivity_MixedContentTypes(t *testing.T) {
	r, db := setupRouter(t)

	// Seed one of each content type.
	status := "upcoming"
	db.Create(&models.Event{Title: "E1", StartDate: time.Now(), Status: &status})
	db.Create(&models.Article{Title: "A1"})
	db.Create(&models.Podcast{Title: "P1"})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/latest-activity", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	resp := parseResp(t, w)
	data := resp["data"].([]any)
	if len(data) != 3 {
		t.Errorf("expected 3 activities (event + article + podcast), got %d", len(data))
	}
}
