package binance

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// this file implements REST endpoings to binance.us
// should work together with websocket to backfill historical data

const BaseURL = "https://api.binance.us"

// IMPORTANT!!! Klines are uniquely identified by their open time

type RestKline struct {
	OpenTime   time.Time
	CloseTime  time.Time
	Symbol     string
	Interval   string
	Open       string
	High       string
	Low        string
	Close      string
	Volume     string
	TradeCount int64
}

type OrderBook struct {
	LastUpdateID int64      `json:"lastUpdateId"`
	Bids         [][]string `json:"bids"`
	Asks         [][]string `json:"asks"`
}

// for binance REST endpoints, we can fetch up to 1000 candles per request
func fetchKlines(ctx context.Context, symbol, interval string, start, end time.Time) ([]RestKline, error) {
	params := url.Values{}
	params.Set("symbol", symbol)
	params.Set("interval", interval)
	params.Set("startTime", strconv.FormatInt(start.UnixMilli(), 10))
	params.Set("endTime", strconv.FormatInt(end.UnixMilli(), 10))
	params.Set("limit", "1000")

	//resp, err := http.Get(BaseURL + "/api/v3/klines?" + params.Encode())
	// fetch with context
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, BaseURL+"/api/v3/klines?"+params.Encode(), nil)

	if err != nil {
		return nil, fmt.Errorf("http request: %w", err)
	}
	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("http do: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected response status: %d", resp.StatusCode)
	}
	defer resp.Body.Close()

	// decode message
	var raw [][]json.RawMessage
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("failed to decode json: %w", err)
	}

	klines := make([]RestKline, 0, len(raw))

	for _, row := range raw {
		// response should have 11 fields
		if len(row) < 10 {
			continue
		}

		var open_time, close_time, trade_count int64
		var open, close, high, low, volume string

		if err := json.Unmarshal(row[0], &open_time); err != nil {
			log.Printf("failed to unmarshal open_time: %v", err)
			continue
		}
		if err := json.Unmarshal(row[6], &close_time); err != nil {
			log.Printf("failed to unmarshal cloes_time: %v", err)
			continue
		}
		if err := json.Unmarshal(row[8], &trade_count); err != nil {
			log.Printf("failed to unmarshal trade_count: %v", err)
			continue
		}
		json.Unmarshal(row[1], &open)
		json.Unmarshal(row[2], &high)
		json.Unmarshal(row[3], &low)
		json.Unmarshal(row[4], &close)
		json.Unmarshal(row[5], &volume)

		klines = append(klines, RestKline{
			OpenTime:   time.UnixMilli(open_time).UTC(),
			CloseTime:  time.UnixMilli(close_time).UTC(),
			Symbol:     symbol,
			Interval:   interval,
			Open:       open,
			Close:      close,
			High:       high,
			Low:        low,
			Volume:     volume,
			TradeCount: trade_count,
		})
	}
	return klines, nil
}

// fillKline should keep calling fetchKlines until full [start, end] window is covered
func FillKlines(ctx context.Context, symbol, interval string, start, end time.Time) ([]RestKline, error) {
	var all []RestKline
	var timePtr = start

	for timePtr.Before(end) {
		if ctx.Err() != nil {
			return all, ctx.Err()
		}

		batch, err := fetchKlines(ctx, symbol, interval, timePtr, end)
		if err != nil {
			return nil, err
		}

		if len(batch) == 0 {
			// if len(batch) == 0, that means we have collected all candles in time window
			break
		}

		all = append(all, batch...)
		// advance timePtr, we add 1 ms on top to avoid overlap
		timePtr = batch[len(batch)-1].OpenTime.Add(time.Millisecond)
		// sleep due to binanace rate limit
		time.Sleep(200 * time.Millisecond)
	}
	return all, nil
}

// GET /api/v3/depth
func FetchOrderBook(ctx context.Context, symbol string, limit int) (*OrderBook, error) {
	params := url.Values{}
	params.Set("symbol", symbol)
	params.Set("limit", strconv.Itoa(limit))

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, BaseURL+"/api/v3/depth?"+params.Encode(), nil)

	if err != nil {
		return nil, fmt.Errorf("http request: %w", err)
	}
	resp, err := http.DefaultClient.Do(request)

	if err != nil {
		return nil, fmt.Errorf("http do: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected response status: %d", resp.StatusCode)
	}
	defer resp.Body.Close()

	var book OrderBook
	if err := json.NewDecoder(resp.Body).Decode(&book); err != nil {
		return nil, fmt.Errorf("failed to decode order book: %w", err)
	}

	return &book, nil
}
