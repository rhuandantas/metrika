package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/rhuandantas/metrika/internal/ingest"
	"github.com/rhuandantas/metrika/internal/repository"
	client "github.com/rhuandantas/metrika/internal/smartblox"
)

func main() {
	var (
		baseURL   = "http://localhost:8080"
		sqLiteDns = "file:data/metrika.db?cache=shared&_journal=WAL&_busy_timeout=5000"
		//eventsPath = "./data/events.jsonl"
		pool    = 3 * time.Second
		timeout = 4 * time.Second
	)

	logger := log.New(os.Stdout, "", log.LstdFlags|log.Lmicroseconds)

	cli := client.NewHTTPClient(baseURL, timeout, true)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	repo, err := repository.NewSQLiteMetrics(sqLiteDns)
	if err != nil {
		log.Fatal("Failed to initialize repository: ", err)
	}

	if err = repo.Init(ctx); err != nil {
		log.Fatal("Failed to initialize database schema: ", err)
	}

	ing := ingest.New(cli, pool, logger, repo)

	go func() {
		if err := ing.Run(ctx); err != nil {
			log.Fatal("Server error: ", err)
		}
		stop()
	}()

	<-ctx.Done()
	log.Print("Shutting down server gracefully...")
}
