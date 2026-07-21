// this is the orchestrator for computing indicators
package indicators

import (
	"fmt"
	"time"
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
