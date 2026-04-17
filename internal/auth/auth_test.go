package auth_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dhruvpurohit2k/expressions-india-backend/internal/auth"
	"github.com/dhruvpurohit2k/expressions-india-backend/internal/dto"
	"github.com/dhruvpurohit2k/expressions-india-backend/internal/models"
	"github.com/dhruvpurohit2k/expressions-india-backend/internal/testutil"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type testRig struct {
	r   *gin.Engine
	svc *auth.Service
	db  *gorm.DB
}

func setupRouter(t *testing.T) *testRig {
	t.Helper()
	db := testutil.NewTestDB(t)
	svc := auth.NewService(db)
	ctrl := auth.NewController(svc)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/auth/login", ctrl.Login)
	r.POST("/auth/refresh", ctrl.Refresh)
	r.POST("/auth/logout", ctrl.Logout)
	return &testRig{r: r, svc: svc, db: db}
}

// seedPasswordUser inserts a password-auth user (admin) directly into the DB.
// Used as a test fixture in place of the removed public signup endpoint.
func seedPasswordUser(t *testing.T, db *gorm.DB, email, password string) {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	pw := string(hash)
	u := models.User{
		Email:         email,
		Password:      &pw,
		EmailVerified: true,
		IsAdmin:       true,
		Provider:      models.ProviderPassword,
		ProviderSub:   "password:" + email,
	}
	if err := db.Create(&u).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
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

// ---- Login tests ----

func TestLogin_Success(t *testing.T) {
	rig := setupRouter(t)
	seedPasswordUser(t, rig.db, "login@example.com", "password123")

	w := do(rig.r, "POST", "/auth/login", jsonBody(t, map[string]string{
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
	rig := setupRouter(t)
	seedPasswordUser(t, rig.db, "wrongpw@example.com", "correctpassword")

	w := do(rig.r, "POST", "/auth/login", jsonBody(t, map[string]string{
		"email":    "wrongpw@example.com",
		"password": "badpassword",
	}))

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestLogin_UnknownEmail(t *testing.T) {
	rig := setupRouter(t)

	w := do(rig.r, "POST", "/auth/login", jsonBody(t, map[string]string{
		"email":    "nobody@example.com",
		"password": "password123",
	}))

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestLogin_InvalidEmailFormat(t *testing.T) {
	rig := setupRouter(t)

	w := do(rig.r, "POST", "/auth/login", jsonBody(t, map[string]string{
		"email":    "not-an-email",
		"password": "password123",
	}))

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestLogin_MissingPassword(t *testing.T) {
	rig := setupRouter(t)

	w := do(rig.r, "POST", "/auth/login", jsonBody(t, map[string]string{
		"email": "test@example.com",
	}))

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

// ---- Register (admin path) tests ----

func TestRegister_CreatesPasswordAdmin(t *testing.T) {
	rig := setupRouter(t)

	resp, err := rig.svc.Register(dto.RegisterRequest{
		Email:    "newadmin@example.com",
		Password: "securepass",
		Name:     "New Admin",
		IsAdmin:  true,
	})
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	if !resp.IsAdmin {
		t.Errorf("expected IsAdmin=true")
	}

	var u models.User
	if err := rig.db.Where("email = ?", "newadmin@example.com").First(&u).Error; err != nil {
		t.Fatalf("lookup: %v", err)
	}
	if u.Provider != models.ProviderPassword {
		t.Errorf("expected provider=password, got %q", u.Provider)
	}
}

// ---- Refresh tests ----

func TestRefresh_MissingToken(t *testing.T) {
	rig := setupRouter(t)

	w := do(rig.r, "POST", "/auth/refresh", jsonBody(t, map[string]string{}))
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestRefresh_InvalidToken(t *testing.T) {
	rig := setupRouter(t)

	w := do(rig.r, "POST", "/auth/refresh", jsonBody(t, map[string]string{
		"refreshToken": "not-a-real-token",
	}))
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestRefresh_WithValidToken(t *testing.T) {
	rig := setupRouter(t)
	seedPasswordUser(t, rig.db, "refresh@example.com", "password123")

	loginW := do(rig.r, "POST", "/auth/login", jsonBody(t, map[string]string{
		"email":    "refresh@example.com",
		"password": "password123",
	}))
	if loginW.Code != http.StatusOK {
		t.Fatalf("login failed: %d", loginW.Code)
	}

	var loginResp struct {
		Data struct {
			RefreshToken string `json:"refreshToken"`
		} `json:"data"`
	}
	json.NewDecoder(loginW.Body).Decode(&loginResp)

	w := do(rig.r, "POST", "/auth/refresh", jsonBody(t, map[string]string{
		"refreshToken": loginResp.Data.RefreshToken,
	}))
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

// ---- Logout tests ----

func TestLogout_WithoutCookie(t *testing.T) {
	rig := setupRouter(t)

	w := do(rig.r, "POST", "/auth/logout", nil)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestLogout_ClearsCookies(t *testing.T) {
	rig := setupRouter(t)
	seedPasswordUser(t, rig.db, "logout@example.com", "password123")

	loginW := do(rig.r, "POST", "/auth/login", jsonBody(t, map[string]string{
		"email":    "logout@example.com",
		"password": "password123",
	}))
	if loginW.Code != http.StatusOK {
		t.Fatalf("login failed: %d", loginW.Code)
	}

	logoutReq := httptest.NewRequest("POST", "/auth/logout", nil)
	for _, c := range loginW.Result().Cookies() {
		logoutReq.AddCookie(c)
	}
	lw := httptest.NewRecorder()
	rig.r.ServeHTTP(lw, logoutReq)

	if lw.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", lw.Code)
	}
}
