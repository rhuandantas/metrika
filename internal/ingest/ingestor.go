package ingest

import (
	"context"
	"errors"
	"time"

	"github.com/rhuandantas/metrika/internal/models"
	"github.com/rhuandantas/metrika/internal/repository"
	"github.com/rhuandantas/metrika/internal/smartblox"
	"github.com/rs/zerolog"
)

type Ingestor struct {
	cli             smartblox.Client
	poolEvery       time.Duration
	logger          zerolog.Logger
	repo            repository.Repository
	eventJsonWriter EventDataWriter
}

func New(cli smartblox.Client, poolEvery time.Duration, logger zerolog.Logger, repo repository.Repository, eventJsonWriter EventDataWriter) *Ingestor {
	return &Ingestor{cli: cli, poolEvery: poolEvery, logger: logger, repo: repo, eventJsonWriter: eventJsonWriter}
}

func (i *Ingestor) Run(ctx context.Context) error {
	i.logger.Info().Msg("Starting Ingestor...")
	i.logger.Debug().Msg("Tick interval: " + i.poolEvery.String())
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
				i.logger.Error().Msgf("processing error: %v", err)
			}
		}
	}
}

func (i *Ingestor) process(ctx context.Context) error {
	status, err := i.cli.GetStatus(ctx)
	if err != nil {
		i.logger.Error().Msgf("Error getting status: %v", err)
		return err
	}

	metrics, err := i.repo.LoadMetrics(ctx)
	if err != nil {
		i.logger.Error().Msgf("Error loading metrics: %v", err)
		return err
	}

	if metrics.LastRound == 0 {
		metrics.LastRound = status.LastRound - 1
	}

	for r := metrics.LastRound + 1; r <= status.LastRound; r++ {
		if err = i.processRound(ctx, r, &metrics); err != nil {
			return err
		}
	}

	return nil
}

func (i *Ingestor) processRound(ctx context.Context, round int64, metrics *models.Metrics) error {
	b, err := i.cli.GetBlock(ctx, round)
	if err != nil {
		i.logger.Error().Msgf("Error getting block %d: %v", round, err)
		return err
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
		i.logger.Error().Msgf("Error updating metrics: %v", err)
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
