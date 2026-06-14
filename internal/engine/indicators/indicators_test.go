package indicators

import (
	"math"
	"testing"
)

func approx(a, b float64) bool { return math.Abs(a-b) < 1e-6 }

func TestSMA(t *testing.T) {
	if v, ok := SMA([]float64{10, 20, 30}, 3); !ok || !approx(v, 20) {
		t.Errorf("SMA = %v ok=%v, want 20 true", v, ok)
	}
	if v, ok := SMA([]float64{10, 20, 30}, 2); !ok || !approx(v, 25) {
		t.Errorf("SMA last2 = %v ok=%v, want 25 true", v, ok)
	}
	if _, ok := SMA([]float64{10}, 3); ok {
		t.Error("SMA with insufficient data must return ok=false")
	}
}

func TestRSI_AllGainsIs100(t *testing.T) {
	if v, ok := RSI([]float64{1, 2, 3, 4, 5}, 4); !ok || !approx(v, 100) {
		t.Errorf("RSI all-up = %v ok=%v, want 100 true", v, ok)
	}
}

func TestRSI_MidRange(t *testing.T) {
	// Balanced alternating moves over 6 periods (3 up, 3 down) -> RSI 50.
	v, ok := RSI([]float64{10, 11, 10, 11, 10, 11, 10}, 6)
	if !ok {
		t.Fatal("RSI ok=false")
	}
	if !approx(v, 50) {
		t.Errorf("RSI of balanced series = %v, want 50", v)
	}
	if _, ok := RSI([]float64{1, 2}, 5); ok {
		t.Error("RSI with insufficient data must return ok=false")
	}
}

func TestReturnOver(t *testing.T) {
	// last=120, 3 back = 100 -> +20%
	if v, ok := ReturnOver([]float64{100, 105, 110, 120}, 3); !ok || !approx(v, 0.2) {
		t.Errorf("ReturnOver = %v ok=%v, want 0.2 true", v, ok)
	}
	if _, ok := ReturnOver([]float64{100, 120}, 5); ok {
		t.Error("ReturnOver with insufficient data must return ok=false")
	}
}

func TestHigh(t *testing.T) {
	if v, ok := High([]float64{10, 35, 20, 15}, 4); !ok || !approx(v, 35) {
		t.Errorf("High = %v ok=%v, want 35 true", v, ok)
	}
}
