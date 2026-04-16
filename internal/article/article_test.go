// Package article_test covers the HTTP handlers for the article domain.
// Write operations (Create, Update, Delete) involving S3 are excluded and will
// be covered by integration tests once AWS is configured.
package article_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dhruvpurohit2k/expressions-india-backend/internal/article"
	"github.com/dhruvpurohit2k/expressions-india-backend/internal/models"
	"github.com/dhruvpurohit2k/expressions-india-backend/internal/testutil"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func setupRouter(t *testing.T) (*gin.Engine, *gorm.DB) {
	t.Helper()
	db := testutil.NewTestDB(t)
	svc := article.NewService(db, nil)
	ctrl := article.NewController(svc)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/admin/article", ctrl.GetArticleList)
	r.GET("/admin/article/:id", ctrl.GetArticleById)
	r.GET("/api/article", ctrl.GetArticleListPaginated)
	r.GET("/api/article/:id", ctrl.GetArticleById)
	r.GET("/api/article/audience/:audience", ctrl.GetArticlesByAudience)
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
// Returns 0 when meta is absent or when total is omitted (omitempty when zero).
func metaTotal(resp map[string]any) float64 {
	meta, ok := resp["meta"].(map[string]any)
	if !ok {
		return 0
	}
	total, _ := meta["total"].(float64)
	return total
}

func seedArticle(t *testing.T, db *gorm.DB, title, category string) string {
	t.Helper()
	a := models.Article{Title: title, Category: category}
	if err := db.Create(&a).Error; err != nil {
		t.Fatalf("seed article %q: %v", title, err)
	}
	return a.ID
}

// ---- Admin GetArticleList ----

func TestGetArticleList_Empty(t *testing.T) {
	r, _ := setupRouter(t)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/admin/article", nil)
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

func TestGetArticleList_WithData(t *testing.T) {
	r, db := setupRouter(t)
	seedArticle(t, db, "Mental Health 101", "health")
	seedArticle(t, db, "Study Tips", "education")

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/admin/article", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	resp := parseResp(t, w)
	if metaTotal(resp) != 2 {
		t.Errorf("expected total=2, got %v", metaTotal(resp))
	}
}

func TestGetArticleList_SearchFilter(t *testing.T) {
	r, db := setupRouter(t)
	seedArticle(t, db, "Mental Health 101", "health")
	seedArticle(t, db, "Study Tips", "education")

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/admin/article?search=mental", nil)
	r.ServeHTTP(w, req)

	resp := parseResp(t, w)
	if metaTotal(resp) != 1 {
		t.Errorf("expected 1 result for search=mental, got %v", metaTotal(resp))
	}
}

// ---- Admin GetArticleById ----

func TestGetArticleById_Admin_NotFound(t *testing.T) {
	r, _ := setupRouter(t)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/admin/article/nonexistent-id", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestGetArticleById_Admin_Success(t *testing.T) {
	r, db := setupRouter(t)
	id := seedArticle(t, db, "Test Article", "general")

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", fmt.Sprintf("/admin/article/%s", id), nil)
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

// ---- Public GetArticleListPaginated ----

func TestGetArticleListPaginated_Empty(t *testing.T) {
	r, _ := setupRouter(t)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/article", nil)
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

func TestGetArticleListPaginated_WithData(t *testing.T) {
	r, db := setupRouter(t)
	seedArticle(t, db, "Article 1", "cat1")
	seedArticle(t, db, "Article 2", "cat2")

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/article?limit=10&offset=0", nil)
	r.ServeHTTP(w, req)

	resp := parseResp(t, w)
	if metaTotal(resp) != 2 {
		t.Errorf("expected total=2, got %v", metaTotal(resp))
	}
}

// ---- Public GetArticleById ----

func TestGetArticleById_Public_NotFound(t *testing.T) {
	r, _ := setupRouter(t)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/article/nonexistent", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

// ---- Public GetArticlesByAudience ----

func TestGetArticlesByAudience_Empty(t *testing.T) {
	r, _ := setupRouter(t)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/article/audience/student", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	resp := parseResp(t, w)
	if metaTotal(resp) != 0 {
		t.Errorf("expected total=0")
	}
}
