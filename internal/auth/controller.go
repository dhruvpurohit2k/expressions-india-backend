package auth

import (
	"net/http"
	"strings"

	"github.com/dhruvpurohit2k/expressions-india-backend/internal/dto"
	"github.com/dhruvpurohit2k/expressions-india-backend/internal/pkg/utils"
	"github.com/gin-gonic/gin"
)

const (
	cookieAccess  = "ei_access"
	cookieRefresh = "ei_refresh"
)

type Controller struct {
	service *Service
}

func NewController(service *Service) *Controller {
	return &Controller{service: service}
}

func (ctrl *Controller) setTokenCookies(c *gin.Context, resp *dto.AuthResponse) {
	secure := gin.Mode() == gin.ReleaseMode
	c.SetCookie(cookieAccess, resp.AccessToken, 15*60, "/", "", secure, true)
	c.SetCookie(cookieRefresh, resp.RefreshToken, 7*24*60*60, "/auth", "", secure, true)
}

func (ctrl *Controller) clearTokenCookies(c *gin.Context) {
	secure := gin.Mode() == gin.ReleaseMode
	c.SetCookie(cookieAccess, "", -1, "/", "", secure, true)
	c.SetCookie(cookieRefresh, "", -1, "/auth", "", secure, true)
}

func (ctrl *Controller) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		return
	}
	resp, err := ctrl.service.Login(req)
	if err != nil {
		utils.Fail(c, StatusCode(err), "AUTH_ERROR", err.Error())
		return
	}
	ctrl.setTokenCookies(c, resp)
	utils.OK(c, resp)
}

func (ctrl *Controller) Refresh(c *gin.Context) {
	// Web clients send the refresh token as an HttpOnly cookie.
	// Mobile clients (no cookie support) send it in the JSON body.
	var raw string
	if cookie, err := c.Cookie(cookieRefresh); err == nil && cookie != "" {
		raw = cookie
	}
	if raw == "" {
		var body struct {
			RefreshToken string `json:"refreshToken"`
		}
		if err := c.ShouldBindJSON(&body); err == nil {
			raw = body.RefreshToken
		}
	}
	if raw == "" {
		utils.Fail(c, http.StatusUnauthorized, "AUTH_ERROR", "missing refresh token")
		return
	}
	resp, err := ctrl.service.Refresh(raw)
	if err != nil {
		ctrl.clearTokenCookies(c)
		utils.Fail(c, StatusCode(err), "AUTH_ERROR", err.Error())
		return
	}
	ctrl.setTokenCookies(c, resp)
	utils.OK(c, resp)
}

func (ctrl *Controller) Logout(c *gin.Context) {
	raw, _ := c.Cookie(cookieRefresh)
	if raw != "" {
		// Best-effort: invalidate the refresh token in the DB.
		_ = ctrl.service.Logout(raw)
	}
	ctrl.clearTokenCookies(c)
	utils.OK(c, gin.H{"message": "logged out"})
}

// Signup is the public self-registration endpoint — anyone can create an account.
func (ctrl *Controller) Signup(c *gin.Context) {
	var req dto.SignupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		return
	}
	resp, err := ctrl.service.Signup(req)
	if err != nil {
		msg := err.Error()
		if strings.Contains(msg, "UNIQUE") || strings.Contains(msg, "unique") || strings.Contains(msg, "duplicate") {
			utils.Fail(c, http.StatusConflict, "SIGNUP_ERROR", "an account with this email already exists")
		} else {
			utils.FailInternal(c, "SIGNUP_ERROR", "could not create account", err)
		}
		return
	}
	utils.OK(c, resp)
}

// Register is admin-only — protected by RequireAdmin middleware in server.go.
func (ctrl *Controller) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		return
	}
	resp, err := ctrl.service.Register(req)
	if err != nil {
		utils.FailInternal(c, "REGISTER_ERROR", "could not register user", err)
		return
	}
	utils.OK(c, resp)
}
