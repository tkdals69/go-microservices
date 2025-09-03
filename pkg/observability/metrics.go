package observability

import (
	"net/http"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	requestCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "route"},
	)

	responseDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_response_duration_seconds",
			Help:    "Duration of HTTP responses in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "route"},
	)
)

func init() {
	prometheus.MustRegister(requestCount)
	prometheus.MustRegister(responseDuration)
}

func MetricsHandler() http.Handler {
	return promhttp.Handler()
}

func RecordMetrics(method, route string, duration float64) {
	requestCount.WithLabelValues(method, route).Inc()
	responseDuration.WithLabelValues(method, route).Observe(duration)
}