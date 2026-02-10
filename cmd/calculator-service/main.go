package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"calculator-service/internal/httpapi"
	"calculator-service/internal/service"
	"calculator-service/internal/store"

	"github.com/prometheus/client_golang/prometheus"
)

func main() {
	addr := flag.String("addr", ":8080", "listen address")
	persist := flag.Bool("persist", false, "persist recent results to disk (jsonl)")
	dataPath := flag.String("data-path", "./data/recent.jsonl", "path to the jsonl file used for persistence")
	flag.Parse()

	logger := log.New(os.Stdout, "", log.LstdFlags|log.Lmicroseconds)

	var st store.ResultStore
	var closers []func() error

	if *persist {
		fb, err := store.NewFileBackedStore(*dataPath, 20)
		if err != nil {
			logger.Fatalf("failed to init file-backed store: %v", err)
		}
		st = fb
		closers = append(closers, fb.Close)
		logger.Printf("persistence enabled: %s", *dataPath)
	} else {
		st = store.NewMemoryStore(20)
		logger.Printf("persistence disabled: using in-memory store only")
	}

	app := &httpapi.App{
		Calc:   service.NewBaseCalculator(),
		Store:  st,
		Logger: logger,
	}

	mux := http.NewServeMux()
	app.RegisterRoutes(mux)
	httpapi.RegisterMetrics(prometheus.DefaultRegisterer)

	srv := &http.Server{
		Addr:              *addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	// Signal handling
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		logger.Printf("listening on %s", *addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
		close(errCh)
	}()

	select {
	case <-ctx.Done():
		logger.Printf("shutdown signal received")
	case err := <-errCh:
		if err != nil {
			logger.Fatalf("server error: %v", err)
		}
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_ = srv.Shutdown(shutdownCtx)

	for _, c := range closers {
		_ = c()
	}
	logger.Printf("shutdown complete")
}
