package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	appconfig "github.com/mariana/rate-limiter/internal/config"
	"github.com/mariana/rate-limiter/internal/limiter"
	redisstrategy "github.com/mariana/rate-limiter/internal/limiter/strategy/redis"
	"github.com/mariana/rate-limiter/internal/middleware"
)

func main() {
	cfg, err := appconfig.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	storage, err := redisstrategy.New(cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB)
	if err != nil {
		log.Fatalf("failed to connect to redis: %v", err)
	}
	defer storage.Close()

	service := limiter.New(storage, limiter.Config{
		IPLimit:       cfg.IPLimit,
		TokenLimit:    cfg.TokenLimit,
		BlockDuration: cfg.BlockDuration,
		RateWindow:    cfg.RateWindow,
	})

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("hello world"))
	})

	handler := middleware.New(service).Handler(mux)
	addr := fmt.Sprintf(":%s", cfg.ServerPort)

	log.Printf("rate limiter listening on %s", addr)
	if err := http.ListenAndServe(addr, handler); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server failed: %v", err)
		os.Exit(1)
	}
}
