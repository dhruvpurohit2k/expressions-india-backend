package auth

import (
	"errors"
	"net/http"
	"time"

	"github.com/dhruvpurohit2k/expressions-india-backend/internal/auth/oidc"
	"github.com/dhruvpurohit2k/expressions-india-backend/internal/dto"
	"github.com/dhruvpurohit2k/expressions-india-backend/internal/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type Service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

var (
	ErrInvalidCredentials    = errors.New("invalid email or password")
	ErrTokenExpiredOrInvalid = errors.New("refresh token expired or invalid")
	ErrAdminLoginRequired    = errors.New("this account must sign in with its identity provider")
	ErrOIDCRejected          = errors.New("identity provider rejected the token")
	ErrOIDCAdminConflict     = errors.New("an admin account already exists for this email")
)

// Register creates a password-auth account. Admin-only — used for seeding
// other admins. End users sign up via OIDC.
func (s *Service) Register(req dto.RegisterRequest) (*dto.AuthResponse, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	pw := string(hash)
	user := models.User{
		Email:         req.Email,
		Password:      &pw,
		EmailVerified: true,
		Name:          req.Name,
		Phone:         req.Phone,
		IsAdmin:       req.IsAdmin,
		Provider:      models.ProviderPassword,
		ProviderSub:   "password:" + req.Email,
	}
	if err := s.db.Create(&user).Error; err != nil {
		return nil, err
	}
	return s.issueTokens(&user)
}

// Login is the password sign-in path. Restricted to password-provider accounts
// (i.e. admins). OIDC users must use the provider-specific endpoint.
func (s *Service) Login(req dto.LoginRequest) (*dto.AuthResponse, error) {
	var user models.User
	if err := s.db.Where("email = ? AND provider = ?", req.Email, models.ProviderPassword).First(&user).Error; err != nil {
		return nil, ErrInvalidCredentials
	}
	if user.Password == nil {
		return nil, ErrInvalidCredentials
	}
	if err := bcrypt.CompareHashAndPassword([]byte(*user.Password), []byte(req.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}
	return s.issueTokens(&user)
}

// LoginWithGoogle verifies a Google ID token, upserts the user, and issues
// app tokens. Never grants admin — admins use password login.
func (s *Service) LoginWithGoogle(req dto.OIDCLoginRequest) (*dto.AuthResponse, error) {
	v, err := oidc.GoogleVerifier()
	if err != nil {
		return nil, err
	}
	claims, err := v.Verify(req.IDToken)
	if err != nil {
		return nil, ErrOIDCRejected
	}
	return s.upsertOIDCUser(models.ProviderGoogle, claims, req.Name)
}

// LoginWithApple verifies an Apple ID token. Apple omits name on subsequent
// sign-ins, so the client should pass it through on first contact.
func (s *Service) LoginWithApple(req dto.OIDCLoginRequest) (*dto.AuthResponse, error) {
	v, err := oidc.AppleVerifier()
	if err != nil {
		return nil, err
	}
	claims, err := v.Verify(req.IDToken)
	if err != nil {
		return nil, ErrOIDCRejected
	}
	return s.upsertOIDCUser(models.ProviderApple, claims, req.Name)
}

func (s *Service) upsertOIDCUser(provider string, claims *oidc.IDClaims, fallbackName string) (*dto.AuthResponse, error) {
	var user models.User
	err := s.db.Where("provider = ? AND provider_sub = ?", provider, claims.Sub).First(&user).Error
	if err == nil {
		// Existing OIDC user — refresh email/verified status if they changed.
		updates := map[string]any{}
		if claims.Email != "" && claims.Email != user.Email {
			updates["email"] = claims.Email
		}
		if v := claims.EmailIsVerified(); v != user.EmailVerified {
			updates["email_verified"] = v
		}
		if len(updates) > 0 {
			if err := s.db.Model(&user).Updates(updates).Error; err != nil {
				return nil, err
			}
		}
		return s.issueTokens(&user)
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// Brand-new account. If an admin (password) account exists for this email,
	// refuse — the admin must sign in with their password instead of accidentally
	// creating a parallel non-admin OIDC account.
	if claims.Email != "" {
		var existing int64
		if err := s.db.Model(&models.User{}).
			Where("email = ? AND is_admin = ?", claims.Email, true).
			Count(&existing).Error; err != nil {
			return nil, err
		}
		if existing > 0 {
			return nil, ErrOIDCAdminConflict
		}
	}

	name := claims.Name
	if name == "" {
		name = fallbackName
	}
	user = models.User{
		Email:         claims.Email,
		EmailVerified: claims.EmailIsVerified(),
		Name:          name,
		IsAdmin:       false,
		Provider:      provider,
		ProviderSub:   claims.Sub,
	}
	if err := s.db.Create(&user).Error; err != nil {
		return nil, err
	}
	return s.issueTokens(&user)
}

func (s *Service) Refresh(raw string) (*dto.AuthResponse, error) {
	hashed := HashRefreshToken(raw)
	var user models.User
	if err := s.db.Where("refresh_token_hash = ?", hashed).First(&user).Error; err != nil {
		return nil, ErrTokenExpiredOrInvalid
	}
	if user.RefreshTokenExpiry == nil || time.Now().After(*user.RefreshTokenExpiry) {
		return nil, ErrTokenExpiredOrInvalid
	}
	return s.issueTokens(&user)
}

func (s *Service) Logout(raw string) error {
	hashed := HashRefreshToken(raw)
	return s.db.Model(&models.User{}).
		Where("refresh_token_hash = ?", hashed).
		Updates(map[string]any{
			"refresh_token_hash":   "",
			"refresh_token_expiry": nil,
		}).Error
}

func (s *Service) issueTokens(user *models.User) (*dto.AuthResponse, error) {
	accessToken, err := GenerateAccessToken(user.ID, user.Email, user.IsAdmin)
	if err != nil {
		return nil, err
	}
	raw, hashed, err := GenerateRefreshToken()
	if err != nil {
		return nil, err
	}
	expiry := time.Now().Add(refreshTokenTTL)
	if err := s.db.Model(user).Updates(map[string]any{
		"refresh_token_hash":   hashed,
		"refresh_token_expiry": expiry,
	}).Error; err != nil {
		return nil, err
	}
	return &dto.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: raw,
		UserID:       user.ID,
		Email:        user.Email,
		Name:         user.Name,
		Phone:        user.Phone,
		IsAdmin:      user.IsAdmin,
	}, nil
}

// StatusCode maps known auth errors to HTTP statuses.
func StatusCode(err error) int {
	switch {
	case errors.Is(err, ErrInvalidCredentials),
		errors.Is(err, ErrTokenExpiredOrInvalid),
		errors.Is(err, ErrOIDCRejected):
		return http.StatusUnauthorized
	case errors.Is(err, ErrAdminLoginRequired),
		errors.Is(err, ErrOIDCAdminConflict):
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}
