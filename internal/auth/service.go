package auth

import (
	"errors"
	"net/http"
	"time"

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

var ErrInvalidCredentials = errors.New("invalid email or password")
var ErrTokenExpiredOrInvalid = errors.New("refresh token expired or invalid")

// Signup creates a non-admin user from a public self-registration request.
func (s *Service) Signup(req dto.SignupRequest) (*dto.AuthResponse, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	user := models.User{
		Email:    req.Email,
		Password: string(hash),
		Name:     req.Name,
		Phone:    req.Phone,
		IsAdmin:  false,
	}
	if err := s.db.Create(&user).Error; err != nil {
		return nil, err
	}
	return s.issueTokens(&user)
}

func (s *Service) Register(req dto.RegisterRequest) (*dto.AuthResponse, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	user := models.User{
		Email:    req.Email,
		Password: string(hash),
		Name:     req.Name,
		Phone:    req.Phone,
		IsAdmin:  req.IsAdmin,
	}
	if err := s.db.Create(&user).Error; err != nil {
		return nil, err
	}
	return s.issueTokens(&user)
}

func (s *Service) Login(req dto.LoginRequest) (*dto.AuthResponse, error) {
	var user models.User
	if err := s.db.Where("email = ?", req.Email).First(&user).Error; err != nil {
		return nil, ErrInvalidCredentials
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, ErrInvalidCredentials
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

// StatusCode returns the appropriate HTTP status for known auth errors.
func StatusCode(err error) int {
	switch {
	case errors.Is(err, ErrInvalidCredentials):
		return http.StatusUnauthorized
	case errors.Is(err, ErrTokenExpiredOrInvalid):
		return http.StatusUnauthorized
	default:
		return http.StatusInternalServerError
	}
}
