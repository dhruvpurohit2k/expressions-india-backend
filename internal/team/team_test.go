package team_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dhruvpurohit2k/expressions-india-backend/internal/team"
	"github.com/dhruvpurohit2k/expressions-india-backend/internal/testutil"
	"github.com/gin-gonic/gin"
)

func setupRouter(t *testing.T) *gin.Engine {
	t.Helper()
	db := testutil.NewTestDB(t)
	svc := team.NewService(db)
	ctrl := team.NewController(svc)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/team", ctrl.GetList)
	r.GET("/team/:id", ctrl.GetById)
	r.POST("/team", ctrl.Create)
	r.PUT("/team/:id", ctrl.Update)
	r.DELETE("/team/:id", ctrl.Delete)
	return r
}

func jsonBody(t *testing.T, v any) *bytes.Buffer {
	t.Helper()
	b, _ := json.Marshal(v)
	return bytes.NewBuffer(b)
}

func parseResp(t *testing.T, w *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var m map[string]any
	if err := json.NewDecoder(w.Body).Decode(&m); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	return m
}

func createTeam(t *testing.T, r *gin.Engine) string {
	t.Helper()
	body := map[string]any{
		"description": "A test team",
		"members": []map[string]any{
			{"position": "Lead", "holders": []string{"Alice"}},
		},
	}
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/team", jsonBody(t, body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("createTeam failed: %d %s", w.Code, w.Body.String())
	}

	// Fetch list to get the created team's ID.
	w = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/team", nil)
	r.ServeHTTP(w, req)
	var listResp struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	json.NewDecoder(w.Body).Decode(&listResp)
	if len(listResp.Data) == 0 {
		t.Fatalf("no teams in list after create")
	}
	return listResp.Data[0].ID
}

// ---- GetList ----

func TestGetList_Empty(t *testing.T) {
	r := setupRouter(t)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/team", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	resp := parseResp(t, w)
	if !resp["success"].(bool) {
		t.Errorf("expected success=true")
	}
	data := resp["data"].([]any)
	if len(data) != 0 {
		t.Errorf("expected empty list, got %d items", len(data))
	}
}

func TestGetList_WithData(t *testing.T) {
	r := setupRouter(t)
	createTeam(t, r)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/team", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	resp := parseResp(t, w)
	data := resp["data"].([]any)
	if len(data) != 1 {
		t.Errorf("expected 1 team, got %d", len(data))
	}
}

// ---- GetById ----

func TestGetById_NotFound(t *testing.T) {
	r := setupRouter(t)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/team/nonexistent-id", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestGetById_Success(t *testing.T) {
	r := setupRouter(t)
	id := createTeam(t, r)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", fmt.Sprintf("/team/%s", id), nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	resp := parseResp(t, w)
	data := resp["data"].(map[string]any)
	if data["id"] != id {
		t.Errorf("expected id=%s, got %s", id, data["id"])
	}
	if data["description"] != "A test team" {
		t.Errorf("unexpected description: %v", data["description"])
	}
}

// ---- Create ----

func TestCreate_Success(t *testing.T) {
	r := setupRouter(t)

	body := map[string]any{
		"description": "New team",
		"members": []map[string]any{
			{"position": "Developer", "holders": []string{"Bob", "Carol"}},
		},
	}
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/team", jsonBody(t, body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestCreate_InvalidJSON(t *testing.T) {
	r := setupRouter(t)

	// Send malformed JSON — ShouldBindJSON should return 400.
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/team", bytes.NewBufferString("{invalid json"))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestCreate_EmptyMembers(t *testing.T) {
	r := setupRouter(t)

	body := map[string]any{
		"description": "Team without members",
		"members":     []any{},
	}
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/team", jsonBody(t, body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

// ---- Update ----

func TestUpdate_Success(t *testing.T) {
	r := setupRouter(t)
	id := createTeam(t, r)

	body := map[string]any{
		"description": "Updated description",
		"members": []map[string]any{
			{"position": "Manager", "holders": []string{"Dave"}},
		},
	}
	w := httptest.NewRecorder()
	req := httptest.NewRequest("PUT", fmt.Sprintf("/team/%s", id), jsonBody(t, body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	// Verify the update persisted.
	w = httptest.NewRecorder()
	req = httptest.NewRequest("GET", fmt.Sprintf("/team/%s", id), nil)
	r.ServeHTTP(w, req)

	var getResp struct {
		Data struct {
			Description string `json:"description"`
		} `json:"data"`
	}
	json.NewDecoder(w.Body).Decode(&getResp)
	if getResp.Data.Description != "Updated description" {
		t.Errorf("update not persisted, got: %s", getResp.Data.Description)
	}
}

func TestUpdate_NotFound(t *testing.T) {
	r := setupRouter(t)

	body := map[string]any{
		"description": "Whatever",
		"members":     []any{},
	}
	w := httptest.NewRecorder()
	req := httptest.NewRequest("PUT", "/team/nonexistent-id", jsonBody(t, body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

// ---- Delete ----

func TestDelete_Success(t *testing.T) {
	r := setupRouter(t)
	id := createTeam(t, r)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("DELETE", fmt.Sprintf("/team/%s", id), nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	// Verify it no longer exists.
	w = httptest.NewRecorder()
	req = httptest.NewRequest("GET", fmt.Sprintf("/team/%s", id), nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404 after delete, got %d", w.Code)
	}
}

func TestDelete_NotFound(t *testing.T) {
	r := setupRouter(t)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("DELETE", "/team/nonexistent-id", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}
