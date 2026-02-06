package main

import (
	"errors"
	"flag"
	"log"
	"net/http"
	"os"
	"time"

	"calculator-service/internal/httpapi"
	"calculator-service/internal/service"
	"calculator-service/internal/store"
)

func main() {
	addr := flag.String("addr", ":8080", "listen address")
	persist := flag.Bool("persist", false, "persist recent results to disk (jsonl)")
	dataPath := flag.String("data-path", "./data/recent.jsonl",
		"path to the jsonl file used for persistence")
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

	srv := &http.Server{
		Addr:              *addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	defer func() {
		for _, c := range closers {
			_ = c()
		}
	}()

	logger.Printf("listening on %s", *addr)
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Fatalf("server error: %v", err)
	}
}
