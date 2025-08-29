package ingest

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/rhuandantas/metrika/internal/models"
	"github.com/rhuandantas/metrika/internal/repository"
	"github.com/rhuandantas/metrika/internal/smartblox"
	"github.com/rs/zerolog"
)

const transactionType = "txfer"

type Ingestor struct {
	cli          smartblox.Client
	poolEvery    time.Duration
	logger       zerolog.Logger
	repo         repository.Repository
	eventLogger  zerolog.Logger
	metricsCache *models.Metrics
	cacheTime    time.Time
}

func New(cli smartblox.Client, poolEvery time.Duration, logger, eventLogger zerolog.Logger, repo repository.Repository) *Ingestor {
	return &Ingestor{cli: cli, poolEvery: poolEvery, logger: logger, repo: repo, eventLogger: eventLogger}
}

// Run starts the ingestor process, polling the SmartBlox API at regular intervals defined by poolEvery.
// It continues to run until the provided context is canceled, at which point it returns context.Canceled.
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
			i.logger.Info().Msg("Polling SmartBlox API...")
			if err := i.process(ctx); err != nil {
				if errors.Is(err, context.Canceled) {
					return err
				}
				i.logger.Error().Msgf("processing error: %v", err)
			}
		}
	}
}

// process fetches the latest status from the SmartBlox API and processes any new rounds
func (i *Ingestor) process(ctx context.Context) error {
	status, err := i.cli.GetStatus(ctx)
	if err != nil {
		i.logger.Error().Msgf("Error getting status: %v", err)
		return err
	}

	metrics, err := i.getMetrics(ctx)
	if err != nil {
		i.logger.Error().Msgf("Error loading metrics: %v", err)
		return err
	}

	for r := metrics.LastRound + 1; r <= status.LastRound; r++ {
		if err = i.processRound(ctx, r, metrics); err != nil {
			return err
		}
	}

	return nil
}

// processRound processes a single round, extracting relevant events and updating metrics
func (i *Ingestor) processRound(ctx context.Context, round int64, metrics *models.Metrics) error {
	b, err := i.cli.GetBlock(ctx, round)
	if err != nil {
		i.logger.Error().Msgf("Error getting block %d: %v", round, err)
		return err
	}

	events := make([]Event, 0)
	for _, env := range b.Txs {
		if env.Tx.Type != transactionType {
			continue
		}
		recipient := env.Tx.Receipient

		events = append(events, Event{
			Round:     round,
			Sig:       env.Sig,
			Sender:    env.Tx.Sender,
			Recipient: recipient,
			Amount:    env.Tx.Amount,
		})

		metrics.Update(env.Tx.Amount, round)
	}

	if len(events) > 0 {
		marshal, _ := json.Marshal(events)
		i.eventLogger.Println(string(marshal))
	}

	return nil
}

// getMetrics retrieves the current metrics from the repository, using a simple in-memory cache to avoid frequent database hits
func (i *Ingestor) getMetrics(ctx context.Context) (*models.Metrics, error) {
	// TODO improve caching strategy using redis or similar
	if i.metricsCache == nil || time.Since(i.cacheTime) > 5*time.Minute {
		metrics, err := i.repo.LoadMetrics(ctx)
		if err != nil {
			i.logger.Error().Msgf("Error loading metrics: %v", err)
			return nil, err
		}
		i.metricsCache = &metrics
		i.cacheTime = time.Now()
	}
	return i.metricsCache, nil
}

// updateMetrics saves the updated metrics to the repository and updates the in-memory cache
func (i *Ingestor) updateMetrics(ctx context.Context, metrics models.Metrics) error {
	err := i.repo.SaveMetrics(ctx, metrics)
	if err != nil {
		return err
	}

	i.metricsCache = &metrics
	return nil
}
