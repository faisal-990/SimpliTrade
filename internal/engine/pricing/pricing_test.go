package pricing

import (
	"math"
	"testing"
)

func approx(a, b float64) bool { return math.Abs(a-b) < 1e-6 }

func TestGrahamNumber(t *testing.T) {
	// √(22.5 · 4 · 25) = √2250 ≈ 47.434
	got := GrahamNumber(4, 25)
	if !approx(got, math.Sqrt(2250)) {
		t.Errorf("GrahamNumber(4,25) = %v, want %v", got, math.Sqrt(2250))
	}
	if GrahamNumber(-1, 25) != 0 || GrahamNumber(4, 0) != 0 {
		t.Error("non-positive EPS/BVPS must yield 0 (not a candidate)")
	}
}

func TestIntrinsicValue(t *testing.T) {
	// EPS 5, growth 10% -> 5·(8.5 + 2·10) = 5·28.5 = 142.5
	if got := IntrinsicValue(5, 0.10); !approx(got, 142.5) {
		t.Errorf("IntrinsicValue(5,0.10) = %v, want 142.5", got)
	}
	// Growth capped at 15%: 20% should behave like 15% -> 5·(8.5+30)=192.5
	if got := IntrinsicValue(5, 0.20); !approx(got, 192.5) {
		t.Errorf("growth cap failed: got %v, want 192.5", got)
	}
	// Negative growth floored at 0 -> 5·8.5 = 42.5
	if got := IntrinsicValue(5, -0.5); !approx(got, 42.5) {
		t.Errorf("negative growth floor failed: got %v, want 42.5", got)
	}
	if IntrinsicValue(0, 0.1) != 0 {
		t.Error("non-positive EPS must yield 0")
	}
}

func TestBuyBelowPrice(t *testing.T) {
	// 30% margin of safety on intrinsic 100 -> buy below 70.
	if got := BuyBelowPrice(100, 0.30); !approx(got, 70) {
		t.Errorf("BuyBelowPrice(100,0.30) = %v, want 70", got)
	}
	if BuyBelowPrice(0, 0.3) != 0 {
		t.Error("zero intrinsic must yield 0")
	}
	// No margin -> buy below intrinsic itself.
	if got := BuyBelowPrice(100, 0); !approx(got, 100) {
		t.Errorf("BuyBelowPrice(100,0) = %v, want 100", got)
	}
}
