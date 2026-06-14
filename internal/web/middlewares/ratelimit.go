package middlewares

import (
	"sync"
	"time"

	"github.com/faisal-990/ProjectInvestApp/internal/web/httpx"
	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// ipRateLimiter holds a token-bucket limiter per client IP. Idle entries are
// reaped so the map can't grow without bound under many distinct IPs.
type ipRateLimiter struct {
	mu       sync.Mutex
	visitors map[string]*visitor
	rps      rate.Limit
	burst    int
}

type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// RateLimit enforces per-IP request limits (token bucket). rps<=0 disables it.
// This is the first line of defense for public endpoints (especially auth).
func RateLimit(rps, burst int) gin.HandlerFunc {
	if rps <= 0 {
		return func(c *gin.Context) { c.Next() }
	}
	l := &ipRateLimiter{
		visitors: map[string]*visitor{},
		rps:      rate.Limit(rps),
		burst:    burst,
	}
	go l.reap()

	return func(c *gin.Context) {
		if !l.get(c.ClientIP()).Allow() {
			httpx.Fail(c, httpx.TooManyRequests("rate limit exceeded, slow down"))
			return
		}
		c.Next()
	}
}

func (l *ipRateLimiter) get(ip string) *rate.Limiter {
	l.mu.Lock()
	defer l.mu.Unlock()
	v, ok := l.visitors[ip]
	if !ok {
		v = &visitor{limiter: rate.NewLimiter(l.rps, l.burst)}
		l.visitors[ip] = v
	}
	v.lastSeen = time.Now()
	return v.limiter
}

// reap evicts visitors idle for more than 10 minutes.
func (l *ipRateLimiter) reap() {
	for range time.Tick(time.Minute) {
		l.mu.Lock()
		for ip, v := range l.visitors {
			if time.Since(v.lastSeen) > 10*time.Minute {
				delete(l.visitors, ip)
			}
		}
		l.mu.Unlock()
	}
}
