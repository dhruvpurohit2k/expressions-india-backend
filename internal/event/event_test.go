// Package event_test covers the HTTP handlers for the event domain.
// Write operations (Create, Update, Delete) that involve S3 are excluded from
// unit tests and will be covered by integration tests once AWS is configured.
package event_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/dhruvpurohit2k/expressions-india-backend/internal/event"
	"github.com/dhruvpurohit2k/expressions-india-backend/internal/models"
	"github.com/dhruvpurohit2k/expressions-india-backend/internal/testutil"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func setupRouter(t *testing.T) (*gin.Engine, *gorm.DB) {
	t.Helper()
	db := testutil.NewTestDB(t)
	// Unit tests do not exercise S3; pass nil for the storage.S3 parameter.
	// Only GET endpoints that do not touch S3 are tested here.
	svc := event.NewService(db, nil)
	ctrl := event.NewController(svc)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/admin/allEvents", ctrl.GetAll)
	r.GET("/admin/event", ctrl.GetEventList)
	r.GET("/admin/event/:id", ctrl.GetEventById)
	r.GET("/api/event/upcoming", ctrl.GetUpcomingEvents)
	r.GET("/api/event/past", ctrl.GetPastEvents)
	r.GET("/api/event/:id", ctrl.GetEventById)
	r.GET("/api/event/audience/:audience", ctrl.GetUpcomingEventsByAudience)
	r.GET("/api/home/images", ctrl.GetHomePageImages)
	r.GET("/api/home/upcoming-images", ctrl.GetUpcomingCarouselImages)
	r.GET("/api/home/completed-images", ctrl.GetCompletedCarouselImages)
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

// metaTotal safely reads "total" from a paginated response.
// Returns 0 when meta is absent or when total is omitted due to omitempty (value is 0).
func metaTotal(resp map[string]any) float64 {
	meta, ok := resp["meta"].(map[string]any)
	if !ok {
		return 0
	}
	total, _ := meta["total"].(float64)
	return total
}

// seedEvent inserts a minimal event directly via GORM, bypassing the S3 upload
// path, so we can test GET endpoints with real data.
func seedEvent(t *testing.T, db *gorm.DB, title, status string, startDate time.Time) {
	t.Helper()
	ev := models.Event{
		Title:     title,
		StartDate: startDate,
		Status:    &status,
	}
	if err := db.Create(&ev).Error; err != nil {
		t.Fatalf("seed event %q: %v", title, err)
	}
}

// ---- GetAll ----

func TestGetAll_Empty(t *testing.T) {
	r, _ := setupRouter(t)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/admin/allEvents", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	resp := parseResp(t, w)
	if !resp["success"].(bool) {
		t.Errorf("expected success=true")
	}
}

// ---- GetEventList (admin paginated) ----

func TestGetEventList_Empty(t *testing.T) {
	r, _ := setupRouter(t)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/admin/event", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	resp := parseResp(t, w)
	if !resp["success"].(bool) {
		t.Errorf("expected success=true")
	}
	if metaTotal(resp) != 0 {
		t.Errorf("expected total=0, got %v", metaTotal(resp))
	}
}

func TestGetEventList_WithData(t *testing.T) {
	r, db := setupRouter(t)
	seedEvent(t, db, "Workshop A", "upcoming", time.Now().Add(24*time.Hour))
	seedEvent(t, db, "Workshop B", "completed", time.Now().Add(-24*time.Hour))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/admin/event", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	resp := parseResp(t, w)
	if metaTotal(resp) != 2 {
		t.Errorf("expected total=2, got %v", metaTotal(resp))
	}
}

// ---- GetEventById ----

func TestGetEventById_NotFound(t *testing.T) {
	r, _ := setupRouter(t)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/admin/event/nonexistent-id", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

// ---- GetUpcomingEvents ----

func TestGetUpcomingEvents_Empty(t *testing.T) {
	r, _ := setupRouter(t)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/event/upcoming", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	resp := parseResp(t, w)
	if !resp["success"].(bool) {
		t.Errorf("expected success=true")
	}
	if metaTotal(resp) != 0 {
		t.Errorf("expected total=0")
	}
}

func TestGetUpcomingEvents_WithData(t *testing.T) {
	r, db := setupRouter(t)
	seedEvent(t, db, "Future Workshop", "upcoming", time.Now().Add(48*time.Hour))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/event/upcoming", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	resp := parseResp(t, w)
	if metaTotal(resp) != 1 {
		t.Errorf("expected total=1, got %v", metaTotal(resp))
	}
}

func TestGetUpcomingEvents_DoesNotIncludePastEvents(t *testing.T) {
	r, db := setupRouter(t)
	// Past event with status "upcoming" should not appear (filter checks start_date >= now).
	seedEvent(t, db, "Old Upcoming", "upcoming", time.Now().Add(-1*time.Hour))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/event/upcoming", nil)
	r.ServeHTTP(w, req)

	resp := parseResp(t, w)
	if metaTotal(resp) != 0 {
		t.Errorf("past event with 'upcoming' status should not appear in upcoming list")
	}
}

// ---- GetPastEvents ----

func TestGetPastEvents_Empty(t *testing.T) {
	r, _ := setupRouter(t)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/event/past", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	resp := parseResp(t, w)
	if metaTotal(resp) != 0 {
		t.Errorf("expected total=0")
	}
}

func TestGetPastEvents_WithData(t *testing.T) {
	r, db := setupRouter(t)
	seedEvent(t, db, "Past Seminar", "completed", time.Now().Add(-48*time.Hour))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/event/past", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	resp := parseResp(t, w)
	if metaTotal(resp) != 1 {
		t.Errorf("expected total=1, got %v", metaTotal(resp))
	}
}

// ---- Home page image endpoints ----

func TestGetHomePageImages_Empty(t *testing.T) {
	r, _ := setupRouter(t)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/home/images", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	resp := parseResp(t, w)
	if !resp["success"].(bool) {
		t.Errorf("expected success=true")
	}
	data := resp["data"].([]any)
	if len(data) != 0 {
		t.Errorf("expected empty image list, got %d", len(data))
	}
}

func TestGetUpcomingCarouselImages_Empty(t *testing.T) {
	r, _ := setupRouter(t)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/home/upcoming-images", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	resp := parseResp(t, w)
	if !resp["success"].(bool) {
		t.Errorf("expected success=true")
	}
}

func TestGetCompletedCarouselImages_Empty(t *testing.T) {
	r, _ := setupRouter(t)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/home/completed-images", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	resp := parseResp(t, w)
	if !resp["success"].(bool) {
		t.Errorf("expected success=true")
	}
}

// ---- GetUpcomingEventsByAudience ----

func TestGetUpcomingEventsByAudience_Empty(t *testing.T) {
	r, _ := setupRouter(t)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/event/audience/student", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	resp := parseResp(t, w)
	if metaTotal(resp) != 0 {
		t.Errorf("expected total=0")
	}
}

// ---- Public GetEventById ----

func TestPublicGetEventById_NotFound(t *testing.T) {
	r, _ := setupRouter(t)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/event/nonexistent-id", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}
