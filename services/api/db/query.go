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
// IMPORTANT!!!!!
// THESE ARE RETURN TYPES FROM DATABASE QUERY
// NOT RETURN TYPES FROM BINANACE WEBSOCKET
// FOR WEBSOCKET RETURN TYPES LOOK IN binance/stream.go
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

type Kline struct {
	OpenTime   time.Time `json:"open_time"`
	CloseTime  time.Time `json:"close_time"`
	Symbol     string    `json:"symbol"`
	Interval   string    `json:"interval"`
	OpenPrice  float64   `json:"open"`
	High       float64   `json:"high"`
	Low        float64   `json:"low"`
	ClosePrice float64   `json:"close"`
	Volume     float64   `json:"volume"`
	TradeCount int64     `json:"trade_count"`
	IsClosed   bool      `json:"is_closed"`
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

func (s *Store) GetKline(ctx context.Context, symbol string, interval string, since time.Time) ([]Kline, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT open_time, close_time, symbol, interval, open, high, low, close, volume, trade_count, is_closed
		FROM klines WHERE symbol = $1 AND interval = $2 AND open_time >= $3
		ORDER BY open_time ASC`,
		symbol, interval, since)

	if err != nil {
		return nil, fmt.Errorf("failed to query kline: %s: %v", symbol, err)
	}

	defer rows.Close()

	klines := make([]Kline, 0)

	for rows.Next() {
		var k Kline
		if err := rows.Scan(&k.OpenTime, &k.CloseTime, &k.Symbol, &k.Interval, &k.OpenPrice, &k.High,
			&k.Low, &k.ClosePrice, &k.Volume, &k.TradeCount, &k.IsClosed); err != nil {
			return nil, err
		}

		klines = append(klines, k)
	}

	return klines, nil
}

// this getKline function always get the last 'limit' candles, instead of a time window (since)
func (s *Store) GetKlineLimit(ctx context.Context, symbol, interval string, limit int) ([]Kline, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT open_time, close_time, symbol, interval, open, high, low, close, volume, trade_count, is_closed
				FROM (
				SELECT open_time, close_time, symbol, interval, open, high, low, close, volume, trade_count, is_closed
				FROM klines WHERE symbol = $1 AND interval = $2
				ORDER BY open_time DESC LIMIT $3
				) sub ORDER BY open_time ASC`,
		symbol, interval, limit)

	if err != nil {
		return nil, fmt.Errorf("failed to query kline limit: %s: %v", symbol, err)
	}
	defer rows.Close()

	klines := make([]Kline, 0)
	for rows.Next() {
		var k Kline
		if err := rows.Scan(&k.OpenTime, &k.CloseTime, &k.Symbol, &k.Interval, &k.OpenPrice,
			&k.High, &k.Low, &k.ClosePrice, &k.Volume, &k.TradeCount, &k.IsClosed); err != nil {
			return nil, err
		}
		klines = append(klines, k)
	}
	return klines, nil
}
