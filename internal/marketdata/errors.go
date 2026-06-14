package marketdata

import "errors"

// ErrRateLimited is returned when a provider responds with HTTP 429. Callers
// (the caching wrapper / engine) back off rather than hammering the quota.
var ErrRateLimited = errors.New("marketdata: rate limited")
