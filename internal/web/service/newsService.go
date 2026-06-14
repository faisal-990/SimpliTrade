package service

import (
	"encoding/json"
	"fmt"

	"github.com/faisal-990/ProjectInvestApp/internal/web/dto"
	"github.com/faisal-990/ProjectInvestApp/internal/web/seed"
)

// maxNewsItems caps how many headlines the feed endpoint returns.
const maxNewsItems = 10

type News interface {
	GetNews() ([]dto.NewsDTO, error)
}

type news struct{}

func NewNewsService() News { return &news{} }

// GetNews returns the latest market headlines from the embedded seed feed (a
// live provider replaces this later). Reading embedded bytes means it works from
// any working directory and in tests — the previous os.Getwd path was wrong
// (it pointed at internal/seed, not internal/web/seed) and CWD-dependent.
func (n *news) GetNews() ([]dto.NewsDTO, error) {
	var wrapper struct {
		Feed []dto.NewsDTO `json:"feed"`
	}
	if err := json.Unmarshal(seed.NewsJSON, &wrapper); err != nil {
		return nil, fmt.Errorf("news: parsing seed feed: %w", err)
	}
	if len(wrapper.Feed) > maxNewsItems {
		return wrapper.Feed[:maxNewsItems], nil
	}
	return wrapper.Feed, nil
}
