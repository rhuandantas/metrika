package main

import (
	"context"
	"os/signal"
	"syscall"
	"time"

	"github.com/natefinch/lumberjack"
	"github.com/rhuandantas/metrika/internal/ingest"
	"github.com/rhuandantas/metrika/internal/repository"
	client "github.com/rhuandantas/metrika/internal/smartblox"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	var (
		// TODO move to env vars or config file
		baseURL   = "http://localhost:8080"
		sqLiteDns = "file:data/db/metrika.db?cache=shared&_journal=WAL&_busy_timeout=5000"
		pool      = 5 * time.Second
		timeout   = 60 * time.Second
	)

	cli := client.NewHTTPClient(baseURL, timeout, true)

	logger := log.Logger.With().Logger()

	ctxParent := logger.WithContext(context.Background())
	ctx, stop := signal.NotifyContext(ctxParent, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	repo, err := repository.NewSQLiteMetrics(sqLiteDns)
	if err != nil {
		logger.Fatal().Msgf("Failed to initialize repository: %v", err)
	}

	if err = repo.Init(ctx); err != nil {
		logger.Fatal().Msgf("Failed to initialize database schema: %v", err)
	}

	ing := ingest.New(cli, pool, logger, setupEventLogger(), repo)

	go func() {
		if err := ing.Run(ctx); err != nil {
			log.Fatal().Msgf("Server error: %v", err)
		}
		stop()
	}()

	<-ctx.Done()
	log.Info().Msgf("Shutting down server gracefully...")
}

func setupEventLogger() zerolog.Logger {
	logFile := &lumberjack.Logger{
		Filename: "./data/events.log",
		MaxAge:   30,
		Compress: true,
	}

	return zerolog.New(logFile).With().Logger()
}
