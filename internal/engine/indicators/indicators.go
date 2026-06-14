// Package indicators computes the technical signals used by the momentum/trend
// evaluator (SMA, RSI, trailing return, 52-week high) from a slice of closing
// prices ordered oldest-to-newest. Each function returns (value, ok); ok is
// false when there is insufficient history, so the caller fails the gate rather
// than acting on a spurious value.
package indicators

// SMA returns the simple moving average of the last n closes.
func SMA(closes []float64, n int) (float64, bool) {
	if n <= 0 || len(closes) < n {
		return 0, false
	}
	var sum float64
	for _, c := range closes[len(closes)-n:] {
		sum += c
	}
	return sum / float64(n), true
}

// RSI returns the Wilder Relative Strength Index over n periods (0–100). It needs
// at least n+1 closes.
func RSI(closes []float64, n int) (float64, bool) {
	if n <= 0 || len(closes) < n+1 {
		return 0, false
	}
	var gains, losses float64
	for i := len(closes) - n; i < len(closes); i++ {
		change := closes[i] - closes[i-1]
		if change >= 0 {
			gains += change
		} else {
			losses -= change
		}
	}
	if losses == 0 {
		return 100, true // no down moves -> maximally overbought
	}
	rs := (gains / float64(n)) / (losses / float64(n))
	return 100 - (100 / (1 + rs)), true
}

// ReturnOver returns the fractional price change over the last `lookback`
// periods (e.g. 0.12 = +12%).
func ReturnOver(closes []float64, lookback int) (float64, bool) {
	if lookback <= 0 || len(closes) <= lookback {
		return 0, false
	}
	past := closes[len(closes)-1-lookback]
	if past == 0 {
		return 0, false
	}
	return closes[len(closes)-1]/past - 1, true
}

// High returns the maximum close over the last n periods.
func High(closes []float64, n int) (float64, bool) {
	if n <= 0 || len(closes) < n {
		return 0, false
	}
	max := closes[len(closes)-n]
	for _, c := range closes[len(closes)-n:] {
		if c > max {
			max = c
		}
	}
	return max, true
}

// Latest returns the most recent close.
func Latest(closes []float64) (float64, bool) {
	if len(closes) == 0 {
		return 0, false
	}
	return closes[len(closes)-1], true
}
