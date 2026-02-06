package httpapi

import (
	"log"
	"net/http"

	"calculator-service/internal/service"
	"calculator-service/internal/store"
)

type App struct {
	Calc   service.Calculator
	Store  store.ResultStore
	Logger *log.Logger
}

func (a *App) RegisterRoutes(mux *http.ServeMux) {
	// Math: POST
	mux.HandleFunc("/add", a.wrapMath("add", a.handleAdd))
	mux.HandleFunc("/subtract", a.wrapMath("subtract", a.handleAdd))
	mux.HandleFunc("/multiply", a.wrapMath("multiply", a.handleAdd))
	mux.HandleFunc("/divide", a.wrapMath("divide", a.handleAdd))

	mux.HandleFunc("/recent", a.wrapRead(a.handleRecent))

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

func (a *App) wrapMath(op string, h http.HandlerFunc) http.HandlerFunc {
	return a.withRecover(a.withLogging(a.withMemory(op)(h)))
}

func (a *App) wrapRead(h http.HandlerFunc) http.HandlerFunc {
	return a.withRecover(a.withLogging(h))
}
