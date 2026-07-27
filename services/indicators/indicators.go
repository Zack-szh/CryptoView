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

// returns SMA at every valid position (index >= period-1 of the input)
func SMASeries(values []float64, period int) (series []float64, ok bool) {
	if period <= 0 || len(values) < period {
		return nil, false
	}

	series = make([]float64, len(values)-period+1)
	for i := range series {
		series[i] = mean(values[i : i+period])
	}
	return series, true
}

// in financial analysis, period is just the number of datapoints to consider
// returns simple moving average of last 'period' datapoints
func SMA(values []float64, period int) (avg float64, ok bool) {
	series, ok := SMASeries(values, period)

	if !ok {
		return 0, false
	}
	return series[len(series)-1], true
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

// Wilder-smoothed RSI series, one value per delta once at least 'period' deltas have been smoothed
func RSISeries(values []float64, period int) (series []float64, ok bool) {
	// RSI captures delta between datapoints
	// so we do len(values) < period + 1
	if period <= 0 || len(values) < period+1 {
		return nil, false
	}

	deltas := make([]float64, len(values)-1)
	for i := 0; i < len(values)-1; i++ {
		deltas[i] = values[i+1] - values[i]
	}

	var avgGain, avgLoss float64
	series = make([]float64, len(deltas)-period+1)

	for i, d := range deltas {
		gain, loss := 0.0, 0.0
		if d > 0 {
			gain = d
		} else {
			loss = -d
		}
		avgGain = (avgGain*float64(period-1) + gain) / float64(period)
		avgLoss = (avgLoss*float64(period-1) + loss) / float64(period)

		if i+1 < period {
			continue
		}

		var rsi float64
		if avgLoss == 0 {
			if avgGain == 0 {
				// case 1: no gain and no loss
				rsi = 50
			} else {
				// case 2: gain exists, no loss
				rsi = 100
			}
		} else {
			// case 3: normal calculation
			rs := avgGain / avgLoss
			rsi = 100 - 100/(1+rs)
		}
		series[i+1-period] = rsi
	}

	return series, true
}

// Wilder-smoothed RSI over 'period'
func RSI(values []float64, period int) (rsi float64, ok bool) {
	series, ok := RSISeries(values, period)

	if !ok {
		return 0, false
	}
	return series[len(series)-1], true
}

// MACD = EMA_fast - EMA_slow, returned as full macd/signal/histogram series
// default: fast = 12, slow = 29, signal=9
func MACDSeries(values []float64, fast, slow, signal int) (macdLine, signalLine, histogram []float64, ok bool) {
	fastSeries, ok1 := EMASeries(values, fast)
	slowSeries, ok2 := EMASeries(values, slow)

	if !ok1 || !ok2 {
		return nil, nil, nil, false
	}

	offset := len(fastSeries) - len(slowSeries)
	fullMacdLine := make([]float64, len(slowSeries))

	for i := range slowSeries {
		fullMacdLine[i] = fastSeries[i+offset] - slowSeries[i]
	}
	// then we do EMASeries on macdLine to get signalLine
	// macdLine: what direction is the momentum moving in right now
	// signalLine: a smoothed macdLine behaviour of momentum on average recently
	// if MACD = 2.5, signal = 2, that means:
	// 		current momentum (macd) is stronger than its recent average (signal)
	signalSeries, ok3 := EMASeries(fullMacdLine, signal)

	if !ok3 {
		return nil, nil, nil, false
	}

	// signalSeries is shorter than fullMacdLine (it needs its own warmup),
	// so trim macdLine down to line up index-for-index with signalLine
	macdOffset := len(fullMacdLine) - len(signalSeries)
	macdLine = fullMacdLine[macdOffset:]
	signalLine = signalSeries

	// histogram is the difference between MACD line and signal line
	// basically the delta away from average momentum at a given time
	histogram = make([]float64, len(signalLine))
	for i := range signalLine {
		histogram[i] = macdLine[i] - signalLine[i]
	}

	return macdLine, signalLine, histogram, true
}

// MACD = EMA_fast - EMA_slow
// default: fast = 12, slow = 29, signal=9
func MACD(values []float64, fast, slow, signal int) (macd, signalLine, histogram float64, ok bool) {
	macdSeries, signalSeries, histSeries, ok := MACDSeries(values, fast, slow, signal)

	if !ok {
		return 0, 0, 0, false
	}

	n := len(signalSeries)
	return macdSeries[n-1], signalSeries[n-1], histSeries[n-1], true
}

// BollingerBandsL middle = SMA(period), bands = middle +/- numStd * population std
// defaults: period=20, numStdDev=2.0
// this measures price volatility around a moving average
// returns full middle/upper/lower series at every valid position (index >= period-1)
func BollingerSeries(values []float64, period int, numStdDev float64) (middle, upper, lower []float64, ok bool) {
	if period <= 0 || len(values) < period {
		return nil, nil, nil, false
	}

	n := len(values) - period + 1
	middle = make([]float64, n)
	upper = make([]float64, n)
	lower = make([]float64, n)

	for i := 0; i < n; i++ {
		window := values[i : i+period]
		m := mean(window)
		sd := std(window, m)

		middle[i] = m
		upper[i] = m + numStdDev*sd
		lower[i] = m - numStdDev*sd
	}

	return middle, upper, lower, true
}

func BollingerBands(values []float64, period int, numStdDev float64) (middle, upper, lower float64, ok bool) {
	middleSeries, upperSeries, lowerSeries, ok := BollingerSeries(values, period, numStdDev)

	if !ok {
		return 0, 0, 0, false
	}

	n := len(middleSeries)
	return middleSeries[n-1], upperSeries[n-1], lowerSeries[n-1], true
}

// population std of log returns over the last `window`
// gives annualized volatility via barsPerYear (pass 1 for per-bar vol)
// returns full series at every valid position (needs `window`+1 closes per point)
func RealizedVolatilitySeries(values []float64, window int, barsPerYear float64) (series []float64, ok bool) {
	if window <= 0 || len(values) < window+1 {
		return nil, false
	}

	returns := make([]float64, len(values)-1)
	for i := 1; i < len(values); i++ {
		returns[i-1] = math.Log(values[i] / values[i-1])
	}

	n := len(returns) - window + 1
	series = make([]float64, n)

	for i := 0; i < n; i++ {
		sub := returns[i : i+window]
		m := mean(sub)
		sd := std(sub, m)
		series[i] = sd * math.Sqrt(barsPerYear)
	}

	return series, true
}

func RealizedVolatility(values []float64, window int, barsPerYear float64) (vol float64, ok bool) {
	series, ok := RealizedVolatilitySeries(values, window, barsPerYear)

	if !ok {
		return 0, false
	}
	return series[len(series)-1], true
}

// VWAP: average price on an asset traded at, weighted by how much volume is traded at each price level
// over the last 'window' bars
// ALL SLICES SHOULD BE THE SMAE LENGTH, ASCENDING TIME ORDER
// returns full rolling VWAP series at every valid position (index >= window-1)
func VWAPSeries(highs, lows, closes, volumes []float64, window int) (series []float64, ok bool) {
	n := len(closes)

	if window <= 0 || n < window || n != len(highs) || n != len(lows) || n != len(volumes) {
		return nil, false
	}

	typical := make([]float64, n)
	for i := 0; i < n; i++ {
		typical[i] = (highs[i] + lows[i] + closes[i]) / 3
	}

	series = make([]float64, n-window+1)
	for i := range series {
		var num, denominator float64
		for j := i; j < i+window; j++ {
			num += typical[j] * volumes[j]
			denominator += volumes[j]
		}

		if denominator == 0 {
			return nil, false
		}
		series[i] = num / denominator
	}

	return series, true
}

func VWAP(highs, lows, closes, volumes []float64, window int) (vwap float64, ok bool) {
	series, ok := VWAPSeries(highs, lows, closes, volumes, window)

	if !ok {
		return 0, false
	}
	return series[len(series)-1], true
}

// calculates order book imbalance: measures whether more selling or buying pressure in the order book
// no series version: this operates on order book depth levels (bidQty/askQty), not time-indexed
// kline data, and no historical order book snapshots are stored — it is inherently a single
// point-in-time read of the live book
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
