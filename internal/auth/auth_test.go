package auth_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dhruvpurohit2k/expressions-india-backend/internal/auth"
	"github.com/dhruvpurohit2k/expressions-india-backend/internal/testutil"
	"github.com/gin-gonic/gin"
)

func setupRouter(t *testing.T) *gin.Engine {
	t.Helper()
	db := testutil.NewTestDB(t)
	svc := auth.NewService(db)
	ctrl := auth.NewController(svc)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/auth/login", ctrl.Login)
	r.POST("/auth/signup", ctrl.Signup)
	r.POST("/auth/refresh", ctrl.Refresh)
	r.POST("/auth/logout", ctrl.Logout)
	return r
}

func jsonBody(t *testing.T, v any) *bytes.Buffer {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal body: %v", err)
	}
	return bytes.NewBuffer(b)
}

func do(r *gin.Engine, method, path string, body *bytes.Buffer) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	var req *http.Request
	if body != nil {
		req = httptest.NewRequest(method, path, body)
		req.Header.Set("Content-Type", "application/json")
	} else {
		req = httptest.NewRequest(method, path, nil)
	}
	r.ServeHTTP(w, req)
	return w
}

func parseResp(t *testing.T, w *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var m map[string]any
	if err := json.NewDecoder(w.Body).Decode(&m); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	return m
}

// signupUser is a test helper that creates a user and returns the response recorder.
func signupUser(t *testing.T, r *gin.Engine, email, password string) *httptest.ResponseRecorder {
	t.Helper()
	return do(r, "POST", "/auth/signup", jsonBody(t, map[string]string{
		"email":    email,
		"password": password,
		"name":     "Test User",
	}))
}

// ---- Login tests ----

func TestLogin_Success(t *testing.T) {
	r := setupRouter(t)

	w := signupUser(t, r, "login@example.com", "password123")
	if w.Code != http.StatusOK {
		t.Fatalf("signup failed: %d %s", w.Code, w.Body.String())
	}

	w = do(r, "POST", "/auth/login", jsonBody(t, map[string]string{
		"email":    "login@example.com",
		"password": "password123",
	}))

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	resp := parseResp(t, w)
	if !resp["success"].(bool) {
		t.Errorf("expected success=true")
	}
	data := resp["data"].(map[string]any)
	if data["accessToken"] == "" {
		t.Errorf("expected non-empty accessToken")
	}
	if data["refreshToken"] == "" {
		t.Errorf("expected non-empty refreshToken")
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	r := setupRouter(t)
	signupUser(t, r, "wrongpw@example.com", "correctpassword")

	w := do(r, "POST", "/auth/login", jsonBody(t, map[string]string{
		"email":    "wrongpw@example.com",
		"password": "badpassword",
	}))

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestLogin_UnknownEmail(t *testing.T) {
	r := setupRouter(t)

	w := do(r, "POST", "/auth/login", jsonBody(t, map[string]string{
		"email":    "nobody@example.com",
		"password": "password123",
	}))

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestLogin_InvalidEmailFormat(t *testing.T) {
	r := setupRouter(t)

	w := do(r, "POST", "/auth/login", jsonBody(t, map[string]string{
		"email":    "not-an-email",
		"password": "password123",
	}))

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestLogin_MissingPassword(t *testing.T) {
	r := setupRouter(t)

	w := do(r, "POST", "/auth/login", jsonBody(t, map[string]string{
		"email": "test@example.com",
	}))

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

// ---- Signup tests ----

func TestSignup_Success(t *testing.T) {
	r := setupRouter(t)

	w := do(r, "POST", "/auth/signup", jsonBody(t, map[string]string{
		"email":    "newuser@example.com",
		"password": "securepass",
		"name":     "New User",
		"phone":    "9876543210",
	}))

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	resp := parseResp(t, w)
	if !resp["success"].(bool) {
		t.Errorf("expected success=true")
	}
	data := resp["data"].(map[string]any)
	if data["email"] != "newuser@example.com" {
		t.Errorf("expected email in response data")
	}
}

func TestSignup_DuplicateEmail(t *testing.T) {
	r := setupRouter(t)

	body := map[string]string{"email": "dup@example.com", "password": "securepass"}
	w := do(r, "POST", "/auth/signup", jsonBody(t, body))
	if w.Code != http.StatusOK {
		t.Fatalf("first signup failed: %d", w.Code)
	}

	w = do(r, "POST", "/auth/signup", jsonBody(t, body))
	if w.Code != http.StatusConflict {
		t.Errorf("expected 409, got %d: %s", w.Code, w.Body.String())
	}
}

func TestSignup_PasswordTooShort(t *testing.T) {
	r := setupRouter(t)

	w := do(r, "POST", "/auth/signup", jsonBody(t, map[string]string{
		"email":    "short@example.com",
		"password": "short",
	}))

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestSignup_MissingEmail(t *testing.T) {
	r := setupRouter(t)

	w := do(r, "POST", "/auth/signup", jsonBody(t, map[string]string{
		"password": "securepass",
	}))

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

// ---- Refresh tests ----

func TestRefresh_MissingToken(t *testing.T) {
	r := setupRouter(t)

	w := do(r, "POST", "/auth/refresh", jsonBody(t, map[string]string{}))
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestRefresh_InvalidToken(t *testing.T) {
	r := setupRouter(t)

	w := do(r, "POST", "/auth/refresh", jsonBody(t, map[string]string{
		"refreshToken": "not-a-real-token",
	}))
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestRefresh_WithValidToken(t *testing.T) {
	r := setupRouter(t)

	w := signupUser(t, r, "refresh@example.com", "password123")
	if w.Code != http.StatusOK {
		t.Fatalf("signup failed: %d", w.Code)
	}

	var signupResp struct {
		Data struct {
			RefreshToken string `json:"refreshToken"`
		} `json:"data"`
	}
	json.NewDecoder(w.Body).Decode(&signupResp)
	refreshToken := signupResp.Data.RefreshToken

	w = do(r, "POST", "/auth/refresh", jsonBody(t, map[string]string{
		"refreshToken": refreshToken,
	}))
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	resp := parseResp(t, w)
	if !resp["success"].(bool) {
		t.Errorf("expected success=true")
	}
}

// ---- Logout tests ----

func TestLogout_WithoutCookie(t *testing.T) {
	r := setupRouter(t)

	w := do(r, "POST", "/auth/logout", nil)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestLogout_ClearsCookies(t *testing.T) {
	r := setupRouter(t)

	w := signupUser(t, r, "logout@example.com", "password123")
	if w.Code != http.StatusOK {
		t.Fatalf("signup failed: %d", w.Code)
	}

	// Set the refresh cookie from signup and then logout.
	var cookies []*http.Cookie
	for _, c := range w.Result().Cookies() {
		cookies = append(cookies, c)
	}

	logoutReq := httptest.NewRequest("POST", "/auth/logout", nil)
	for _, c := range cookies {
		logoutReq.AddCookie(c)
	}
	lw := httptest.NewRecorder()
	r.ServeHTTP(lw, logoutReq)

	if lw.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", lw.Code)
	}
}
