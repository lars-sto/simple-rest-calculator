package httpapi

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests processed.",
		},
		[]string{"method", "path", "status"},
	)

	httpRequestDurationSeconds = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request latency distributions.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	httpRequestsInFlight = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "http_requests_in_flight",
			Help: "Number of HTTP requests currently being served.",
		},
		[]string{"path"},
	)
)

func RegisterMetrics(reg prometheus.Registerer) {
	reg.MustRegister(httpRequestsTotal, httpRequestDurationSeconds, httpRequestsInFlight)
}

func (a *App) withMetrics(pathLabel string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		httpRequestsInFlight.WithLabelValues(pathLabel).Inc()
		start := time.Now()

		rec := newResponseRecorder(w)
		next(rec, r)

		status := rec.status
		if status == 0 {
			status = http.StatusOK
		}

		httpRequestsTotal.WithLabelValues(r.Method, pathLabel, strconv.Itoa(status)).Inc()
		httpRequestDurationSeconds.WithLabelValues(r.Method, pathLabel).Observe(time.Since(start).Seconds())
		httpRequestsInFlight.WithLabelValues(pathLabel).Dec()
	}
}
