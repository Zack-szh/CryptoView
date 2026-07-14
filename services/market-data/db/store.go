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
		`INSERT INTO trades (time, symbol, price, quantity, is_maker, trade_id)
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

func (s *Store) InsertTicker(ctx context.Context, ticker binance.TickerEvent) error {
	_, err := s.pool.Exec(ctx,
		`INSERT INTO tickers (time, symbol, last_price, open_price, 
	 high, low, volume, quote_volume, weighted_avg_price, trade_count)
	 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	 ON CONFLICT DO NOTHING`,
		time.UnixMilli(ticker.EventTime),
		ticker.Symbol,
		ticker.LastPrice,
		ticker.OpenPrice,
		ticker.High,
		ticker.Low,
		ticker.Volume,
		ticker.QuoteVolume,
		ticker.WeightedAvgPrice,
		ticker.TradeCount,
	)

	if err != nil {
		return fmt.Errorf("failed to insert ticker: %w", err)
	}

	return nil
}

func (s *Store) InsertKline(ctx context.Context, kline binance.KlineEvent) error {
	_, err := s.pool.Exec(ctx,
		`INSERT INTO klines (open_time, close_time, symbol, interval, open, high, low,
		close, volume, trade_count, is_closed)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (open_time, symbol, interval) DO UPDATE SET
			close_time  = EXCLUDED.close_time,
			high        = EXCLUDED.high,
			low         = EXCLUDED.low,
			close       = EXCLUDED.close,
			volume      = EXCLUDED.volume,
			trade_count = EXCLUDED.trade_count,
			is_closed   = EXCLUDED.is_closed`,
		time.UnixMilli(kline.Kline.OpenTime),
		time.UnixMilli(kline.Kline.CloseTime),
		kline.Kline.Symbol,
		kline.Kline.Interval,
		kline.Kline.OpenPrice,
		kline.Kline.High,
		kline.Kline.Low,
		kline.Kline.ClosePrice,
		kline.Kline.Volume,
		kline.Kline.TradeCount,
		kline.Kline.IsClosed,
	)

	if err != nil {
		return fmt.Errorf("failed to insert kline: %w", err)
	}

	return nil
}

// insert historical kline to fill gap, from Binance REST endpoints
func (s *Store) InsertRestKline(ctx context.Context, k binance.RestKline) error {
	_, err := s.pool.Exec(ctx,
		`INSERT INTO klines (open_time, close_time, symbol, interval, open, high, low,
              close, volume, trade_count, is_closed)
              VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, true)
              ON CONFLICT (open_time, symbol, interval) DO UPDATE SET
                      close_time  = EXCLUDED.close_time,
                      high        = EXCLUDED.high,
                      low         = EXCLUDED.low,
                      close       = EXCLUDED.close,
                      volume      = EXCLUDED.volume,
                      trade_count = EXCLUDED.trade_count,
                      is_closed   = EXCLUDED.is_closed`,
		k.OpenTime, k.CloseTime, k.Symbol, k.Interval,
		k.Open, k.High, k.Low, k.Close, k.Volume, k.TradeCount,
	)
	if err != nil {
		return fmt.Errorf("failed to insert rest kline: %w", err)
	}
	return nil
}
