package models

import "math"

// Metrics keeps running stats; Average is derived (Sum/Count).
type Metrics struct {
	Count     int64 `json:"count"`
	Sum       int64 `json:"sum"`
	Min       int64 `json:"min"`
	Max       int64 `json:"max"`
	LastRound int64 `json:"last_round"`
}

func NewMetrics() Metrics {
	return Metrics{Min: math.MaxInt64}
}

func (m *Metrics) Update(amount int64, round int64) {
	m.Count++
	m.Sum += amount
	if amount < m.Min {
		m.Min = amount
	}
	if amount > m.Max {
		m.Max = amount
	}
	if round > m.LastRound {
		m.LastRound = round
	}
}

func (m *Metrics) Average() float64 {
	if m.Count == 0 {
		return 0
	}
	return float64(m.Sum) / float64(m.Count)
}
