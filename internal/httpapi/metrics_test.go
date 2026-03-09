package httpapi

import (
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"calculator-service/internal/service"
	"calculator-service/internal/store"

	"github.com/prometheus/client_golang/prometheus"
)

var registerMetricsOnce sync.Once

func ensureDefaultMetricsRegistered() {
	registerMetricsOnce.Do(func() {
		RegisterMetrics(prometheus.DefaultRegisterer)
	})
}

func discardLogger() *log.Logger {
	return log.New(io.Discard, "", 0)
}

// Test 1: /metrics is served (route exists) and returns Prometheus exposition text.
func TestMetricsEndpoint_Served(t *testing.T) {
	ensureDefaultMetricsRegistered()

	app := &App{
		Calc:   service.NewBaseCalculator(),
		Store:  store.NewMemoryStore(20),
		Logger: discardLogger(), // needed if your metrics increments live in withLogging
	}

	mux := http.NewServeMux()
	app.RegisterRoutes(mux)

	srv := httptest.NewServer(mux)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/metrics")
	if err != nil {
		t.Fatalf("GET /metrics failed: %v", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	b, _ := io.ReadAll(resp.Body)
	txt := string(b)

	// Very lightweight sanity checks that it's Prometheus text format.
	// promhttp usually includes "# HELP" / "# TYPE" lines for many metrics.
	if !strings.Contains(txt, "#") && !strings.Contains(txt, "go_") && !strings.Contains(txt, "process_") {
		t.Fatalf("metrics output does not look like Prometheus exposition format:\n%s", txt)
	}
}

// Test 2: after one request, /metrics contains our custom metric families.
func TestMetricsEndpoint_ContainsCustomMetrics(t *testing.T) {
	ensureDefaultMetricsRegistered()

	app := &App{
		Calc:   service.NewBaseCalculator(),
		Store:  store.NewMemoryStore(20),
		Logger: discardLogger(), // needed if metrics are emitted from withLogging
	}

	mux := http.NewServeMux()
	app.RegisterRoutes(mux)

	srv := httptest.NewServer(mux)
	defer srv.Close()

	// Trigger a route that goes through wrapRead/wrapMath so metrics increments happen.
	// /healthz is not wrapped in your code, so it would not create series.
	r1, err := http.Get(srv.URL + "/recent")
	if err != nil {
		t.Fatalf("GET /recent failed: %v", err)
	}
	_ = r1.Body.Close()

	resp, err := http.Get(srv.URL + "/metrics")
	if err != nil {
		t.Fatalf("GET /metrics failed: %v", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	b, _ := io.ReadAll(resp.Body)
	txt := string(b)

	// Check that our metric families exist.
	if !strings.Contains(txt, "http_requests_total") {
		t.Fatalf("expected http_requests_total in metrics output, got:\n%s", txt)
	}
	if !strings.Contains(txt, "http_request_duration_seconds") {
		t.Fatalf("expected http_request_duration_seconds in metrics output, got:\n%s", txt)
	}

	// Optional but nice: check that the /recent label appears at least once.
	// If you use routeLabel mapping, it should be path="/recent".
	if !strings.Contains(txt, `path="/recent"`) {
		t.Fatalf("expected path label for /recent in metrics output, got:\n%s", txt)
	}
}
