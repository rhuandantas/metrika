package ingest

import (
	"encoding/json"
	"os"
)

// Event is a normalized txfer we dump to JSONL
type Event struct {
	Round     int64  `json:"round"`
	Sig       string `json:"sig"`
	Sender    int64  `json:"sender"`
	Recipient int64  `json:"recipient"`
	Amount    int64  `json:"amount"`
}

type EventDataWriter interface {
	AppendJSONL(v any) error
}

type EventJsonlWriter struct {
	path string
}

func NewEventJsonlWriter(path string) *EventJsonlWriter {
	return &EventJsonlWriter{path: path}
}

func (w *EventJsonlWriter) AppendJSONL(v any) error {
	f, err := os.OpenFile(w.path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	return enc.Encode(v)
}
