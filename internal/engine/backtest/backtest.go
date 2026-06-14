// Package backtest replays a strategy over historical price data to show how it
// would have performed. It reuses the same pure decide.Decide core the live
// engine uses — so a backtest and a live run make identical decisions given the
// same inputs — and executes the resulting intents against an in-memory ledger.
// It performs no I/O: callers load the history and hand it in.
//
// Limitation: fundamentals are point-in-time (the current snapshot), because the
// platform stores only the latest fundamentals, not a historical series. Price
// and technical signals are fully historical; fundamental screens use today's
// values. Callers should surface this caveat to users.
package backtest

import (
	"sort"
	"time"

	"github.com/faisal-990/ProjectInvestApp/internal/engine/decide"
	"github.com/faisal-990/ProjectInvestApp/internal/engine/strategy"
	"github.com/faisal-990/ProjectInvestApp/internal/platform/models"
)

// Bar is one historical close for a symbol.
type Bar struct {
	Date  time.Time
	Close float64
}

// SymbolHistory is a stock's metadata plus its close series (oldest→newest).
type SymbolHistory struct {
	Symbol       string
	Sector       string
	AssetClass   string
	Fundamentals models.Fundamentals
	Bars         []Bar
}

// Params configures a run.
type Params struct {
	StartCash float64
}

// EquityPoint is the portfolio's total value at a point in time.
type EquityPoint struct {
	Date  time.Time `json:"date"`
	Value float64   `json:"value"`
}

// TradeLog is one executed order in the replay.
type TradeLog struct {
	Date       time.Time `json:"date"`
	Side       string    `json:"side"`
	Symbol     string    `json:"symbol"`
	Quantity   float64   `json:"quantity"`
	Price      float64   `json:"price"`
	TotalValue float64   `json:"total_value"`
}

// HoldingOut is a position the strategy is still holding at the end of the run —
// i.e. names it chose to keep ("hold") rather than sell.
type HoldingOut struct {
	Symbol       string  `json:"symbol"`
	Quantity     float64 `json:"quantity"`
	AvgPrice     float64 `json:"avg_price"`
	LastPrice    float64 `json:"last_price"`
	MarketValue  float64 `json:"market_value"`
	UnrealizedPL float64 `json:"unrealized_pl"`
}

// Result is the outcome of a backtest.
type Result struct {
	StartCash   float64
	FinalValue  float64
	ROI         float64
	MaxDrawdown float64
	TradeCount  int
	BuyCount    int
	SellCount   int
	WinRate     float64 // share of closing sells that realized a profit
	StartDate   time.Time
	EndDate     time.Time
	Equity      []EquityPoint
	Trades      []TradeLog
	Holdings    []HoldingOut // positions still held at the end (the "hold" decision)
	EndCash     float64
}

// position tracks held quantity and average cost for realized-P&L accounting.
type position struct {
	qty float64
	avg float64
}

// series is one symbol's de-duplicated, date-ascending close history.
type series struct {
	meta   SymbolHistory
	dates  []int64   // unix-day keys, ascending
	closes []float64 // aligned with dates
}

// asOf returns the close as of the given day (carrying forward the latest prior
// close) and the close series up to and including that day. ok is false when the
// symbol has no bar on or before the day.
func (s *series) asOf(day int64) (price float64, closes []float64, ok bool) {
	// Largest index whose date <= day.
	i := sort.Search(len(s.dates), func(i int) bool { return s.dates[i] > day }) - 1
	if i < 0 {
		return 0, nil, false
	}
	return s.closes[i], s.closes[: i+1], true
}

