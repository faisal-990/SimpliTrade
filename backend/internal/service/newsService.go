package service

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/faisal-990/ProjectInvestApp/backend/internal/dto"
	"github.com/faisal-990/ProjectInvestApp/backend/internal/utils"
)

type News interface {
	GetNews() ([]dto.NewsDTO, error)
}

type news struct{}

func NewNewsService() News {
	return &news{}
}

func (n *news) GetNews() ([]dto.NewsDTO, error) {
	time.Sleep(time.Second * 2)
	cwd, err := os.Getwd()
	if err != nil {
		utils.LogError("failed to load working dir", err)
	}
	seedpath := filepath.Join(cwd, "internal", "seed", "financial_market_news.json")
	data, err := os.ReadFile(seedpath)
	if err != nil {
		utils.LogError("failed to load news from json", err)
		return nil, err
	}
	var wrapper struct {
		Feed []dto.NewsDTO `json:"feed"`
	}
	if err := json.Unmarshal(data, &wrapper); err != nil {
		utils.LogError("failed to unmarshal the json blob", err)
		return nil, err
	}
	if len(wrapper.Feed) <= 10 {
		return wrapper.Feed, nil
	}

	return wrapper.Feed[:10], nil
}
