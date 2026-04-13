package dto

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	Name     string `json:"name"`
	Phone    string `json:"phone"`
	IsAdmin  bool   `json:"isAdmin"`
}

// SignupRequest is the public self-registration payload — no isAdmin field.
type SignupRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	Name     string `json:"name"`
	Phone    string `json:"phone"`
}

type AuthResponse struct {
	AccessToken  string `json:"accessToken"` // returned in body so mobile clients can store it
	RefreshToken string `json:"-"`
	UserID       string `json:"userId"`
	Email        string `json:"email"`
	Name         string `json:"name"`
	Phone        string `json:"phone"`
	IsAdmin      bool   `json:"isAdmin"`
}
