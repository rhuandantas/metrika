package repository

import (
	"context"
	"database/sql"

	"github.com/rhuandantas/metrika/internal/models"
	_ "modernc.org/sqlite"
)

type Repository interface {
	// SaveMetrics persists the current metrics state.
	SaveMetrics(ctx context.Context, metrics models.Metrics) error
	// LoadMetrics retrieves the current metrics state.
	LoadMetrics(ctx context.Context) (models.Metrics, error)
	// Init initializes the database schema if not exists.
	Init(ctx context.Context) error
}

type SQLiteMetrics struct{ db *sql.DB }

func NewSQLiteMetrics(dsn string) (Repository, error) {
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	// SQLite can only handle one writer at a time.
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(0)

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &SQLiteMetrics{db: db}, nil
}

func (s *SQLiteMetrics) Init(ctx context.Context) error {
	// Use WAL journal mode for better concurrency.
	stmts := []string{
		`PRAGMA journal_mode=WAL;`,
		`CREATE TABLE IF NOT EXISTS metrics(
			id INTEGER PRIMARY KEY CHECK(id=1),
			last_round INTEGER,
			count INTEGER,
			sum INTEGER,
			min INTEGER,
			max INTEGER
			);`,
		`INSERT OR IGNORE INTO metrics(id, last_round, count, sum, min, max) VALUES(1, 0, 0, 0, 9223372036854775807, 0);`,
	}
	for _, q := range stmts {
		if _, err := s.db.ExecContext(ctx, q); err != nil {
			return err
		}
	}
	return nil
}

func (s *SQLiteMetrics) LoadMetrics(ctx context.Context) (models.Metrics, error) {
	row := s.db.QueryRowContext(ctx, `SELECT count,sum,min,max,last_round FROM metrics WHERE id=1`)
	m := models.NewMetrics()
	if err := row.Scan(&m.Count, &m.Sum, &m.Min, &m.Max, &m.LastRound); err != nil {
		return models.Metrics{}, err
	}
	return m, nil
}

func (s *SQLiteMetrics) SaveMetrics(ctx context.Context, m models.Metrics) error {
	_, err := s.db.ExecContext(ctx, `UPDATE metrics SET count=?, sum=?, min=?, max=?, last_round=? WHERE id=1`, m.Count, m.Sum, m.Min, m.Max, m.LastRound)
	return err
}
