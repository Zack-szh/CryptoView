// this file contains the formula for all financial indicators
// called by ./indicators/snapshot.go

package indicators

import "math"

func mean(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}

	var sum float64
	for _, val := range values {
		sum += val
	}
	return sum / float64(len(values))
}

func std(values []float64, mean float64) float64 {
	if len(values) == 0 {
		return 0
	}

	var sum float64
	for _, val := range values {
		diff := val - mean
		sum += diff * diff
	}

	variance := sum / float64(len(values))

	return math.Sqrt(variance)
}

// in financial analysis, period is just the number of datapoints to consider
// returns simple moving average of last 'period' datapoints
func SMA(values []float64, period int) (avg float64, ok bool) {
	if period <= 0 || len(values) < period {
		return 0, false
	}

	window := values[len(values)-period:]
	return mean(window), true
}

// returns EMA series seeded with the SMA of the first period
func EMASeries(values []float64, period int) (series []float64, ok bool) {
	if period <= 0 || len(values) < period {
		return nil, false
	}

	alpha := 2.0 / float64(period+1)
	seed := mean(values[:period])
	series = make([]float64, len(values)-period+1)
	series[0] = seed
	prev := seed

	for i := period; i < len(values); i++ {
		cur := (values[i]-prev)*alpha + prev
		series[i-period+1] = cur
		prev = cur
	}
	return series, true
}

// return the latest EMA value
func EMA(values []float64, period int) (ema float64, ok bool) {
	series, ok := EMASeries(values, period)

	if !ok {
		return 0, false
	}
	return series[len(series)-1], true
}

// Wilder-smoothed RSI over 'period'
func RSI(values []float64, period int) (rsi float64, ok bool) {
	// RSI captures delta between datapoints
	// so we do len(values) < period + 1
	if period <= 0 || len(values) < period+1 {
		return 0, false
	}

	deltas := make([]float64, len(values)-1)
	for i := 0; i < len(values)-1; i++ {
		deltas[i] = values[i+1] - values[i]
	}

	var avgGain, avgLoss float64

	for _, d := range deltas {
		gain, loss := 0.0, 0.0
		if d > 0 {
			gain = d
		} else {
			loss = -d
		}
		avgGain = (avgGain*float64(period-1) + gain) / float64(period)
		avgLoss = (avgLoss*float64(period-1) + loss) / float64(period)
	}

	if avgLoss == 0 {
		if avgGain == 0 {
			// case 1: no gain and no loss
			return 50, true
		}
		// case 2: gain exists, no loss
		return 100, true
	}
	// case 3: normal calculation
	rs := avgGain / avgLoss
	return 100 - 100/(1+rs), true
}

// MACD = EMA_fast - EMA_slow
// default: fast = 12, slow = 29, signal=9
func MACD(values []float64, fast, slow, signal int) (macd, signalLine, histogram float64, ok bool) {
	fastSeries, ok1 := EMASeries(values, fast)
	slowSeries, ok2 := EMASeries(values, slow)

	if !ok1 || !ok2 {
		return 0, 0, 0, false
	}

	offset := len(fastSeries) - len(slowSeries)
	macdLine := make([]float64, len(slowSeries))

	for i := range slowSeries {
		macdLine[i] = fastSeries[i+offset] - slowSeries[i]
	}
	// then we do EMASeries on macdLine to get signalLine
	// macdLine: what direction is the momentum moving in right now
	// signalLine: a smoothed macdLine behaviour of momentum on average recently
	// if MACD = 2.5, signal = 2, that means:
	// 		current momentum (macd) is stronger than its recent average (signal)
	signalSeries, ok3 := EMASeries(macdLine, signal)

	if !ok3 {
		return 0, 0, 0, false
	}

	m := macdLine[len(macdLine)-1]
	s := signalSeries[len(signalSeries)-1]

	// histogram is the difference between MACD line and signal line
	// basically the delta away from average momentum at a given time
	return m, s, m - s, true
}

// BollingerBandsL middle = SMA(period), bands = middle +/- numStd * population std
// defaults: period=20, numStdDev=2.0
// this measures price volatility around a moving average
func BollingerBands(values []float64, period int, numStdDev float64) (middle, upper, lower float64, ok bool) {
	if period <= 0 || len(values) < period {
		return 0, 0, 0, false
	}

	window := values[len(values)-period:]
	m := mean(window)
	sd := std(window, m)

	return m, m + numStdDev*sd, m - numStdDev*sd, true
}

// population std of log returns over the last `window`
// gives annualized volatility via barsPerYear (pass 1 for per-bar vol)
func RealizedVolatility(values []float64, window int, barsPerYear float64) (vol float64, ok bool) {
	if window <= 0 || len(values) < window+1 {
		return 0, false
	}

	tail := values[len(values)-window-1:]
	returns := make([]float64, window)

	for i := 1; i < len(tail); i++ {
		returns[i-1] = math.Log(tail[i] / tail[i-1])
	}

	m := mean(returns)
	sd := std(returns, m)

	return sd * math.Sqrt(barsPerYear), true
}

// VWAP: average price on an asset traded at, weighted by how much volume is traded at each price level
// over the last 'window' bars
// ALL SLICES SHOULD BE THE SMAE LENGTH, ASCENDING TIME ORDER
func VWAP(highs, lows, closes, volumes []float64, window int) (vwap float64, ok bool) {
	n := len(closes)

	if window <= 0 || n < window || n != len(highs) || n != len(lows) || n != len(volumes) {
		return 0, false
	}

	var num, denominator float64
	for i := n - window; i < n; i++ {
		typical := (highs[i] + lows[i] + closes[i]) / 3
		num += typical * volumes[i]
		denominator += volumes[i]
	}

	if denominator == 0 {
		return 0, false
	}

	return num / denominator, true
}

// calculates order book imbalance: measures whether more selling or buying pressure in the order book
func OrderBookImbalance(bidQty, askQty []float64, levels int) (imbalance float64, ok bool) {
	sum := func(q []float64, n int) float64 {
		if n > len(q) {
			n = len(q)
		}
		var s float64
		for _, v := range q[:n] {
			s += v
		}
		return s
	}

	bidSum := sum(bidQty, levels)
	askSum := sum(askQty, levels)

	if bidSum+askSum == 0 {
		return 0, false
	}

	return (bidSum - askSum) / (bidSum + askSum), true
}
