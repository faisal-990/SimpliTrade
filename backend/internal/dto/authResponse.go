package dto

type AuthResponse struct {
	Token string `json:"token"`
	Name  string `json:"name"`
	Email string `json:"email"`
	ID    string `json:"id"`
}
