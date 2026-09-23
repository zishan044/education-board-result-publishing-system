package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/zishan044/education-board-result-publishing-system/internal/api"
	"github.com/zishan044/education-board-result-publishing-system/internal/cache"
	"github.com/zishan044/education-board-result-publishing-system/internal/config"
	"github.com/zishan044/education-board-result-publishing-system/internal/middleware"
	"github.com/zishan044/education-board-result-publishing-system/internal/store"
)

func main() {
	if err := run(); err != nil {
		slog.Error("fatal", "err", err)
		os.Exit(1)
	}
}

func run() error {

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	st, err := store.New(ctx, cfg.DatabaseURL, cfg.DBMaxConns)
	if err != nil {
		return err
	}
	defer st.Close()

	handler := api.NewHandler(st, cfg.RequestTimeout)

	rdb := redis.NewClient(&redis.Options{Addr: cfg.ValkeyAddr})
	ch := cache.New(rdb)

	ipFilter, err := middleware.NewIPFilter(cfg.BlockedCIDRs, nil, ch)
	if err != nil {
		return err
	}

	rateLimiter := middleware.NewRateLimiter(rdb, 20, time.Minute)

	router := api.NewRouter(handler,
		ipFilter.Handler(),
		rateLimiter.Handler(),
	)

	srv := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		slog.Info("listening", "addr", cfg.HTTPAddr)
		errCh <- srv.ListenAndServe()
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		slog.Info("shutting down")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}