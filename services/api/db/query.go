package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// boilerplate code, same as in market-data/db/store.go

type Store struct {
	pool *pgxpool.Pool
}

func New(ctx context.Context, dbURL string) (*Store, error) {
	pool, err := pgxpool.New(ctx, dbURL)

	if err != nil {
		return nil, fmt.Errorf("failed to connect to db: %w", err)
	}
	return &Store{pool: pool}, nil
}

func (s *Store) Close() {
	s.pool.Close()
}

// return types for QUERY
type Ticker struct {
	Time             time.Time `json:"time"`
	Symbol           string    `json:"symbol"`
	LastPrice        float64   `json:"last_price"`
	OpenPrice        float64   `json:"open_price"`
	High             float64   `json:"high"`
	Low              float64   `json:"low"`
	Volume           float64   `json:"volume"`
	QuoteVolume      float64   `json:"quote_volume"`
	WeightedAvgPrice float64   `json:"weighted_avg_price"`
	TradeCount       int64     `json:"trade_count"`
}

type Trade struct {
	Time     time.Time `json:"time"`
	Symbol   string    `json:"symbol"`
	Price    float64   `json:"price"`
	Quantity float64   `json:"quantity"`
	IsMaker  bool      `json:"is_maker"`
	TradeID  int64     `json:"trade_id"`
}

func (s *Store) GetSymbol(ctx context.Context) ([]string, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT DISTINCT symbol FROM tickers ORDER BY symbol`)

	if err != nil {
		return nil, fmt.Errorf("failed to query symbols: %w", err)
	}

	defer rows.Close()

	var symbols []string
	for rows.Next() {
		var symbol string
		if err := rows.Scan(&symbol); err != nil {
			return nil, err
		}
		symbols = append(symbols, symbol)
	}
	return symbols, nil
}

func (s *Store) GetTicker(ctx context.Context) ([]string, error) {
	return nil, nil
}

func (s *Store) GetTrade(ctx context.Context) ([]string, error) {
	return nil, nil
}
