package httpapi

import (
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"calculator-service/internal/service"
	"calculator-service/internal/store"
)

func TestWithRecover_ConvertsPanicTo500(t *testing.T) {
	app := &App{
		Calc:   service.NewBaseCalculator(),
		Store:  store.NewMemoryStore(20),
		Logger: log.New(httptest.NewRecorder().Body, "", 0), // any logger, not used here
	}

	panicHandler := func(w http.ResponseWriter, r *http.Request) {
		panic("boom")
	}

	h := app.withRecover(panicHandler)

	req := httptest.NewRequest(http.MethodGet, "http://example/any", nil)
	rr := httptest.NewRecorder()

	h(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", rr.Code)
	}
}
