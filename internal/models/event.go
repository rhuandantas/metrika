package models

type Event struct {
	Round     int64  `json:"round"`
	Sig       string `json:"sig"`
	Sender    int64  `json:"sender"`
	Recipient int64  `json:"recipient"`
	Amount    int64  `json:"amount"`
}
