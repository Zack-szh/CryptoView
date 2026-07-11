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

	symbols := make([]string, 0)

	for rows.Next() {
		var symbol string
		if err := rows.Scan(&symbol); err != nil {
			return nil, err
		}
		symbols = append(symbols, symbol)
	}
	return symbols, nil
}

func (s *Store) GetTicker(ctx context.Context, symbol string, limit int) ([]Ticker, error) {
	// return last limit entries of ticker given a symbol
	rows, err := s.pool.Query(ctx,
		`SELECT time, symbol, last_price, open_price, high, low, volume, 
		quote_volume, weighted_avg_price, trade_count 
		FROM tickers WHERE symbol = $1 ORDER BY time DESC LIMIT $2`,
		symbol, limit)

	if err != nil {
		return nil, fmt.Errorf("failed to query tickers: %s: %v", symbol, err)
	}

	defer rows.Close()

	tickers := make([]Ticker, 0)

	for rows.Next() {
		var t Ticker
		if err := rows.Scan(&t.Time, &t.Symbol, &t.LastPrice, &t.OpenPrice, &t.High,
			&t.Low, &t.Volume, &t.QuoteVolume, &t.WeightedAvgPrice, &t.TradeCount); err != nil {
			return nil, err
		}
		tickers = append(tickers, t)
	}

	return tickers, nil
}

func (s *Store) GetTrade(ctx context.Context, symbol string, limit int) ([]Trade, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT time, symbol, price, quantity, is_maker, trade_id FROM 
		trades WHERE symbol = $1 ORDER BY time DESC LIMIT $2`,
		symbol, limit)

	if err != nil {
		return nil, fmt.Errorf("failed to query trade: %s: %v", symbol, err)
	}

	defer rows.Close()

	trades := make([]Trade, 0)

	for rows.Next() {
		var t Trade
		if err := rows.Scan(&t.Time, &t.Symbol, &t.Price, &t.Quantity, &t.IsMaker, &t.TradeID); err != nil {
			return nil, err
		}

		trades = append(trades, t)
	}
	return trades, nil
}
