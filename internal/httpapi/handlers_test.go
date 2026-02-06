package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"math"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"calculator-service/internal/model"
	"calculator-service/internal/service"
	"calculator-service/internal/store"
)

const eps = 1e-5

func newTestApp(t *testing.T) *App {
	t.Helper()
	return &App{
		Calc:   service.NewBaseCalculator(),
		Store:  store.NewMemoryStore(20),
		Logger: log.New(&bytes.Buffer{}, "", 0),
	}
}

func TestHandleTwoIntOp_MethodNotAllowed(t *testing.T) {
	app := newTestApp(t)

	req := httptest.NewRequest(http.MethodGet, "http://example/add", nil)
	rr := httptest.NewRecorder()

	app.handleAdd(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rr.Code)
	}
}

func TestHandleTwoIntOp_InvalidJSON(t *testing.T) {
	app := newTestApp(t)

	req := httptest.NewRequest(http.MethodPost, "http://example/add", bytes.NewBufferString(`{"a":1,`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	app.handleAdd(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestHandleArithmetic_Success(t *testing.T) {
	app := newTestApp(t)

	type tc struct {
		name   string
		path   string
		call   func(http.ResponseWriter, *http.Request)
		body   string
		expect float64
	}

	tests := []tc{
		{name: "add_int", path: "http://example/add", call: app.handleAdd,
			body: `{"a":1,"b":2}`, expect: 3},

		{name: "add_negative", path: "http://example/add", call: app.handleAdd,
			body: `{"a":4,"b":-5}`, expect: -1},

		{name: "add_float", path: "http://example/add", call: app.handleAdd,
			body: `{"a":1.5,"b":2.2}`, expect: 3.7},

		{name: "subtract_int", path: "http://example/subtract", call: app.handleSub,
			body: `{"a":5,"b":3}`, expect: 2},

		{name: "subtract_big_second_number", path: "http://example/subtract", call: app.handleSub,
			body: `{"a":1,"b":5}`, expect: -4},

		{name: "subtract_negative", path: "http://example/subtract", call: app.handleSub,
			body: `{"a":1,"b":-5}`, expect: 6},

		{name: "multiply", path: "http://example/multiply", call: app.handleMul,
			body: `{"a":4,"b":6}`, expect: 24},

		{name: "multiply_negative", path: "http://example/multiply", call: app.handleMul,
			body: `{"a":3,"b":-3}`, expect: -9},

		{name: "multiply_float", path: "http://example/multiply", call: app.handleMul,
			body: `{"a":2.5,"b":5.5}`, expect: 13.75},

		{name: "divide_exact", path: "http://example/divide", call: app.handleDiv,
			body: `{"a":10,"b":2}`, expect: 5},

		{name: "divide_fraction", path: "http://example/divide", call: app.handleDiv,
			body: `{"a":9,"b":2}`, expect: 4.5},

		{name: "add_big_ok", path: "http://example/add", call: app.handleAdd,
			body: `{"a":1000000000000,"b":2000000000000}`, expect: 3000000000000},

		{name: "mul_big_ok", path: "http://example/multiply", call: app.handleMul,
			body: `{"a":1000000,"b":1000000}`, expect: 1000000000000},

		{name: "div_big_ok", path: "http://example/divide", call: app.handleDiv,
			body: `{"a":1000000000000,"b":3}`, expect: 333333333333.33333},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, tt.path, bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()

			tt.call(rr, req)

			if rr.Code != http.StatusOK {
				t.Fatalf("expected 200, got %d (body=%q)", rr.Code, rr.Body.String())
			}

			var resp oneNumberResponse
			if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
				t.Fatalf("invalid json response: %v (body=%q)", err, rr.Body.String())
			}
			if math.Abs(resp.Value-tt.expect) > eps {
				t.Fatalf("expected %.5f, got %.5f", tt.expect, resp.Value)
			}
		})
	}
}

func TestHandleDivide_DivisionByZero(t *testing.T) {
	app := newTestApp(t)

	req := httptest.NewRequest(http.MethodPost, "http://example/divide", bytes.NewBufferString(`{"a":1,"b":0}`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	app.handleDiv(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
	if !bytes.Contains(rr.Body.Bytes(), []byte("division by zero")) {
		t.Fatalf("expected body to contain 'division by zero', got %q", rr.Body.String())
	}
}

func TestHandleRecent_DefaultAndClamp(t *testing.T) {
	app := newTestApp(t)

	// Fill store with 20 entries
	for i := float64(1); i <= 20; i++ {
		v := i
		err := app.Store.Add(context.Background(), model.ResultEntry{
			Time:      time.Now(),
			Op:        model.OpAdd,
			A:         i,
			B:         0,
			Result:    &v,
			ExprHuman: "x",
		})
		if err != nil {
			t.Fatalf("Add error: %v", err)
		}
	}

	// default n=5
	{
		req := httptest.NewRequest(http.MethodGet, "http://example/recent", nil)
		rr := httptest.NewRecorder()
		app.handleRecent(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rr.Code)
		}

		var res []model.ResultEntry
		if err := json.Unmarshal(rr.Body.Bytes(), &res); err != nil {
			t.Fatalf("invalid json response: %v", err)
		}
		if len(res) != 5 {
			t.Fatalf("expected 5 entries, got %d", len(res))
		}
		if res[0].A != 20 {
			t.Fatalf("expected newest A=20, got %.5f", res[0].A)
		}
	}

	// clamp n=999 -> 20
	{
		req := httptest.NewRequest(http.MethodGet, "http://example/recent?n=999", nil)
		rr := httptest.NewRecorder()
		app.handleRecent(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rr.Code)
		}

		var res []model.ResultEntry
		if err := json.Unmarshal(rr.Body.Bytes(), &res); err != nil {
			t.Fatalf("invalid json response: %v", err)
		}
		if len(res) != 20 {
			t.Fatalf("expected 20 entries, got %d", len(res))
		}
	}
}

func TestHandleTwoNumberOp_RejectsInvalidNumbers(t *testing.T) {
	app := newTestApp(t)

	cases := []struct {
		name string
		path string
		call func(http.ResponseWriter, *http.Request)
		body string
	}{
		{name: "add_too_large_a", path: "http://example/add", call: app.handleAdd,
			body: `{"a":1e308,"b":1}`},

		{name: "add_too_large_b", path: "http://example/add", call: app.handleAdd,
			body: `{"a":1,"b":1e308}`},

		{name: "multiply_too_large", path: "http://example/multiply", call: app.handleMul,
			body: `{"a":1e200,"b":1e200}`},

		{name: "divide_inf_input", path: "http://example/divide", call: app.handleDiv,
			body: `{"a":1e309,"b":2}`},

		{name: "nan_input", path: "http://example/add", call: app.handleAdd,
			body: `{"a":NaN,"b":1}`},

		{name: "result_overflow_add", path: "http://example/add", call: app.handleAdd,
			body: `{"a":1e15,"b":1e15}`},

		{name: "result_overflow_multiply", path: "http://example/multiply", call: app.handleMul,
			body: `{"a":1e15,"b":1e15}`},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, tt.path, bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()

			tt.call(rr, req)

			if rr.Code != http.StatusBadRequest {
				t.Fatalf(
					"expected status 400, got %d (body=%q)",
					rr.Code,
					rr.Body.String(),
				)
			}
		})
	}
}

func TestHandleRecent_InvalidN(t *testing.T) {
	app := newTestApp(t)

	req := httptest.NewRequest(http.MethodGet, "http://example/recent?n=-1", nil)
	rr := httptest.NewRecorder()

	app.handleRecent(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}
