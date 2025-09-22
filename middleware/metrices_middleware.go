package middleware

import (
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"

)

var (
	requestCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"path", "method", "status"},
	)

	requestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duration of HTTP requests in seconds",
		},
		[]string{"path", "method"},
	)

	statusCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_response_status_total",
			Help: "Total number of HTTP response statuses",
		},
		[]string{"path", "method", "status_code"},
	)
)

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func init() {
	prometheus.MustRegister(requestCounter, requestDuration, statusCounter)
}


func MetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Start time of the request
		start := time.Now()

		ww := &responseWriter{ResponseWriter: w}

		next.ServeHTTP(ww, r)

		// Duration of the request
		duration := time.Since(start).Seconds()

		requestCounter.WithLabelValues(r.URL.Path, r.Method, http.StatusText(ww.statusCode)).Inc()
		requestDuration.WithLabelValues(r.URL.Path, r.Method).Observe(duration)
		statusCounter.WithLabelValues(r.URL.Path, r.Method, http.StatusText(ww.statusCode)).Inc()

	})
}

func (rw *responseWriter) WriteHeader(statusCode int) {
	rw.statusCode = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}