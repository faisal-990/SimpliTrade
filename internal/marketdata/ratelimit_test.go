package marketdata

import (
	"context"
	"testing"
	"time"
)

func TestThrottle_SpacesCalls(t *testing.T) {
	inner := &countingProvider{}
	// 20ms minimum interval between upstream calls.
	thr := NewThrottledProvider(inner, 20*time.Millisecond)

	start := time.Now()
	for range 3 {
		if _, err := thr.Quote(context.Background(), "AAPL"); err != nil {
			t.Fatal(err)
		}
	}
	// 3 calls => at least 2 inter-call gaps of 20ms => >= 40ms.
	if elapsed := time.Since(start); elapsed < 40*time.Millisecond {
		t.Errorf("3 calls took %v, want >= 40ms (throttled)", elapsed)
	}
	if inner.quotes != 3 {
		t.Errorf("inner called %d times, want 3", inner.quotes)
	}
}

func TestThrottle_RespectsContextCancel(t *testing.T) {
	thr := NewThrottledProvider(&countingProvider{}, time.Hour) // huge interval
	if _, err := thr.Quote(context.Background(), "AAPL"); err != nil {
		t.Fatal(err) // first call doesn't wait
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // second call would wait an hour; cancellation must short-circuit
	if _, err := thr.Quote(ctx, "AAPL"); err == nil {
		t.Fatal("expected context cancellation error while throttled")
	}
}
