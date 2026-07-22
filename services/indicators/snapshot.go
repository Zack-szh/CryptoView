// this is the orchestrator for computing indicators
package indicators

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/szh/cryptoview/services/api/db"
	"github.com/szh/cryptoview/services/market-data/binance"
)

// default values for technical indicators

const (
	LookbackBars    = 200 // ONLY CLOSED BARS
	SMAShort        = 20
	SMALong         = 50
	EMAFast         = 12
	EMASlow         = 26
	RSIPeriod       = 14
	MACDFast        = 12
	MACDSlow        = 26
	MACDSignal      = 9
	BollingerPeriod = 20
	BollingerStdDev = 2.0
	VolWindow       = 20
	VWAPWindow      = 20
	OrderBookLevels = 20
)

type MACDResult struct {
	MACD      float64 `json:"macd"`
	Signal    float64 `json:"signal"`
	Histogram float64 `json:"histogram"`
}

type BollingerResult struct {
	Middle float64 `json:"middle"`
	Upper  float64 `json:"upper"`
	Lower  float64 `json:"lower"`
}

// all indicators to be calculated and displayed
// IMPORTANT!! Timescale is determined by Interval!
type Snapshot struct {
	Symbol             string           `json:"symbol"`
	Interval           string           `json:"interval"`
	Time               time.Time        `json:"time"`
	LastPrice          float64          `json:"last_price"`
	SMA20              *float64         `json:"sma_20"`
	SMA50              *float64         `json:"sma_50"`
	EMA12              *float64         `json:"ema_12"`
	EMA26              *float64         `json:"ema_26"`
	RSI14              *float64         `json:"rsi_14"`
	MACD               *MACDResult      `json:"macd"`
	Bollinger          *BollingerResult `json:"bollinger"`
	RealizedVolatility *float64         `json:"realized_volatility"`
	VWAP               *float64         `json:"vwap"`
	OrderBookImbalance *float64         `json:"orderbook_imbalance"`
}

// for annualized realized vol
func barsPerYear(interval string) float64 {
	switch interval {
	case "1m":
		return 365 * 24 * 60
	case "5m":
		return 365 * 24 * 12
	case "15m":
		return 365 * 24 * 4
	case "30m":
		return 365 * 24 * 2
	case "1h":
		return 365 * 24
	case "4h":
		return 365 * 6
	case "1d":
		return 365
	default:
		return 365 * 24 * 60
	}
}

// extract quantity for each order book price level
func extractQty(levels [][]string) []float64 {
	qty := make([]float64, 0, len(levels))

	for _, lvl := range levels {
		if len(lvl) < 2 {
			continue
		}
		var q float64
		if _, err := fmt.Sscanf(lvl[1], "%f", &q); err != nil {
			continue
		}
		qty = append(qty, q)
	}

	return qty
}

func BuildSnapshot(ctx context.Context, store *db.Store, symbol, interval string) (*Snapshot, error) {
	klines, err := store.GetKlineLimit(ctx, symbol, interval, LookbackBars+1)

	if err != nil {
		return nil, fmt.Errorf("features: fetch klines: %w", err)
	}

	if len(klines) == 0 {
		return nil, fmt.Errorf("features: no kline data for %s/%s", symbol, interval)
	}

	last := klines[len(klines)-1]
	snap := &Snapshot{
		Symbol:    symbol,
		Interval:  interval,
		Time:      last.CloseTime,
		LastPrice: last.ClosePrice,
	}

	// only use closed candles for indicator math
	closedKlines := make([]db.Kline, 0, len(klines))
	for _, k := range klines {
		if k.IsClosed {
			closedKlines = append(closedKlines, k)
		}
	}

	highs := make([]float64, len(closedKlines))
	lows := make([]float64, len(closedKlines))
	closes := make([]float64, len(closedKlines))
	volumes := make([]float64, len(closedKlines))

	for i, k := range closedKlines {
		highs[i] = k.High
		lows[i] = k.Low
		closes[i] = k.ClosePrice
		volumes[i] = k.Volume
	}

	if v, ok := SMA(closes, SMAShort); ok {
		snap.SMA20 = &v
	}
	if v, ok := SMA(closes, SMALong); ok {
		snap.SMA50 = &v
	}
	if v, ok := EMA(closes, EMASlow); ok {
		snap.EMA26 = &v
	}
	if v, ok := EMA(closes, EMAFast); ok {
		snap.EMA12 = &v
	}
	if v, ok := RSI(closes, RSIPeriod); ok {
		snap.RSI14 = &v
	}
	if m, s, h, ok := MACD(closes, MACDFast, MACDSlow, MACDSignal); ok {
		snap.MACD = &MACDResult{MACD: m, Signal: s, Histogram: h}
	}
	if mid, up, low, ok := BollingerBands(closes, BollingerPeriod, BollingerStdDev); ok {
		snap.Bollinger = &BollingerResult{Middle: mid, Upper: up, Lower: low}
	}
	if v, ok := RealizedVolatility(closes, VolWindow, barsPerYear(interval)); ok {
		snap.RealizedVolatility = &v
	}
	if v, ok := VWAP(highs, lows, closes, volumes, VWAPWindow); ok {
		snap.VWAP = &v
	}

	// best-effort order book imbalance
	// if failed, the field would just be nil
	book, err := binance.FetchOrderBook(ctx, symbol, OrderBookLevels)
	if err != nil {
		log.Printf("[indicators] orderbook fetch failed for %s: %v", symbol, err)
	} else if book != nil {
		askQty := extractQty(book.Asks)
		bidQty := extractQty(book.Bids)
		if v, ok := OrderBookImbalance(bidQty, askQty, OrderBookLevels); ok {
			snap.OrderBookImbalance = &v
		}
	}

	return snap, nil
}
