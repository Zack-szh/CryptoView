package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/szh/cryptoview/services/market-data/binance"
)

// we are using PGX connection pool
// instead of opening a new db connection on every query
// pgx pool keeps a pool of open connection ready, reuses them efficiently

// wrapper
type Store struct {
	pool *pgxpool.Pool
}

// open new connection, either returns a connection pool or error
func New(ctx context.Context, dbURL string) (*Store, error) {
	pool, err := pgxpool.New(ctx, dbURL)

	if err != nil {
		// if cant connect to db, return error, no conn
		return nil, fmt.Errorf("failed to connect to db: %w", err)
	}
	// connection successful, return connection pool, no error
	return &Store{pool: pool}, nil
}

func (s *Store) Close() {
	// this calls pgxpool.Pool.Close()
	s.pool.Close()
}

func (s *Store) InsertTrade(ctx context.Context, trade binance.TradeEvent) error {
	_, err := s.pool.Exec(ctx,
		`INSERT INTO trades (time, ticker, price, quantity, is_maker, trade_id)
         VALUES ($1, $2, $3, $4, $5, $6)
         ON CONFLICT DO NOTHING`,
		time.UnixMilli(trade.TradeTime),
		trade.Symbol,
		trade.Price,
		trade.Quantity,
		trade.IsMarketMaker,
		trade.TradeID,
	)
	if err != nil {
		return fmt.Errorf("failed to insert trade: %w", err)
	}
	return nil
}
