package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/Metrika-Inc/smartblox"
)

type Status struct {
	LastRound int64 `json:"last-round"`
}

type Transaction struct {
	Amount int64  `json:"amount"`
	Sender int64  `json:"sender"`
	Type   string `json:"type"`
	// Handle both spellings: recipient & receipient.
	Recipient int64 `json:"recipient"`
	// Receipient is captured only for JSON decoding compatibility.
	Receipient int64 `json:"receipient"`
}

type TransactionEnvelope struct {
	Sig string      `json:"sig"`
	Tx  Transaction `json:"tx"`
}

// Block models /api/blocks/{round}
type Block struct {
	Round int64                 `json:"round"`
	Txs   []TransactionEnvelope `json:"txs"`
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	go func() {
		if err := run(ctx); err != nil {
			log.Fatal("Server error: ", err)
		}
		stop()
	}()

	<-ctx.Done()
	log.Print("Shutting down server gracefully...")

}

func run(ctx context.Context) error {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return context.Canceled
		case <-ticker.C:
			round, err := smartblox.GetStatus()
			if err != nil {
				return err
			}
			var s Status
			err = json.Unmarshal(round, &s)
			if err != nil {
				return err
			}

			block, err := smartblox.GetBlock(s.LastRound)
			if err != nil {
				return err
			}

			var b Block
			err = json.Unmarshal(block, &b)
			if err != nil {
				return err
			}

			log.Printf("Block Round: %d", b.Round)
			for _, tx := range b.Txs {
				if tx.Tx.Type == "txfer" {
					log.Printf("Tx Sig: %s, Sender: %d, Recipient: %d, Amount: %d, Type: %s",
						tx.Sig, tx.Tx.Sender, tx.Tx.GetRecipient(), tx.Tx.Amount, tx.Tx.Type)
				}
			}
		}
	}
}

func (t *Transaction) GetRecipient() int64 {
	if t.Recipient != 0 {
		return t.Recipient
	}
	return t.Receipient
}
