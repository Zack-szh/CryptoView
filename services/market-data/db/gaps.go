package db

import (
	"context"
	"fmt"
	"time"
)

// this file finds the gap in kline data entries from local DB
// the gap [start, end] is then passed to fetchKlines to backfill historical candles

type Gap struct {
	Start time.Time
	End   time.Time
}

// convert binance interval to Postgres Interval
var intervalSteps = map[string]string{
	"1m":  "1 minute",
	"3m":  "3 minutes",
	"5m":  "5 minutes",
	"15m": "15 minutes",
	"30m": "30 minutes",
	"1h":  "1 hour",
	"2h":  "2 hours",
	"4h":  "4 hours",
	"6h":  "6 hours",
	"12h": "12 hours",
	"1d":  "1 day",
}

// FIndGaps() would look for missing windows of data starting from [since, present_time] using LEAD() window query
// NOTE: here we return a list of Gaps
func (s *Store) FindGaps(ctx context.Context, symbol, interval string, since time.Time) ([]Gap, error) {
	step, ok := intervalSteps[interval]
	if !ok {
		return nil, fmt.Errorf("unknown interval: %s", interval)
	}

	// Notice the use of LEAD() to peek at neighboring row
	// if neightbor_open_time - current_open_time > interval, that means we are missing candles
	query := fmt.Sprintf(`
              SELECT
                      open_time + interval '%s' AS gap_start,
                      next_time               AS gap_end
              FROM (
                      SELECT
                              open_time,
                              LEAD(open_time) OVER (ORDER BY open_time) AS next_time
                      FROM klines
                      WHERE symbol = $1 AND interval = $2 AND open_time >= $3
              ) t
              WHERE next_time - open_time > interval '%s'
              ORDER BY gap_start`, step, step)

	rows, err := s.pool.Query(ctx, query, symbol, interval, since)
	if err != nil {
		return nil, fmt.Errorf("gap query: %w", err)
	}
	defer rows.Close()

	var gaps []Gap
	for rows.Next() {
		var g Gap
		if err := rows.Scan(&g.Start, &g.End); err != nil {
			return nil, err
		}
		gaps = append(gaps, g)
	}
	return gaps, nil
}
