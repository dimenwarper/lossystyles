package renderers

import (
	"math"
	"strings"
)

var sparkBlocks = []rune{'▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}

// Sparkline renders a slice of float64 values as a Unicode sparkline string.
func Sparkline(values []float64, width int) string {
	if len(values) == 0 || width <= 0 {
		return ""
	}

	// Take the last `width` values
	if len(values) > width {
		values = values[len(values)-width:]
	}

	min, max := values[0], values[0]
	for _, v := range values {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}

	var b strings.Builder
	span := max - min
	for _, v := range values {
		idx := 0
		if span > 0 {
			normalized := (v - min) / span
			idx = int(math.Round(normalized * float64(len(sparkBlocks)-1)))
			if idx >= len(sparkBlocks) {
				idx = len(sparkBlocks) - 1
			}
		} else {
			idx = len(sparkBlocks) / 2
		}
		b.WriteRune(sparkBlocks[idx])
	}
	return b.String()
}
