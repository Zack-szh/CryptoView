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
