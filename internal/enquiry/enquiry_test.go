package enquiry_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dhruvpurohit2k/expressions-india-backend/internal/enquiry"
	"github.com/dhruvpurohit2k/expressions-india-backend/internal/testutil"
	"github.com/gin-gonic/gin"
)

func setupRouter(t *testing.T) *gin.Engine {
	t.Helper()
	db := testutil.NewTestDB(t)
	svc := enquiry.NewService(db)
	ctrl := enquiry.NewController(svc)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/admin/enquiry", ctrl.GetList)
	r.GET("/admin/enquiry/:id", ctrl.GetById)
	r.DELETE("/admin/enquiry/:id", ctrl.Delete)
	r.POST("/api/enquiry", ctrl.CreateEnquiry)
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

// createEnquiry posts a valid enquiry and returns its ID.
func createEnquiry(t *testing.T, r *gin.Engine) string {
	t.Helper()
	body := map[string]string{
		"subject": "Test Subject",
		"name":    "Test User",
		"email":   "test@example.com",
		"phone":   "9876543210",
		"message": "This is a test message",
	}
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/enquiry", jsonBody(t, body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("createEnquiry failed: %d %s", w.Code, w.Body.String())
	}

	// Fetch the list to get the ID.
	w = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/admin/enquiry", nil)
	r.ServeHTTP(w, req)
	var listResp struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	json.NewDecoder(w.Body).Decode(&listResp)
	if len(listResp.Data) == 0 {
		t.Fatalf("no enquiries in list after create")
	}
	return listResp.Data[0].ID
}

// ---- CreateEnquiry ----

func TestCreateEnquiry_Success(t *testing.T) {
	r := setupRouter(t)

	body := map[string]string{
		"subject": "Question about workshops",
		"name":    "Alice Smith",
		"email":   "alice@example.com",
		"phone":   "1234567890",
		"message": "I would like to know more about upcoming workshops.",
	}
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/enquiry", jsonBody(t, body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	resp := parseResp(t, w)
	if !resp["success"].(bool) {
		t.Errorf("expected success=true")
	}
}

func TestCreateEnquiry_MissingSubject(t *testing.T) {
	r := setupRouter(t)

	body := map[string]string{
		"name":    "Alice",
		"email":   "alice@example.com",
		"phone":   "1234567890",
		"message": "Hello",
	}
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/enquiry", jsonBody(t, body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestCreateEnquiry_InvalidEmail(t *testing.T) {
	r := setupRouter(t)

	body := map[string]string{
		"subject": "Hi",
		"name":    "Alice",
		"email":   "not-an-email",
		"phone":   "1234567890",
		"message": "Hello",
	}
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/enquiry", jsonBody(t, body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestCreateEnquiry_MissingMessage(t *testing.T) {
	r := setupRouter(t)

	body := map[string]string{
		"subject": "Hi",
		"name":    "Alice",
		"email":   "alice@example.com",
		"phone":   "1234567890",
	}
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/enquiry", jsonBody(t, body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

// ---- GetList ----

func TestGetList_Empty(t *testing.T) {
	r := setupRouter(t)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/admin/enquiry", nil)
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
		t.Errorf("expected empty list, got %d", len(data))
	}
}

func TestGetList_WithData(t *testing.T) {
	r := setupRouter(t)
	createEnquiry(t, r)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/admin/enquiry", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	resp := parseResp(t, w)
	data := resp["data"].([]any)
	if len(data) != 1 {
		t.Errorf("expected 1 enquiry, got %d", len(data))
	}
}

func TestGetList_FilterByName(t *testing.T) {
	r := setupRouter(t)
	createEnquiry(t, r)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/admin/enquiry?name=Test", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	resp := parseResp(t, w)
	data := resp["data"].([]any)
	if len(data) != 1 {
		t.Errorf("expected 1 result for name filter, got %d", len(data))
	}

	// Non-matching filter should return empty.
	w = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/admin/enquiry?name=Nobody", nil)
	r.ServeHTTP(w, req)
	resp = parseResp(t, w)
	data = resp["data"].([]any)
	if len(data) != 0 {
		t.Errorf("expected 0 results for non-matching name, got %d", len(data))
	}
}

// ---- GetById ----

func TestGetById_NotFound(t *testing.T) {
	r := setupRouter(t)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/admin/enquiry/nonexistent-id", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestGetById_Success(t *testing.T) {
	r := setupRouter(t)
	id := createEnquiry(t, r)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", fmt.Sprintf("/admin/enquiry/%s", id), nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	resp := parseResp(t, w)
	data := resp["data"].(map[string]any)
	if data["id"] != id {
		t.Errorf("expected id=%s, got %v", id, data["id"])
	}
	if data["subject"] != "Test Subject" {
		t.Errorf("unexpected subject: %v", data["subject"])
	}
}

// ---- Delete ----

func TestDelete_Success(t *testing.T) {
	r := setupRouter(t)
	id := createEnquiry(t, r)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("DELETE", fmt.Sprintf("/admin/enquiry/%s", id), nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	// Verify it no longer exists.
	w = httptest.NewRecorder()
	req = httptest.NewRequest("GET", fmt.Sprintf("/admin/enquiry/%s", id), nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404 after delete, got %d", w.Code)
	}
}
