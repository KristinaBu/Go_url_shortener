package main

import (
	"context"
	"database/sql"
	"errors"
	"github.com/KristinaBu/Go_url_shortener/internal/domain"
	"github.com/KristinaBu/Go_url_shortener/pkg/cache"
	_ "github.com/jackc/pgx/v5/stdlib"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/KristinaBu/Go_url_shortener/internal/config"
	"github.com/KristinaBu/Go_url_shortener/internal/handler"
	"github.com/KristinaBu/Go_url_shortener/internal/repository"
	"github.com/KristinaBu/Go_url_shortener/internal/service"
)

func main() {
	cfg, err := config.Parse()
	if err != nil {
		slog.Error("failed to parse config", slog.Any("error", err))
		os.Exit(1)
	}

	logOutput, err := config.OpenLogOutput(cfg.LogOutput)
	if err != nil {
		slog.Error(
			"failed to open log output",
			slog.Any("error", err),
		)
		os.Exit(1)
	}
	defer logOutput.Close()

	appLogger := slog.New(
		slog.NewJSONHandler(logOutput, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		}),
	)

	linkCache, err := cache.NewLRU[string, domain.Link](cfg.CacheSize)
	if err != nil {
		appLogger.Error(
			"failed to create cache",
			slog.Any("error", err),
		)
		os.Exit(1)
	}

	var (
		repo service.LinkRepository
		db   *sql.DB
	)

	switch cfg.Storage {
	case config.StorageMemory:
		repo = repository.NewMemoryRepository()

	case config.StoragePostgres:
		db, err = sql.Open("pgx", cfg.Database)
		if err != nil {
			appLogger.Error(
				"failed to open database",
				slog.Any("error", err),
			)
			os.Exit(1)
		}

		defer db.Close()

		ctx, cancel := context.WithTimeout(
			context.Background(),
			5*time.Second,
		)
		defer cancel()

		if err := db.PingContext(ctx); err != nil {
			appLogger.Error(
				"failed to connect to database",
				slog.Any("error", err),
			)
			os.Exit(1)
		}

		repo = repository.NewPostgresRepository(db)
	}

	svc := service.NewLinkService(
		repo,
		linkCache,
	)
	h := handler.New(svc)

	mux := http.NewServeMux()

	mux.HandleFunc("/links", h.CreateLink)
	mux.HandleFunc("/links/", h.GetLink)

	httpHandler := handler.RequestIDMiddleware(
		handler.LoggingMiddleware(appLogger, mux),
	)

	server := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           httpHandler,
		ReadHeaderTimeout: cfg.ReadHeaderTimeout,
		ReadTimeout:       cfg.ReadTimeout,
		WriteTimeout:      cfg.WriteTimeout,
		IdleTimeout:       cfg.IdleTimeout,
	}

	stop := make(chan os.Signal, 1)

	signal.Notify(
		stop,
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer signal.Stop(stop)

	go func() {
		<-stop

		appLogger.Info("shutting down server")

		ctx, cancel := context.WithTimeout(
			context.Background(),
			5*time.Second,
		)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			appLogger.Error(
				"server shutdown failed",
				slog.Any("error", err),
			)
		}
	}()

	appLogger.Info(
		"server started",
		slog.String("addr", cfg.HTTPAddr),
		slog.String("storage", cfg.Storage),
	)

	if err := server.ListenAndServe(); err != nil &&
		!errors.Is(err, http.ErrServerClosed) {
		appLogger.Error(
			"server failed",
			slog.Any("error", err),
		)

		os.Exit(1)
	}

	appLogger.Info("server stopped")
}
