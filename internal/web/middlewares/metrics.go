package middlewares

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// RED metrics (Rate, Errors, Duration). Labels use the route template
// (c.FullPath()), never the raw path, so high-cardinality ids (/investor/:id)
// don't explode the metric series.
var (
	httpRequests = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Total HTTP requests by method, route, and status.",
	}, []string{"method", "route", "status"})

	httpDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "http_request_duration_seconds",
		Help:    "HTTP request latency by method and route.",
		Buckets: prometheus.DefBuckets,
	}, []string{"method", "route"})
)

func init() {
	prometheus.MustRegister(httpRequests, httpDuration)
}

// Metrics records request rate, errors, and latency for every request.
func Metrics() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		route := c.FullPath()
		if route == "" {
			route = "unmatched"
		}
		httpDuration.WithLabelValues(c.Request.Method, route).Observe(time.Since(start).Seconds())
		httpRequests.WithLabelValues(c.Request.Method, route, strconv.Itoa(c.Writer.Status())).Inc()
	}
}

// MetricsHandler serves the Prometheus exposition endpoint (/metrics).
func MetricsHandler() gin.HandlerFunc {
	return gin.WrapH(promhttp.Handler())
}