// Run replays cfg day-by-day over the supplied history and returns performance.
func Run(cfg strategy.Config, hist []SymbolHistory, p Params) Result {
	if p.StartCash <= 0 {
		p.StartCash = models.StartingSimBalance
	}

	bySymbol := make(map[string]*series, len(hist))
	dateSet := make(map[int64]time.Time)
	for _, h := range hist {
		s := &series{meta: h}
		seen := make(map[int64]struct{}, len(h.Bars))
		// Bars are expected oldest→newest; sort defensively.
		bars := append([]Bar(nil), h.Bars...)
		sort.Slice(bars, func(i, j int) bool { return bars[i].Date.Before(bars[j].Date) })
		for _, b := range bars {
			day := dayKey(b.Date)
			if _, dup := seen[day]; dup {
				continue
			}
			seen[day] = struct{}{}
			s.dates = append(s.dates, day)
			s.closes = append(s.closes, b.Close)
			dateSet[day] = time.Unix(day, 0).UTC()
		}
		if len(s.closes) > 0 {
			bySymbol[h.Symbol] = s
		}
	}

	timeline := make([]time.Time, 0, len(dateSet))
	for _, t := range dateSet {
		timeline = append(timeline, t)
	}
	sort.Slice(timeline, func(i, j int) bool { return timeline[i].Before(timeline[j]) })
	if len(timeline) == 0 {
		return Result{StartCash: p.StartCash, FinalValue: p.StartCash}
	}

	cash := p.StartCash
	positions := make(map[string]*position)
	res := Result{
		StartCash: p.StartCash,
		StartDate: timeline[0],
		EndDate:   timeline[len(timeline)-1],
		Equity:    make([]EquityPoint, 0, len(timeline)),
	}
	var sells, wins, buys int
	peak := p.StartCash
	lastPrices := map[string]float64{}

	for _, d := range timeline {
		day := dayKey(d)

		market := make([]decide.StockView, 0, len(bySymbol))
		priceNow := make(map[string]float64, len(bySymbol))
		for sym, s := range bySymbol {
			price, closes, ok := s.asOf(day)
			if !ok || price <= 0 {
				continue
			}
			priceNow[sym] = price
			market = append(market, decide.StockView{
				Symbol:       s.meta.Symbol,
				Sector:       s.meta.Sector,
				AssetClass:   s.meta.AssetClass,
				Price:        price,
				Fundamentals: s.meta.Fundamentals,
				Closes:       closes,
			})
		}

		pf := snapshotPortfolio(cash, positions)
		for _, in := range dedupe(decide.Decide(cfg, market, pf)) {
			price, ok := priceNow[in.Symbol]
			if !ok || price <= 0 || in.Quantity <= 0 {
				continue
			}
			switch in.Action {
			case decide.Sell:
				pos := positions[in.Symbol]
				if pos == nil || pos.qty <= 0 {
					continue
				}
				qty := in.Quantity
				if qty > pos.qty {
					qty = pos.qty
				}
				cash += qty * price
				if price > pos.avg {
					wins++
				}
				sells++
				pos.qty -= qty
				if pos.qty <= 1e-9 {
					delete(positions, in.Symbol)
				}
				res.Trades = append(res.Trades, TradeLog{Date: d, Side: "sell", Symbol: in.Symbol, Quantity: qty, Price: price, TotalValue: qty * price})
			case decide.Buy:
				qty := in.Quantity
				cost := qty * price
				if cost > cash { // no margin; size down to affordable
					qty = cash / price
					cost = qty * price
				}
				if qty <= 1e-9 {
					continue
				}
				pos := positions[in.Symbol]
				if pos == nil {
					pos = &position{}
					positions[in.Symbol] = pos
				}
				total := pos.qty + qty
				pos.avg = (pos.avg*pos.qty + price*qty) / total
				pos.qty = total
				cash -= cost
				buys++
				res.Trades = append(res.Trades, TradeLog{Date: d, Side: "buy", Symbol: in.Symbol, Quantity: qty, Price: price, TotalValue: cost})
			}
		}

		equity := cash
		for sym, pos := range positions {
			equity += pos.qty * priceNow[sym]
		}
		lastPrices = priceNow
		if equity > peak {
			peak = equity
		}
		if peak > 0 {
			if dd := (peak - equity) / peak; dd > res.MaxDrawdown {
				res.MaxDrawdown = dd
			}
		}
		res.Equity = append(res.Equity, EquityPoint{Date: d, Value: equity})
	}

	last := res.Equity[len(res.Equity)-1].Value
	res.FinalValue = last
	res.ROI = last/p.StartCash - 1
	res.TradeCount = len(res.Trades)
	res.BuyCount = buys
	res.SellCount = sells
	res.EndCash = cash
	if sells > 0 {
		res.WinRate = float64(wins) / float64(sells)
	}

	// Final positions = the names the strategy chose to HOLD through to the end.
	res.Holdings = make([]HoldingOut, 0, len(positions))
	for sym, pos := range positions {
		if pos.qty <= 1e-9 {
			continue
		}
		lp := lastPrices[sym]
		res.Holdings = append(res.Holdings, HoldingOut{
			Symbol: sym, Quantity: pos.qty, AvgPrice: pos.avg, LastPrice: lp,
			MarketValue: pos.qty * lp, UnrealizedPL: (lp - pos.avg) * pos.qty,
		})
	}
	sort.Slice(res.Holdings, func(i, j int) bool { return res.Holdings[i].MarketValue > res.Holdings[j].MarketValue })
	return res
}

func dayKey(t time.Time) int64 {
	y, m, d := t.UTC().Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC).Unix()
}

// snapshotPortfolio adapts the in-memory ledger to the decide.Portfolio shape.
func snapshotPortfolio(cash float64, positions map[string]*position) decide.Portfolio {
	out := decide.Portfolio{Cash: cash, Positions: make(map[string]decide.Position, len(positions))}
	for sym, pos := range positions {
		out.Positions[sym] = decide.Position{Quantity: pos.qty, AvgPrice: pos.avg}
	}
	return out
}

// dedupe keeps at most one intent per symbol; a Sell wins over a Buy.
func dedupe(intents []decide.Intent) []decide.Intent {
	bySym := make(map[string]decide.Intent, len(intents))
	for _, in := range intents {
		if ex, ok := bySym[in.Symbol]; ok {
			if ex.Action == decide.Sell || in.Action != decide.Sell {
				continue
			}
		}
		bySym[in.Symbol] = in
	}
	out := make([]decide.Intent, 0, len(bySym))
	for _, in := range bySym {
		out = append(out, in)
	}
	return out
}
