package dto

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// RegisterRequest is admin-only — used by an existing admin to create another
// password-auth admin/user. Public OIDC signups go through /auth/google or
// /auth/apple; there is no public password signup.
type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	Name     string `json:"name"`
	Phone    string `json:"phone"`
	IsAdmin  bool   `json:"isAdmin"`
}

// OIDCLoginRequest is the payload for /auth/google and /auth/apple. The Name
// field is only honored on first sign-in for Apple (which omits name on
// subsequent sign-ins) — clients should pass it through when available.
type OIDCLoginRequest struct {
	IDToken string `json:"idToken" binding:"required"`
	Name    string `json:"name"`
}

type AuthResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"` // also set as HttpOnly cookie; returned in body for mobile clients
	UserID       string `json:"userId"`
	Email        string `json:"email"`
	Name         string `json:"name"`
	Phone        string `json:"phone"`
	IsAdmin      bool   `json:"isAdmin"`
}
