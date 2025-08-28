package ingest

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/rhuandantas/metrika/internal/models"
	"github.com/rhuandantas/metrika/internal/repository"
	"github.com/rhuandantas/metrika/internal/smartblox"
)

type Ingestor struct {
	cli       smartblox.Client
	poolEvery time.Duration
	logger    *log.Logger
	repo      repository.Repository
}

func New(cli smartblox.Client, poolEvery time.Duration, logger *log.Logger, repo repository.Repository) *Ingestor {
	return &Ingestor{cli: cli, poolEvery: poolEvery, logger: logger, repo: repo}
}

func (i *Ingestor) Run(ctx context.Context) error {
	ticker := time.NewTicker(i.poolEvery)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return context.Canceled
		case <-ticker.C:
			log.Print("tick")
			if err := i.process(ctx); err != nil {
				if errors.Is(err, context.Canceled) {
					return err
				}
				i.logger.Printf("step error: %v", err)
			}
		}
	}
}

func (i *Ingestor) process(ctx context.Context) error {
	status, err := i.cli.GetStatus(ctx)
	if err != nil {
		return fmt.Errorf("get status: %w", err)
	}

	metrics, err := i.repo.LoadMetrics(ctx)
	if err != nil {
		return err
	}

	if metrics.LastRound == 0 {
		metrics.LastRound = status.LastRound - 1
	}

	for r := metrics.LastRound + 1; r <= status.LastRound; r++ {
		if err := i.processRound(ctx, r, &metrics); err != nil {
			return err
		}
	}

	return nil
}

func (i *Ingestor) processRound(ctx context.Context, round int64, metrics *models.Metrics) error {
	b, err := i.cli.GetBlock(ctx, round)
	if err != nil {
		return fmt.Errorf("get block %d: %w", round, err)
	}
	events := make([]Event, 0, len(b.Txs))
	for i, env := range b.Txs {
		if env.Tx.Type != "txfer" {
			continue
		}
		recipient := env.Tx.Receipient

		events[i] = Event{
			Round:     round,
			Sig:       env.Sig,
			Sender:    env.Tx.Sender,
			Recipient: recipient,
			Amount:    env.Tx.Amount,
		}

		metrics.Update(env.Tx.Amount, round)
	}

	// TODO this could be processed async via queue
	err = i.updateMetrics(ctx, *metrics)
	if err != nil {
		return err
	}

	return nil
}

func (i *Ingestor) updateMetrics(ctx context.Context, metrics models.Metrics) error {
	err := i.repo.SaveMetrics(ctx, metrics)
	if err != nil {
		return err
	}

	return nil
}
