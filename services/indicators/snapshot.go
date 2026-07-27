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

// like Snapshot, but holds a full time-aligned series per indicator instead of a single
// latest value — meant for plotting on the candle chart rather than a stats sidebar.
// each indicator slice is index-aligned with Times; entries are nil until that
// indicator has enough history to compute (e.g. SMA50 is nil for the first 49 points)
type IndicatorSeries struct {
	Symbol   string      `json:"symbol"`
	Interval string      `json:"interval"`
	Times    []time.Time `json:"times"`

	SMA20 []*float64 `json:"sma_20"`
	SMA50 []*float64 `json:"sma_50"`
	EMA12 []*float64 `json:"ema_12"`
	EMA26 []*float64 `json:"ema_26"`

	BollingerMid   []*float64 `json:"bollinger_mid"`
	BollingerUpper []*float64 `json:"bollinger_upper"`
	BollingerLower []*float64 `json:"bollinger_lower"`

	VWAP []*float64 `json:"vwap"`

	MACD       []*float64 `json:"macd"`
	MACDSignal []*float64 `json:"macd_signal"`
	MACDHist   []*float64 `json:"macd_hist"`

	RSI14 []*float64 `json:"rsi_14"`
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

// every *Series function in indicators.go is trailing-aligned: it always ends at the
// last input point, it just starts late depending on how much warmup the indicator needs.
// so a series can always be lined up against the full input by right-aligning it and
// leaving the front padded with nil (not yet computable)
func alignSeries(times []time.Time, series []float64, ok bool) []*float64 {
	aligned := make([]*float64, len(times))

	if !ok {
		return aligned
	}

	offset := len(times) - len(series)
	for i, v := range series {
		v := v
		aligned[offset+i] = &v
	}
	return aligned
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

// like BuildSnapshot, but returns the full history of each chartable indicator instead of
// just the latest value. order book imbalance is intentionally excluded — it's a read of the
// live order book, not something derived from kline history, so there's no series to build.
// 7/27/2026 update: changed GetKlineLimit to GetKline
// this way both kline and indicator lines cover the same historical window
// so for now GetKlineLimit is just for ease of computing the indicator snapshot
func BuildIndicatorSeries(ctx context.Context, store *db.Store, symbol, interval string, since time.Time) (*IndicatorSeries, error) {
	klines, err := store.GetKline(ctx, symbol, interval, since)

	if err != nil {
		return nil, fmt.Errorf("features: fetch klines: %w", err)
	}

	if len(klines) == 0 {
		return nil, fmt.Errorf("features: no kline data for %s/%s", symbol, interval)
	}

	// only use closed candles for indicator math
	closedKlines := make([]db.Kline, 0, len(klines))
	for _, k := range klines {
		if k.IsClosed {
			closedKlines = append(closedKlines, k)
		}
	}

	times := make([]time.Time, len(closedKlines))
	highs := make([]float64, len(closedKlines))
	lows := make([]float64, len(closedKlines))
	closes := make([]float64, len(closedKlines))
	volumes := make([]float64, len(closedKlines))

	for i, k := range closedKlines {
		// open time, not close time — must match how the frontend keys candles on the
		// time axis (KlineChart.tsx plots each candle at its open_time), so a given
		// indicator point lines up under the candle it was computed from
		times[i] = k.OpenTime
		highs[i] = k.High
		lows[i] = k.Low
		closes[i] = k.ClosePrice
		volumes[i] = k.Volume
	}

	series := &IndicatorSeries{
		Symbol:   symbol,
		Interval: interval,
		Times:    times,
	}

	sma20, ok := SMASeries(closes, SMAShort)
	series.SMA20 = alignSeries(times, sma20, ok)

	sma50, ok := SMASeries(closes, SMALong)
	series.SMA50 = alignSeries(times, sma50, ok)

	ema12, ok := EMASeries(closes, EMAFast)
	series.EMA12 = alignSeries(times, ema12, ok)

	ema26, ok := EMASeries(closes, EMASlow)
	series.EMA26 = alignSeries(times, ema26, ok)

	bollMid, bollUp, bollLow, ok := BollingerSeries(closes, BollingerPeriod, BollingerStdDev)
	series.BollingerMid = alignSeries(times, bollMid, ok)
	series.BollingerUpper = alignSeries(times, bollUp, ok)
	series.BollingerLower = alignSeries(times, bollLow, ok)

	vwap, ok := VWAPSeries(highs, lows, closes, volumes, VWAPWindow)
	series.VWAP = alignSeries(times, vwap, ok)

	macd, macdSignal, macdHist, ok := MACDSeries(closes, MACDFast, MACDSlow, MACDSignal)
	series.MACD = alignSeries(times, macd, ok)
	series.MACDSignal = alignSeries(times, macdSignal, ok)
	series.MACDHist = alignSeries(times, macdHist, ok)

	rsi, ok := RSISeries(closes, RSIPeriod)
	series.RSI14 = alignSeries(times, rsi, ok)

	return series, nil
}
