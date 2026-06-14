package service

import (
	"context"
	"errors"

	"github.com/faisal-990/ProjectInvestApp/internal/platform/repository"
	"github.com/faisal-990/ProjectInvestApp/internal/web/dto"
	"github.com/faisal-990/ProjectInvestApp/internal/web/httpx"
)

// chartCandles is how many recent bars a stock detail returns for its chart.
const chartCandles = 90

// DashboardService serves the read-only market views the frontend renders. It
// reads the universe + price history kept fresh by the engine (Tower 2) — it
// never calls an external API itself.
type DashboardService interface {
	Fundamentals(ctx context.Context, limit, offset int) ([]dto.StockSummaryDTO, error)
	StockDetail(ctx context.Context, symbol string) (*dto.StockDetailDTO, error)
}

type dashboardService struct {
	stocks repository.StockRepo
}

// NewDashboardService reuses the StockRepo — the dashboard is a read view over
// the same stock data the engine maintains.
func NewDashboardService(stocks repository.StockRepo) DashboardService {
	return &dashboardService{stocks: stocks}
}

func (s *dashboardService) Fundamentals(ctx context.Context, limit, offset int) ([]dto.StockSummaryDTO, error) {
	stocks, err := s.stocks.List(ctx, limit, offset)
	if err != nil {
		return nil, httpx.Internal("could not load stocks").WithCause(err)
	}
	out := make([]dto.StockSummaryDTO, 0, len(stocks))
	for _, st := range stocks {
		out = append(out, dto.StockSummaryDTO{
			Symbol: st.Symbol, Name: st.Name, Sector: st.Sector,
			AssetClass: st.AssetClass, CurrentPrice: st.CurrentPrice,
			Fundamentals: st.Fundamentals,
		})
	}
	return out, nil
}

func (s *dashboardService) StockDetail(ctx context.Context, symbol string) (*dto.StockDetailDTO, error) {
	stock, err := s.stocks.GetBySymbol(ctx, symbol)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, httpx.NotFound("unknown symbol")
		}
		return nil, httpx.Internal("could not load stock").WithCause(err)
	}

	candles, err := s.stocks.GetCandles(ctx, symbol, "1d", chartCandles)
	if err != nil {
		return nil, httpx.Internal("could not load price history").WithCause(err)
	}
	// GetCandles returns newest-first; present oldest-first for charting.
	bars := make([]dto.CandleDTO, len(candles))
	for i, c := range candles {
		bars[len(candles)-1-i] = dto.CandleDTO{
			Time: c.Timestamp.Unix(), Open: c.Open, High: c.High,
			Low: c.Low, Close: c.Close, Volume: c.Volume,
		}
	}

	return &dto.StockDetailDTO{
		Symbol: stock.Symbol, Name: stock.Name, Sector: stock.Sector,
		Exchange: stock.Exchange, AssetClass: stock.AssetClass,
		CurrentPrice: stock.CurrentPrice, Fundamentals: stock.Fundamentals,
		Candles: bars,
	}, nil
}
