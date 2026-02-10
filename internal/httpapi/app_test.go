package httpapi

import (
	"bytes"
	"calculator-service/internal/store"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRegisterRoutes_Wiring_MathRoutes(t *testing.T) {
	st := store.NewMemoryStore(50)
	app := &App{
		Calc:   serviceStub{},
		Store:  st,
		Logger: log.New(bytes.NewBuffer(nil), "", 0),
	}
	mux := http.NewServeMux()
	app.RegisterRoutes(mux)

	type tc struct {
		path      string
		wantOp    string
		wantValue float64
	}

	tests := []tc{
		{path: "/add", wantOp: "add", wantValue: 3},
		{path: "/subtract", wantOp: "subtract", wantValue: -1},
		{path: "/multiply", wantOp: "multiply", wantValue: 2},
		{path: "/divide", wantOp: "divide", wantValue: 0.5},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			body := bytes.NewBufferString(`{"a":1,"b":2}`)
			req := httptest.NewRequest(http.MethodPost, tt.path, body)
			rr := httptest.NewRecorder()

			mux.ServeHTTP(rr, req)

			// Assert: Route is registered
			if rr.Code == http.StatusNotFound {
				t.Fatalf("expected %s to be registered, got 404", tt.path)
			}
			if rr.Code != http.StatusOK {
				t.Fatalf("expected 200 OK, got %d; body=%q", rr.Code, rr.Body.String())
			}

			// Assert: Handler result is correct
			var resp oneNumberResponse
			if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
				t.Fatalf("failed to decode response JSON: %v", err)
			}
			if resp.Value != tt.wantValue {
				t.Fatalf("unexpected result: got %v want %v", resp.Value, tt.wantValue)
			}

			// Assert: op has been correctly stored
			entries, err := st.ListRecent(req.Context(), 1)
			if err != nil {
				t.Fatalf("ListRecent failed: %v", err)
			}
			if len(entries) != 1 {
				t.Fatalf("expected 1 stored entry, got %d", len(entries))
			}
			if string(entries[0].Op) != tt.wantOp {
				t.Fatalf("unexpected stored op: got %q want %q", entries[0].Op, tt.wantOp)
			}
		})
	}
}

type serviceStub struct{}

func (serviceStub) Add(ctx context.Context, a, b float64) (float64, error) { return 3, nil }
func (serviceStub) Sub(ctx context.Context, a, b float64) (float64, error) { return -1, nil }
func (serviceStub) Mul(ctx context.Context, a, b float64) (float64, error) { return 2, nil }
func (serviceStub) Div(ctx context.Context, a, b float64) (float64, error) { return 0.5, nil }

func TestRegisterRoutes_Healthz(t *testing.T) {
	app := &App{Store: store.NewMemoryStore(10)}
	mux := http.NewServeMux()
	app.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}
