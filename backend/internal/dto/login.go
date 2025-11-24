package dto

type Login struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type Signup struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}
