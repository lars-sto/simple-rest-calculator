package httpapi

import (
	"net/http"
	"strconv"
	"time"
)

func routeLabel(path string) string {
	switch path {
	case "/add", "/subtract", "/multiply", "/divide",
		"/recent", "/healthz", "/readyz":
		return path
	default:
		return "unknown"
	}
}

func (a *App) withLogging(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rec := newResponseRecorder(w)
		start := time.Now()

		next(rec, r)

		status := rec.status
		if status == 0 {
			status = http.StatusOK
		}

		pathLabel := routeLabel(r.URL.Path)
		httpRequestsTotal.WithLabelValues(
			r.Method,
			pathLabel,
			strconv.Itoa(status),
		).Inc()
		httpRequestDurationSeconds.
			WithLabelValues(r.Method, pathLabel).
			Observe(time.Since(start).Seconds())

		if a.Logger != nil {
			a.Logger.Printf(
				"method=%s path=%s status=%d bytes=%d dur=%s",
				r.Method,
				r.URL.Path,
				status,
				rec.body.Len(),
				time.Since(start),
			)
		}
	}
}
