package dto

type NewsDTO struct {
	Title    string `json:"title"`
	NewsUrl  string `json:"url"`
	Authors  string `json:"authors"`
	Summary  string `json:"summary"`
	Source   string `json:"source"`
	ImageUrl string `json:"banner_image"`
}
