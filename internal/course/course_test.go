// Package course_test covers the HTTP handlers for the course domain.
// Write operations (Create, Update, Delete) that involve S3 are excluded from
// unit tests and will be covered by integration tests once AWS is configured.
package course_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dhruvpurohit2k/expressions-india-backend/internal/auth"
	"github.com/dhruvpurohit2k/expressions-india-backend/internal/course"
	"github.com/dhruvpurohit2k/expressions-india-backend/internal/models"
	"github.com/dhruvpurohit2k/expressions-india-backend/internal/testutil"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func setupRouter(t *testing.T) (*gin.Engine, *gorm.DB) {
	t.Helper()
	db := testutil.NewTestDB(t)
	svc := course.NewService(db, nil)
	ctrl := course.NewController(svc)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/admin/course", ctrl.GetCoursesListAdmin)
	r.GET("/admin/course/:id", ctrl.GetCourseById)
	r.GET("/admin/course/:id/enrolled", ctrl.GetEnrolledUsers)
	r.GET("/admin/course/:id/not-enrolled", ctrl.GetNotEnrolledUsers)
	r.POST("/admin/course/:id/enroll", ctrl.EnrollUser)
	r.DELETE("/admin/course/:id/enrolled/:userId", ctrl.RevokeAccess)
	r.GET("/api/course", ctrl.GetCoursesList)
	r.GET("/api/course/:id", ctrl.GetCourseById)
	r.GET("/api/course/audience/:audience", ctrl.GetCoursesByAudience)
	r.GET("/api/course/:id/chapter/:chapterId", auth.TryExtractClaims(), ctrl.GetChapterById)
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

func seedCourse(t *testing.T, db *gorm.DB, title string) string {
	t.Helper()
	c := models.Course{Title: title, Description: "Test course description"}
	if err := db.Create(&c).Error; err != nil {
		t.Fatalf("seed course %q: %v", title, err)
	}
	return c.ID
}

// ---- Admin GetCoursesListAdmin ----

func TestGetCoursesListAdmin_Empty(t *testing.T) {
	r, _ := setupRouter(t)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/admin/course", nil)
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

func TestGetCoursesListAdmin_WithData(t *testing.T) {
	r, db := setupRouter(t)
	seedCourse(t, db, "Mindfulness 101")
	seedCourse(t, db, "Stress Management")

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/admin/course", nil)
	r.ServeHTTP(w, req)

	resp := parseResp(t, w)
	if metaTotal(resp) != 2 {
		t.Errorf("expected total=2, got %v", metaTotal(resp))
	}
}

// ---- GetCourseById ----

func TestGetCourseById_NotFound(t *testing.T) {
	r, _ := setupRouter(t)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/admin/course/nonexistent-id", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestGetCourseById_Success(t *testing.T) {
	r, db := setupRouter(t)
	id := seedCourse(t, db, "Yoga for Students")

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", fmt.Sprintf("/admin/course/%s", id), nil)
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

// ---- Public GetCoursesList ----

func TestGetCoursesList_Empty(t *testing.T) {
	r, _ := setupRouter(t)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/course", nil)
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

// ---- GetCoursesByAudience ----

func TestGetCoursesByAudience_Empty(t *testing.T) {
	r, _ := setupRouter(t)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/course/audience/student", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	resp := parseResp(t, w)
	if metaTotal(resp) != 0 {
		t.Errorf("expected total=0")
	}
}

// ---- GetEnrolledUsers ----

func TestGetEnrolledUsers_Empty(t *testing.T) {
	r, db := setupRouter(t)
	id := seedCourse(t, db, "Empty Course")

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", fmt.Sprintf("/admin/course/%s/enrolled", id), nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	resp := parseResp(t, w)
	if metaTotal(resp) != 0 {
		t.Errorf("expected total=0")
	}
}

// ---- EnrollUser / RevokeAccess ----

func TestEnrollUser_Success(t *testing.T) {
	r, db := setupRouter(t)
	courseID := seedCourse(t, db, "Enrollment Course")

	// Create a user to enroll.
	pw := "hashed"
	user := models.User{Email: "enroll@example.com", Password: &pw, IsAdmin: false, Provider: models.ProviderPassword, ProviderSub: "password:enroll@example.com"}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("seed user: %v", err)
	}

	body, _ := json.Marshal(map[string]string{"userId": user.ID})
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", fmt.Sprintf("/admin/course/%s/enroll", courseID), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	// Verify enrollment.
	w = httptest.NewRecorder()
	req = httptest.NewRequest("GET", fmt.Sprintf("/admin/course/%s/enrolled", courseID), nil)
	r.ServeHTTP(w, req)

	resp := parseResp(t, w)
	if metaTotal(resp) != 1 {
		t.Errorf("expected 1 enrolled user, got %v", metaTotal(resp))
	}
}

func TestRevokeAccess_Success(t *testing.T) {
	r, db := setupRouter(t)
	courseID := seedCourse(t, db, "Revoke Course")

	pw2 := "hashed"
	user := models.User{Email: "revoke@example.com", Password: &pw2, IsAdmin: false, Provider: models.ProviderPassword, ProviderSub: "password:revoke@example.com"}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("seed user: %v", err)
	}

	// Enroll.
	body, _ := json.Marshal(map[string]string{"userId": user.ID})
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", fmt.Sprintf("/admin/course/%s/enroll", courseID), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("enroll failed: %d", w.Code)
	}

	// Revoke.
	w = httptest.NewRecorder()
	req = httptest.NewRequest("DELETE", fmt.Sprintf("/admin/course/%s/enrolled/%s", courseID, user.ID), nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("revoke failed: %d: %s", w.Code, w.Body.String())
	}

	// Verify no longer enrolled.
	w = httptest.NewRecorder()
	req = httptest.NewRequest("GET", fmt.Sprintf("/admin/course/%s/enrolled", courseID), nil)
	r.ServeHTTP(w, req)
	resp := parseResp(t, w)
	if metaTotal(resp) != 0 {
		t.Errorf("expected 0 enrolled after revoke, got %v", metaTotal(resp))
	}
}
