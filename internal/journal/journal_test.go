// Package journal_test covers the HTTP handlers for the journal domain.
// Write operations that involve S3 are excluded from unit tests and will be
// covered by integration tests once AWS is configured.
package journal_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dhruvpurohit2k/expressions-india-backend/internal/journal"
	"github.com/dhruvpurohit2k/expressions-india-backend/internal/models"
	"github.com/dhruvpurohit2k/expressions-india-backend/internal/testutil"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func setupRouter(t *testing.T) (*gin.Engine, *gorm.DB) {
	t.Helper()
	db := testutil.NewTestDB(t)
	svc := journal.NewService(db, nil)
	ctrl := journal.NewController(svc)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/admin/journal", ctrl.GetList)
	r.GET("/admin/journal/:id", ctrl.GetById)
	r.GET("/api/journal", ctrl.GetList)
	r.GET("/api/journal/:id", ctrl.GetById)
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
func metaTotal(resp map[string]any) float64 {
	meta, ok := resp["meta"].(map[string]any)
	if !ok {
		return 0
	}
	total, _ := meta["total"].(float64)
	return total
}

func seedJournal(t *testing.T, db *gorm.DB, title string, year int) string {
	t.Helper()
	j := models.Journal{
		Title:      title,
		Year:       year,
		Volume:     1,
		Issue:      1,
		StartMonth: "January",
		EndMonth:   "June",
	}
	if err := db.Create(&j).Error; err != nil {
		t.Fatalf("seed journal %q: %v", title, err)
	}
	return j.ID
}

// ---- GetList ----

func TestGetList_Empty(t *testing.T) {
	r, _ := setupRouter(t)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/admin/journal", nil)
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

func TestGetList_WithData(t *testing.T) {
	r, db := setupRouter(t)
	seedJournal(t, db, "Expressions Vol 1", 2023)
	seedJournal(t, db, "Expressions Vol 2", 2024)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/admin/journal", nil)
	r.ServeHTTP(w, req)

	resp := parseResp(t, w)
	if metaTotal(resp) != 2 {
		t.Errorf("expected total=2, got %v", metaTotal(resp))
	}
}

func TestGetList_YearFilter(t *testing.T) {
	r, db := setupRouter(t)
	seedJournal(t, db, "Journal 2023", 2023)
	seedJournal(t, db, "Journal 2024", 2024)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/admin/journal?year=2023", nil)
	r.ServeHTTP(w, req)

	resp := parseResp(t, w)
	if metaTotal(resp) != 1 {
		t.Errorf("expected 1 result for year=2023, got %v", metaTotal(resp))
	}
}

// ---- GetById ----

func TestGetById_NotFound(t *testing.T) {
	r, _ := setupRouter(t)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/admin/journal/nonexistent-id", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestGetById_Success(t *testing.T) {
	r, db := setupRouter(t)
	id := seedJournal(t, db, "Test Journal", 2024)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", fmt.Sprintf("/admin/journal/%s", id), nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	resp := parseResp(t, w)
	data := resp["data"].(map[string]any)
	if data["id"] != id {
		t.Errorf("expected id=%s, got %v", id, data["id"])
	}
	if data["title"] != "Test Journal" {
		t.Errorf("unexpected title: %v", data["title"])
	}
}

// ---- Public routes mirror admin routes ----

func TestPublicGetList_Empty(t *testing.T) {
	r, _ := setupRouter(t)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/journal", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestPublicGetById_NotFound(t *testing.T) {
	r, _ := setupRouter(t)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/journal/nonexistent", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}
