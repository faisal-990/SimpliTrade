package runner

import (
	"context"
	"time"
)

// Run drives RunOnce on a fixed interval until the context is cancelled. It runs
// one cycle immediately, then on each tick. A cycle error is logged and the loop
// continues — a transient market-source failure must not kill the engine.
//
// If a MarketClock is set (WithClock), cycles run only while the market is open:
// the daemon idles overnight/weekends and resumes itself at the next open, so it
// behaves like production with no manual trigger. Without a clock it runs every
// tick (sandbox/demo).
//
// The single interval intentionally collapses the "slow" (market refresh) and
// "fast" (decide) lanes for now. When the slow fundamentals lane needs a longer
// cadence, the two clocks separate — the decision code is unchanged.
func (r *Runner) Run(ctx context.Context, interval time.Duration) {
	r.log.Info("engine: starting", "interval", interval.String(), "bots", len(r.bots), "market_gated", r.clock != nil)

	r.tick(ctx)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			r.log.Info("engine: shutting down")
			return
		case <-ticker.C:
			r.tick(ctx)
		}
	}
}

// tick runs one cycle, respecting the market clock if one is set. It logs the
// open→closed transition once (an end-of-day marker) rather than on every idle
// tick, so overnight logs stay quiet.
func (r *Runner) tick(ctx context.Context) {
	if r.clock != nil && !r.clock.IsOpen(time.Now()) {
		if r.wasOpen {
			r.log.Info("engine: market closed — session ended, idling until next open")
			r.wasOpen = false
		}
		return
	}
	if r.clock != nil && !r.wasOpen {
		r.log.Info("engine: market open — session started")
	}
	r.wasOpen = true
	if err := r.RunOnce(ctx); err != nil {
		r.log.Error("engine: cycle failed", "err", err)
	}
}
