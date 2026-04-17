package auth

import (
	"net/http"
	"os"

	"github.com/dhruvpurohit2k/expressions-india-backend/internal/dto"
	"github.com/dhruvpurohit2k/expressions-india-backend/internal/pkg/utils"
	"github.com/gin-gonic/gin"
)

const (
	cookieAccess  = "ei_access"
	cookieRefresh = "ei_refresh"
)

func isProd() bool { return os.Getenv("APP_ENV") == "production" }

type Controller struct {
	service *Service
}

func NewController(service *Service) *Controller {
	return &Controller{service: service}
}

func sameSite() http.SameSite {
	switch os.Getenv("COOKIE_SAMESITE") {
	case "none":
		return http.SameSiteNoneMode
	case "strict":
		return http.SameSiteStrictMode
	default:
		return http.SameSiteLaxMode
	}
}

func (ctrl *Controller) setTokenCookies(c *gin.Context, resp *dto.AuthResponse) {
	secure := isProd()
	c.SetSameSite(sameSite())
	c.SetCookie(cookieAccess, resp.AccessToken, 15*60, "/", "", secure, true)
	c.SetCookie(cookieRefresh, resp.RefreshToken, 7*24*60*60, "/auth", "", secure, true)
}

func (ctrl *Controller) clearTokenCookies(c *gin.Context) {
	secure := isProd()
	c.SetSameSite(sameSite())
	c.SetCookie(cookieAccess, "", -1, "/", "", secure, true)
	c.SetCookie(cookieRefresh, "", -1, "/auth", "", secure, true)
}

// Login is the password sign-in path used by admins.
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

// Google is the public OIDC sign-in path for end users.
func (ctrl *Controller) Google(c *gin.Context) {
	var req dto.OIDCLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		return
	}
	resp, err := ctrl.service.LoginWithGoogle(req)
	if err != nil {
		utils.Fail(c, StatusCode(err), "AUTH_ERROR", err.Error())
		return
	}
	ctrl.setTokenCookies(c, resp)
	utils.OK(c, resp)
}

// Apple is the public OIDC sign-in path for end users on iOS.
func (ctrl *Controller) Apple(c *gin.Context) {
	var req dto.OIDCLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		return
	}
	resp, err := ctrl.service.LoginWithApple(req)
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
		_ = ctrl.service.Logout(raw)
	}
	ctrl.clearTokenCookies(c)
	utils.OK(c, gin.H{"message": "logged out"})
}

// Register is admin-only — protected by RequireAdmin middleware in server.go.
// Used to create additional password-auth admins.
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
