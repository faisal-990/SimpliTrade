package runner

import (
	"context"
	"time"
)

// Run drives RunOnce on a fixed interval until the context is cancelled. It runs
// one cycle immediately, then on each tick. A cycle error is logged and the loop
// continues — a transient market-source failure must not kill the engine.
//
// The single interval intentionally collapses the "slow" (market refresh) and
// "fast" (decide) lanes for now, because the FakeProvider's fundamentals are
// cheap. When a real, rate-limited provider lands (T9), the slow lane moves to a
// longer cadence and the two clocks separate — the decision code is unchanged.
func (r *Runner) Run(ctx context.Context, interval time.Duration) {
	r.log.Info("engine: starting", "interval", interval.String(), "bots", len(r.bots))

	if err := r.RunOnce(ctx); err != nil {
		r.log.Error("engine: cycle failed", "err", err)
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			r.log.Info("engine: shutting down")
			return
		case <-ticker.C:
			if err := r.RunOnce(ctx); err != nil {
				r.log.Error("engine: cycle failed", "err", err)
			}
		}
	}
}
