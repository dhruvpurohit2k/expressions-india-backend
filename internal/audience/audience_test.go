package audience_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dhruvpurohit2k/expressions-india-backend/internal/audience"
	"github.com/dhruvpurohit2k/expressions-india-backend/internal/testutil"
	"github.com/gin-gonic/gin"
)

func setupRouter(t *testing.T) *gin.Engine {
	t.Helper()
	db := testutil.NewTestDB(t)
	svc := audience.NewService(db)
	ctrl := audience.NewController(svc)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/api/audience/:name", ctrl.GetAudienceByName)
	r.GET("/admin/audience", ctrl.GetAudience)
	r.PUT("/admin/audience/:id", ctrl.UpdateDescription)
	return r
}

func parseResp(t *testing.T, w *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var m map[string]any
	if err := json.NewDecoder(w.Body).Decode(&m); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	return m
}

// ---- GetAudienceByName ----

func TestGetAudienceByName_SeededAudiences(t *testing.T) {
	r := setupRouter(t)

	seeded := []string{"student", "teacher", "parent", "counselor", "head_of_department", "mental_health_professional", "all"}
	for _, name := range seeded {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", fmt.Sprintf("/api/audience/%s", name), nil)
			r.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("expected 200, got %d for audience %q: %s", w.Code, name, w.Body.String())
			}
			resp := parseResp(t, w)
			if !resp["success"].(bool) {
				t.Errorf("expected success=true for %q", name)
			}
			data := resp["data"].(map[string]any)
			if data["name"] != name {
				t.Errorf("expected name=%s, got %v", name, data["name"])
			}
		})
	}
}

func TestGetAudienceByName_NotFound(t *testing.T) {
	r := setupRouter(t)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/audience/nonexistent", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

// ---- GetAudience (admin list) ----

func TestGetAudience_ExcludesAll(t *testing.T) {
	r := setupRouter(t)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/admin/audience", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	resp := parseResp(t, w)
	if !resp["success"].(bool) {
		t.Errorf("expected success=true")
	}
	data := resp["data"].([]any)

	// 6 seeded audiences excluding "all"
	if len(data) != 6 {
		t.Errorf("expected 6 audiences (excluding 'all'), got %d", len(data))
	}
	for _, item := range data {
		entry := item.(map[string]any)
		if entry["name"] == "all" {
			t.Errorf("'all' audience should be excluded from admin list")
		}
	}
}

// ---- UpdateDescription ----

func TestUpdateDescription_Success(t *testing.T) {
	r := setupRouter(t)

	// Fetch the audience list to get an ID.
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/admin/audience", nil)
	r.ServeHTTP(w, req)

	var listResp struct {
		Data []struct {
			ID   uint   `json:"ID"`
			Name string `json:"name"`
		} `json:"data"`
	}
	json.NewDecoder(w.Body).Decode(&listResp)
	if len(listResp.Data) == 0 {
		t.Fatal("no audiences found")
	}
	id := listResp.Data[0].ID

	body, _ := json.Marshal(map[string]string{
		"description": "Updated description for testing",
	})
	w = httptest.NewRecorder()
	req = httptest.NewRequest("PUT", fmt.Sprintf("/admin/audience/%d", id), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestUpdateDescription_NotFound(t *testing.T) {
	r := setupRouter(t)

	body, _ := json.Marshal(map[string]string{
		"description": "Test",
	})
	w := httptest.NewRecorder()
	req := httptest.NewRequest("PUT", "/admin/audience/99999", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestUpdateDescription_MissingBody(t *testing.T) {
	r := setupRouter(t)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("PUT", "/admin/audience/1", bytes.NewBufferString("{}"))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	// description is required; binding should fail with 400.
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}
