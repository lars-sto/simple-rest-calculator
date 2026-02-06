package httpapi

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"calculator-service/internal/model"
	"calculator-service/internal/service"
	"calculator-service/internal/store"
)

func TestWithMemory_StoresSuccessResult(t *testing.T) {
	st := store.NewMemoryStore(20)
	var logBuf bytes.Buffer
	logger := log.New(&logBuf, "", 0)

	app := &App{
		Calc:   service.NewBaseCalculator(),
		Store:  st,
		Logger: logger,
	}

	// Handler returns 200 {"value": 3}
	handler := func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, oneNumberResponse{Value: 3})
	}

	h := app.withRecover(
		app.withLogging(
			app.withMemory("add")(
				handler,
			),
		),
	)

	req := httptest.NewRequest(http.MethodPost, "http://example/add", bytes.NewBufferString(`{"a":1,"b":2}`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h(rr, req)

	got, err := st.ListRecent(req.Context(), 20)
	if err != nil {
		t.Fatalf("ListRecent error: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(got))
	}

	e := got[0]
	if e.Op != model.Operation("add") {
		t.Fatalf("expected op 'add', got %q", e.Op)
	}
	if e.A != 1 || e.B != 2 {
		t.Fatalf("expected args a=1 b=2, got a=%.5f b=%.5f", e.A, e.B)
	}
	if e.Result == nil || *e.Result != 3 {
		t.Fatalf("expected result 3, got %+v", e.Result)
	}
	if e.Error != "" {
		t.Fatalf("expected no error, got %q", e.Error)
	}
}

func TestWithMemory_StoresDomainError(t *testing.T) {
	st := store.NewMemoryStore(20)
	var logBuf bytes.Buffer
	logger := log.New(&logBuf, "", 0)

	app := &App{
		Calc:   service.NewBaseCalculator(),
		Store:  st,
		Logger: logger,
	}

	// Handler returns 400 division by zero
	handler := func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, service.ErrDivisionByZero.Error(), http.StatusBadRequest)
	}

	h := app.withRecover(
		app.withLogging(
			app.withMemory("divide")(
				handler,
			),
		),
	)

	req := httptest.NewRequest(http.MethodPost, "http://example/divide", bytes.NewBufferString(`{"a":1,"b":0}`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h(rr, req)

	got, err := st.ListRecent(req.Context(), 20)
	if err != nil {
		t.Fatalf("ListRecent error: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(got))
	}

	e := got[0]
	if e.Op != ("divide") {
		t.Fatalf("expected op 'divide', got %q", e.Op)
	}
	if e.A != 1.0 || e.B != 0 {
		t.Fatalf("expected args a=1 b=0, got a=%.5f b=%.5f", e.A, e.B)
	}
	if e.Result != nil {
		t.Fatalf("expected nil result on error, got %+v", e.Result)
	}
	if e.Error == "" {
		t.Fatalf("expected error string, got empty")
	}
}

func TestWithMemory_DoesNotStoreInvalidJSON(t *testing.T) {
	st := store.NewMemoryStore(20)
	var logBuf bytes.Buffer
	logger := log.New(&logBuf, "", 0)

	app := &App{
		Calc:   service.NewBaseCalculator(),
		Store:  st,
		Logger: logger,
	}

	// Handler would respond 400 invalid json (simulating your real handler behavior)
	handler := func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "invalid json", http.StatusBadRequest)
	}

	h := app.withRecover(
		app.withLogging(
			app.withMemory("add")(
				handler,
			),
		),
	)

	req := httptest.NewRequest(http.MethodPost, "http://example/add", bytes.NewBufferString(`{"a":1,`)) // malformed
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h(rr, req)

	got, err := st.ListRecent(req.Context(), 20)
	if err != nil {
		t.Fatalf("ListRecent error: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("expected 0 entries, got %d", len(got))
	}
}
