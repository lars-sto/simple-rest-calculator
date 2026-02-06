package httpapi

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestResponseRecorder_CapturesStatusAndBody(t *testing.T) {
	rr := httptest.NewRecorder()
	rec := newResponseRecorder(rr)

	if rec.status != 0 {
		t.Fatalf("expected initial status 0, got %d", rec.status)
	}

	rec.WriteHeader(http.StatusTeapot)
	if rec.status != http.StatusTeapot {
		t.Fatalf("expected status %d, got %d", http.StatusTeapot, rec.status)
	}

	_, _ = rec.Write([]byte("hello"))
	if rec.body.String() != "hello" {
		t.Fatalf("expected body 'hello', got %q", rec.body.String())
	}

	// Ensure it also wrote to the underlying recorder
	if rr.Body.String() != "hello" {
		t.Fatalf("expected underlying response body 'hello', got %q", rr.Body.String())
	}
	if rr.Code != http.StatusTeapot {
		t.Fatalf("expected underlying status %d, got %d", http.StatusTeapot, rr.Code)
	}
}
