package ingest

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/rhuandantas/metrika/internal/smartblox"
)

type Config struct {
	PersistEvery time.Duration
	PollEvery    time.Duration
	EventsPath   string
	StatePath    string
}

type Ingestor struct {
	cli                smartblox.Client
	cfg                Config
	logger             *log.Logger
	LastProcessedRound int64
}

func New(cli smartblox.Client, cfg Config, logger *log.Logger) *Ingestor {
	return &Ingestor{cli: cli, cfg: cfg, logger: logger}
}

func (i *Ingestor) Run(ctx context.Context) error {
	ticker := time.NewTicker(i.cfg.PollEvery)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return context.Canceled
		case <-ticker.C:
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

	for r := i.LastProcessedRound + 1; r <= status.LastRound; r++ {
		if err := i.processRound(ctx, r); err != nil {
			return err
		}
		i.LastProcessedRound = r
	}
	return nil
}

func (i *Ingestor) processRound(ctx context.Context, round int64) error {
	b, err := i.cli.GetBlock(ctx, round)
	if err != nil {
		return fmt.Errorf("get block %d: %w", round, err)
	}
	for _, env := range b.Txs {
		if env.Tx.Type != "txfer" {
			continue
		}
		recipient := env.Tx.Receipient
		evt := Event{
			Round:     round,
			Sig:       env.Sig,
			Sender:    env.Tx.Sender,
			Recipient: recipient,
			Amount:    env.Tx.Amount,
		}

		fmt.Println(evt)
	}
	return nil
}
