package dto

// SignupRequest is the body for POST /auth/signup. Binding tags enforce input
// validation at the controller edge before any business logic runs.
type SignupRequest struct {
	Name     string `json:"name" binding:"required,min=2,max=100"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8,max=72"` // 72 = bcrypt limit
}

// LoginRequest is the body for POST /auth/login.
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// RefreshRequest is the body for POST /auth/refresh and /auth/logout.
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// ForgotPasswordRequest is the body for POST /auth/forgot-password.
type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// ResetPasswordRequest is the body for POST /auth/reset-password.
type ResetPasswordRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Code     string `json:"code" binding:"required,len=6"`            // 6-digit OTP
	Password string `json:"password" binding:"required,min=8,max=72"` // 72 = bcrypt limit
}

// UserDTO is the client-safe projection of a user (never includes the password
// hash).
type UserDTO struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Email         string `json:"email"`
	Role          string `json:"role"`
	EmailVerified bool   `json:"email_verified"`
	AvatarURL     string `json:"avatar_url"`
	Bio           string `json:"bio"`
}

// UpdateProfileRequest is the body for PUT /auth/me (edit the "About me" section).
type UpdateProfileRequest struct {
	Name string `json:"name" binding:"required,min=2,max=100"`
	Bio  string `json:"bio" binding:"max=500"`
}

// AuthResponse is returned by signup, login, and refresh.
type AuthResponse struct {
	AccessToken  string  `json:"access_token"`
	RefreshToken string  `json:"refresh_token"`
	ExpiresAt    int64   `json:"expires_at"` // unix seconds — access token expiry
	User         UserDTO `json:"user"`
}
