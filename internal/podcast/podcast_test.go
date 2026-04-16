// Package podcast_test covers the HTTP handlers for the podcast domain.
package podcast_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dhruvpurohit2k/expressions-india-backend/internal/models"
	"github.com/dhruvpurohit2k/expressions-india-backend/internal/podcast"
	"github.com/dhruvpurohit2k/expressions-india-backend/internal/testutil"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func setupRouter(t *testing.T) (*gin.Engine, *gorm.DB) {
	t.Helper()
	db := testutil.NewTestDB(t)
	svc := podcast.NewService(db)
	ctrl := podcast.NewController(svc)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/admin/podcast", ctrl.GetPodcastList)
	r.GET("/admin/podcast/:id", ctrl.GetById)
	r.DELETE("/admin/podcast/:id", ctrl.Delete)
	r.GET("/api/podcast", ctrl.GetPodcastList)
	r.GET("/api/podcast/:id", ctrl.GetById)
	r.GET("/api/podcast/audience/:audience", ctrl.GetPodcastsByAudience)
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
// Returns 0 when meta is absent or total is omitted (omitempty when zero).
func metaTotal(resp map[string]any) float64 {
	meta, ok := resp["meta"].(map[string]any)
	if !ok {
		return 0
	}
	total, _ := meta["total"].(float64)
	return total
}

func seedPodcast(t *testing.T, db *gorm.DB, title string) string {
	t.Helper()
	p := models.Podcast{Title: title, Link: "https://example.com/podcast"}
	if err := db.Create(&p).Error; err != nil {
		t.Fatalf("seed podcast %q: %v", title, err)
	}
	return p.ID
}

// ---- GetPodcastList ----

func TestGetPodcastList_Empty(t *testing.T) {
	r, _ := setupRouter(t)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/admin/podcast", nil)
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

func TestGetPodcastList_WithData(t *testing.T) {
	r, db := setupRouter(t)
	seedPodcast(t, db, "Episode 1")
	seedPodcast(t, db, "Episode 2")

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/admin/podcast", nil)
	r.ServeHTTP(w, req)

	resp := parseResp(t, w)
	if metaTotal(resp) != 2 {
		t.Errorf("expected total=2, got %v", metaTotal(resp))
	}
}

func TestGetPodcastList_SearchFilter(t *testing.T) {
	r, db := setupRouter(t)
	seedPodcast(t, db, "Mental Health Talk")
	seedPodcast(t, db, "Study Strategies")

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/admin/podcast?search=mental", nil)
	r.ServeHTTP(w, req)

	resp := parseResp(t, w)
	if metaTotal(resp) != 1 {
		t.Errorf("expected 1 result for search=mental, got %v", metaTotal(resp))
	}
}

// ---- GetById ----

func TestGetById_NotFound(t *testing.T) {
	r, _ := setupRouter(t)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/admin/podcast/nonexistent-id", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestGetById_Success(t *testing.T) {
	r, db := setupRouter(t)
	id := seedPodcast(t, db, "Wellness Series")

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", fmt.Sprintf("/admin/podcast/%s", id), nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	resp := parseResp(t, w)
	data := resp["data"].(map[string]any)
	if data["id"] != id {
		t.Errorf("expected id=%s, got %v", id, data["id"])
	}
}

// ---- Delete ----

func TestDelete_Success(t *testing.T) {
	r, db := setupRouter(t)
	id := seedPodcast(t, db, "To Be Deleted")

	w := httptest.NewRecorder()
	req := httptest.NewRequest("DELETE", fmt.Sprintf("/admin/podcast/%s", id), nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	// Verify it no longer exists.
	w = httptest.NewRecorder()
	req = httptest.NewRequest("GET", fmt.Sprintf("/admin/podcast/%s", id), nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404 after delete, got %d", w.Code)
	}
}

// ---- Public GetPodcastsByAudience ----

func TestGetPodcastsByAudience_Empty(t *testing.T) {
	r, _ := setupRouter(t)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/podcast/audience/student", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	resp := parseResp(t, w)
	if metaTotal(resp) != 0 {
		t.Errorf("expected total=0, got %v", metaTotal(resp))
	}
}
