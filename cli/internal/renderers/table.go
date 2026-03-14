package renderers

import (
	"fmt"
	"math"
)

type MetricSummary struct {
	Key     string
	Current float64
	Min     float64
	Max     float64
	Mean    float64
	Trend   string // "↑", "↓", "→"
	Spark   string
}

// Summarize computes summary stats for a metric's history.
func Summarize(key string, values []float64, sparkWidth int) MetricSummary {
	if len(values) == 0 {
		return MetricSummary{Key: key, Trend: "→"}
	}

	current := values[len(values)-1]
	min, max, sum := values[0], values[0], 0.0

	for _, v := range values {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
		sum += v
	}
	mean := sum / float64(len(values))

	trend := "→"
	if len(values) >= 2 {
		prev := values[len(values)-2]
		diff := current - prev
		if math.Abs(diff) > 1e-8 {
			if diff > 0 {
				trend = "↑"
			} else {
				trend = "↓"
			}
		}
	}

	return MetricSummary{
		Key:     key,
		Current: current,
		Min:     min,
		Max:     max,
		Mean:    mean,
		Trend:   trend,
		Spark:   Sparkline(values, sparkWidth),
	}
}

// FormatFloat formats a float for display, using scientific notation for very small/large values.
func FormatFloat(v float64) string {
	abs := math.Abs(v)
	switch {
	case abs == 0:
		return "0"
	case abs < 0.001 || abs >= 10000:
		return fmt.Sprintf("%.3e", v)
	case abs < 1:
		return fmt.Sprintf("%.4f", v)
	default:
		return fmt.Sprintf("%.4f", v)
	}
}
