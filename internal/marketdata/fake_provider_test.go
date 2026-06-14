package marketdata

import (
	"context"
	"testing"
	"time"
)

func fixedClock(t time.Time) func() time.Time { return func() time.Time { return t } }

func TestFakeProvider_FundamentalsDeterministic(t *testing.T) {
	p := NewFakeProvider()
	ctx := context.Background()

	a, err := p.Fundamentals(ctx, "AAPL")
	if err != nil {
		t.Fatalf("Fundamentals: %v", err)
	}
	b, _ := p.Fundamentals(ctx, "AAPL")
	if a != b {
		t.Errorf("fundamentals not deterministic for the same symbol:\n a=%+v\n b=%+v", a, b)
	}

	other, _ := p.Fundamentals(ctx, "MSFT")
	if other == a {
		t.Error("different symbols should produce different fundamentals")
	}
}

func TestFakeProvider_FundamentalsInSaneRanges(t *testing.T) {
	p := NewFakeProvider()
	ctx := context.Background()
	for _, sym := range Symbols() {
		f, err := p.Fundamentals(ctx, sym)
		if err != nil {
			t.Fatalf("%s: %v", sym, err)
		}
		switch {
		case f.PE <= 0 || f.PE > 200:
			t.Errorf("%s PE out of range: %v", sym, f.PE)
		case f.PB <= 0:
			t.Errorf("%s PB must be positive: %v", sym, f.PB)
		case f.GrossMargin < 0 || f.GrossMargin > 1:
			t.Errorf("%s gross margin must be a fraction: %v", sym, f.GrossMargin)
		case f.ROE < 0 || f.ROE > 1:
			t.Errorf("%s ROE must be a fraction: %v", sym, f.ROE)
		case f.MarketCap <= 0:
			t.Errorf("%s market cap must be positive: %v", sym, f.MarketCap)
		case f.Beta <= 0:
			t.Errorf("%s beta must be positive: %v", sym, f.Beta)
		}
	}
}

func TestFakeProvider_QuoteDeterministicForFixedClock(t *testing.T) {
	clock := fixedClock(time.Date(2026, 1, 2, 15, 0, 0, 0, time.UTC))
	p := NewFakeProvider().WithClock(clock)
	ctx := context.Background()

	q1, err := p.Quote(ctx, "AAPL")
	if err != nil {
		t.Fatalf("Quote: %v", err)
	}
	q2, _ := p.Quote(ctx, "AAPL")
	if q1 != q2 {
		t.Errorf("quote not deterministic for fixed clock: %+v vs %+v", q1, q2)
	}
	if q1.Price <= 0 {
		t.Errorf("price must be positive, got %v", q1.Price)
	}
	if q1.High < q1.Price || q1.Low > q1.Price {
		t.Errorf("price %v outside [low %v, high %v]", q1.Price, q1.Low, q1.High)
	}
}

func TestFakeProvider_QuoteMovesOverTime(t *testing.T) {
	p := NewFakeProvider()
	ctx := context.Background()

	early := p.WithClock(fixedClock(time.Date(2026, 1, 2, 9, 30, 0, 0, time.UTC)))
	q1, _ := early.Quote(ctx, "AAPL")
	late := p.WithClock(fixedClock(time.Date(2026, 1, 2, 12, 30, 0, 0, time.UTC)))
	q2, _ := late.Quote(ctx, "AAPL")

	if q1.Price == q2.Price {
		t.Error("price should differ across distant times (market should move)")
	}
}

func TestFakeProvider_BatchQuotesCoversAll(t *testing.T) {
	p := NewFakeProvider()
	syms := []string{"AAPL", "MSFT", "NVDA"}
	quotes, err := p.BatchQuotes(context.Background(), syms)
	if err != nil {
		t.Fatalf("BatchQuotes: %v", err)
	}
	if len(quotes) != len(syms) {
		t.Fatalf("got %d quotes, want %d", len(quotes), len(syms))
	}
	for _, s := range syms {
		if quotes[s].Symbol != s {
			t.Errorf("missing or mislabeled quote for %s", s)
		}
	}
}

func TestFakeProvider_CandlesIntegrity(t *testing.T) {
	p := NewFakeProvider()
	from := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	to := from.AddDate(0, 0, 10)

	candles, err := p.Candles(context.Background(), "AAPL", "1d", from, to)
	if err != nil {
		t.Fatalf("Candles: %v", err)
	}
	if len(candles) != 10 {
		t.Fatalf("got %d candles, want 10", len(candles))
	}
	for i, c := range candles {
		if c.High < c.Low {
			t.Errorf("candle %d: high %v < low %v", i, c.High, c.Low)
		}
		if c.High < c.Open || c.High < c.Close {
			t.Errorf("candle %d: high %v below open/close", i, c.High)
		}
		if c.Low > c.Open || c.Low > c.Close {
			t.Errorf("candle %d: low %v above open/close", i, c.Low)
		}
		if c.Interval != "1d" {
			t.Errorf("candle %d: interval = %q, want 1d", i, c.Interval)
		}
	}

	// Empty/invalid range yields no candles.
	if got, _ := p.Candles(context.Background(), "AAPL", "1d", to, from); got != nil {
		t.Errorf("inverted range should yield no candles, got %d", len(got))
	}
}
