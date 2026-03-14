package renderers

import (
	"math"
	"strings"
)

// BrailleChart renders a line chart using braille characters in the given dimensions.
// Returns a slice of strings, one per row.
func BrailleChart(values []float64, width, height int) []string {
	if len(values) == 0 || width <= 0 || height <= 0 {
		return []string{}
	}

	// Braille characters: each cell is 2 wide x 4 tall dots
	// We use a simplified approach: map values to rows
	chartWidth := width
	chartHeight := height

	// Resample values to fit width
	resampled := Resample(values, chartWidth)

	min, max := resampled[0], resampled[0]
	for _, v := range resampled {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}

	// Build character grid
	grid := make([][]rune, chartHeight)
	for i := range grid {
		grid[i] = make([]rune, chartWidth)
		for j := range grid[i] {
			grid[i][j] = ' '
		}
	}

	// Plot values
	span := max - min
	for col, v := range resampled {
		row := chartHeight - 1
		if span > 0 {
			normalized := (v - min) / span
			row = chartHeight - 1 - int(math.Round(normalized*float64(chartHeight-1)))
			if row < 0 {
				row = 0
			}
			if row >= chartHeight {
				row = chartHeight - 1
			}
		}
		grid[row][col] = '•'

		// Draw vertical line from bottom to point for area effect
		for r := row + 1; r < chartHeight; r++ {
			if grid[r][col] == ' ' {
				grid[r][col] = '·'
			}
		}
	}

	lines := make([]string, chartHeight)
	for i, row := range grid {
		lines[i] = string(row)
	}
	return lines
}

// resample resamples a slice of values to the target length using linear interpolation.
func Resample(values []float64, targetLen int) []float64 {
	n := len(values)
	if n == 0 {
		return nil
	}
	if n == targetLen {
		return values
	}
	if n == 1 {
		result := make([]float64, targetLen)
		for i := range result {
			result[i] = values[0]
		}
		return result
	}

	// Take last targetLen values if we have more
	if n > targetLen {
		return values[n-targetLen:]
	}

	// Interpolate if we have fewer
	result := make([]float64, targetLen)
	for i := range result {
		t := float64(i) * float64(n-1) / float64(targetLen-1)
		lo := int(math.Floor(t))
		hi := int(math.Ceil(t))
		if hi >= n {
			hi = n - 1
		}
		frac := t - float64(lo)
		result[i] = values[lo]*(1-frac) + values[hi]*frac
	}
	return result
}

// MiniChart renders a compact single-line chart using quarter-block characters.
func MiniChart(values []float64, width int) string {
	if len(values) == 0 {
		return strings.Repeat("─", width)
	}

	resampled := Resample(values, width)
	min, max := resampled[0], resampled[0]
	for _, v := range resampled {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}

	blocks := []rune{'▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}
	var b strings.Builder
	span := max - min
	for _, v := range resampled {
		idx := len(blocks) / 2
		if span > 0 {
			normalized := (v - min) / span
			idx = int(math.Round(normalized * float64(len(blocks)-1)))
			if idx >= len(blocks) {
				idx = len(blocks) - 1
			}
		}
		b.WriteRune(blocks[idx])
	}
	return b.String()
}
