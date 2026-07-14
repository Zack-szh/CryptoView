package maintainer

import (
	"context"
	"log"
	"time"

	"github.com/szh/cryptoview/services/market-data/binance"
	"github.com/szh/cryptoview/services/market-data/db"
)

type Config struct {
	Symbols     []string
	Intervals   []string
	HistoryDays int
}

func Run(ctx context.Context, store *db.Store, cfg Config) {
	log.Printf("[maintainer] starting — history=%dd symbols=%v intervals=%v",
		cfg.HistoryDays, cfg.Symbols, cfg.Intervals)

	historyStart := time.Now().UTC().AddDate(0, 0, -cfg.HistoryDays)

	runBackfill(ctx, store, cfg, historyStart)

	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			log.Printf("[maintainer] hourly gap check")
			runGapCheck(ctx, store, cfg, historyStart)
		case <-ctx.Done():
			log.Printf("[maintainer] shutting down")
			return
		}
	}
}

func runBackfill(ctx context.Context, store *db.Store, cfg Config, historyStart time.Time) {
	for _, symbol := range cfg.Symbols {
		for _, interval := range cfg.Intervals {
			if ctx.Err() != nil {
				return
			}
			now := time.Now().UTC()
			log.Printf("[maintainer] backfill %s/%s: %s → %s",
				symbol, interval, historyStart.Format(time.DateOnly), now.Format(time.DateOnly))
			if err := backfill(ctx, store, symbol, interval, historyStart, now); err != nil {
				log.Printf("[maintainer] %s/%s backfill error: %v", symbol, interval, err)
			}
		}
	}
}

func runGapCheck(ctx context.Context, store *db.Store, cfg Config, historyStart time.Time) {
	for _, symbol := range cfg.Symbols {
		for _, interval := range cfg.Intervals {
			if ctx.Err() != nil {
				return
			}
			gaps, err := store.FindGaps(ctx, symbol, interval, historyStart)
			if err != nil {
				log.Printf("[maintainer] gap query %s/%s: %v", symbol, interval, err)
				continue
			}
			for _, gap := range gaps {
				log.Printf("[maintainer] gap fill %s/%s: %s → %s",
					symbol, interval, gap.Start.Format(time.RFC3339), gap.End.Format(time.RFC3339))
				if err := backfill(ctx, store, symbol, interval, gap.Start, gap.End); err != nil {
					log.Printf("[maintainer] gap fill error: %v", err)
				}
			}
		}
	}
}

func backfill(ctx context.Context, store *db.Store, symbol, interval string, start, end time.Time) error {
	klines, err := binance.FillKlines(ctx, symbol, interval, start, end)
	if err != nil {
		return err
	}
	for _, k := range klines {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if err := store.InsertRestKline(ctx, k); err != nil {
			log.Printf("[maintainer] insert error: %v", err)
		}
	}
	log.Printf("[maintainer] inserted %d candles for %s/%s", len(klines), symbol, interval)
	return nil
}
