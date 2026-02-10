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
	mux.HandleFunc("/subtract", a.wrapMath("subtract", a.handleSub))
	mux.HandleFunc("/multiply", a.wrapMath("multiply", a.handleMul))
	mux.HandleFunc("/divide", a.wrapMath("divide", a.handleDiv))

	mux.HandleFunc("/recent", a.wrapRead(a.handleRecent))

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	// Optional. Only needed if startup takes long.
	// mux.HandleFunc("/startupz", func(w http.ResponseWriter, r *http.Request) {
	//     w.WriteHeader(http.StatusOK)
	// })
}

func (a *App) wrapMath(op string, h http.HandlerFunc) http.HandlerFunc {
	return a.withRecover(a.withLogging(a.withMemory(op)(h)))
}

func (a *App) wrapRead(h http.HandlerFunc) http.HandlerFunc {
	return a.withRecover(a.withLogging(h))
}
