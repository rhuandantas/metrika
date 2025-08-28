package main

import (
	"context"
	"os"
	"os/signal"
	"time"

	"github.com/rhuandantas/metrika/internal/ingest"
	"github.com/rhuandantas/metrika/internal/repository"
	client "github.com/rhuandantas/metrika/internal/smartblox"
	"github.com/rs/zerolog/log"
)

func main() {
	var (
		baseURL   = "http://localhost:8080"
		sqLiteDns = "file:data/metrika.db?cache=shared&_journal=WAL&_busy_timeout=5000"
		//eventsPath = "./data/events.jsonl"
		pool    = 3 * time.Second
		timeout = 4 * time.Second
	)

	logger := log.Logger.With().Str("component", "ingestor").Logger()

	cli := client.NewHTTPClient(baseURL, timeout, true)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	repo, err := repository.NewSQLiteMetrics(sqLiteDns)
	if err != nil {
		logger.Fatal().Msgf("Failed to initialize repository: %v", err)
	}

	if err = repo.Init(ctx); err != nil {
		logger.Fatal().Msgf("Failed to initialize database schema: %v", err)
	}

	ing := ingest.New(cli, pool, logger, repo)

	go func() {
		if err := ing.Run(ctx); err != nil {
			log.Fatal().Msgf("Server error: %v", err)
		}
		stop()
	}()

	<-ctx.Done()
	log.Info().Msgf("Shutting down server gracefully...")
}
