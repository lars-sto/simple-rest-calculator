package httpapi

import (
	"bytes"
	"net/http"
)

type responseRecorder struct {
	w      http.ResponseWriter
	status int
	body   bytes.Buffer
}

func newResponseRecorder(w http.ResponseWriter) *responseRecorder {
	return &responseRecorder{w: w}
}

func (r *responseRecorder) Header() http.Header {
	return r.w.Header()
}

func (r *responseRecorder) WriteHeader(code int) {
	r.status = code
	r.w.WriteHeader(code)
}

func (r *responseRecorder) Write(p []byte) (int, error) {
	r.body.Write(p)
	return r.w.Write(p)
}
