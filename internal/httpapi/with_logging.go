package httpapi

import (
	"net/http"
	"time"
)

func (a *App) withLogging(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if a.Logger == nil {
			next(w, r)
			return
		}

		rec := newResponseRecorder(w)
		start := time.Now()

		next(rec, r)

		status := rec.status
		if status == 0 {
			status = http.StatusOK
		}
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
