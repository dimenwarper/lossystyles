package renderers

import (
	"math"
	"strings"
)

// Density characters ordered by intensity (low to high).
var densityChars = []rune{' ', '░', '▒', '▓', '█'}

// ElectrophoresisChart renders values as vertical smeared bands,
// simulating DNA fragments migrating through an agarose gel.
// Each data point produces a gaussian-spread band centered at the value's Y position.
func ElectrophoresisChart(values []float64, width, height int) []string {
	if len(values) == 0 || width <= 0 || height <= 0 {
		return []string{}
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

	// Sigma controls band thickness (in rows). Wider = more smear.
	sigma := 1.3

	grid := make([][]rune, height)
	for i := range grid {
		grid[i] = make([]rune, width)
		for j := range grid[i] {
			grid[i][j] = ' '
		}
	}

	span := max - min
	for col, v := range resampled {
		// Map value to center row (inverted: low value = bottom)
		centerRow := float64(height-1) * 0.5
		if span > 0 {
			normalized := (v - min) / span
			centerRow = float64(height-1) - normalized*float64(height-1)
		}

		for row := 0; row < height; row++ {
			dist := math.Abs(float64(row) - centerRow)
			intensity := math.Exp(-(dist * dist) / (2 * sigma * sigma))

			var ch rune
			switch {
			case intensity >= 0.85:
				ch = '█'
			case intensity >= 0.55:
				ch = '▓'
			case intensity >= 0.30:
				ch = '▒'
			case intensity >= 0.12:
				ch = '░'
			default:
				ch = ' '
			}
			grid[row][col] = ch
		}
	}

	// Add gel lane markers at the edges
	lines := make([]string, height)
	for i, row := range grid {
		lines[i] = string(row)
	}
	return lines
}

// HelixSparkline renders a sparkline as a DNA double helix.
// Two strands weave sinusoidally — the "front" strand shows metric block chars,
// the "back" strand shows lighter density chars, and crossover points use ╳.
func HelixSparkline(values []float64, width int) string {
	if len(values) == 0 {
		return strings.Repeat("·", width)
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

	// Block chars for front strand (bold, full blocks)
	frontBlocks := []rune{'▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}
	// Density chars for back strand (lighter, behind)
	backBlocks := []rune{'·', '·', '░', '░', '▒', '▒', '▓', '▓'}

	period := 10.0 // columns per full helix turn
	crossoverThreshold := 0.25

	var b strings.Builder
	span := max - min
	for i, v := range resampled {
		phase := 2 * math.Pi * float64(i) / period
		sinVal := math.Sin(phase)

		// Map value to block index
		idx := len(frontBlocks) / 2
		if span > 0 {
			normalized := (v - min) / span
			idx = int(math.Round(normalized * float64(len(frontBlocks)-1)))
			if idx >= len(frontBlocks) {
				idx = len(frontBlocks) - 1
			}
			if idx < 0 {
				idx = 0
			}
		}

		switch {
		case math.Abs(sinVal) < crossoverThreshold:
			// Crossover point — strands cross
			b.WriteRune('╳')
		case sinVal > 0:
			// Front strand — bold blocks
			b.WriteRune(frontBlocks[idx])
		default:
			// Back strand — lighter chars
			b.WriteRune(backBlocks[idx])
		}
	}
	return b.String()
}

// HelixSparklineDouble renders a 2-row DNA double helix sparkline.
// Returns top and bottom strand lines.
func HelixSparklineDouble(values []float64, width int) [2]string {
	if len(values) == 0 {
		return [2]string{
			strings.Repeat(" ", width),
			strings.Repeat(" ", width),
		}
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
	period := 8.0

	var top, bot strings.Builder
	span := max - min
	for i, v := range resampled {
		phase := 2 * math.Pi * float64(i) / period
		sinVal := math.Sin(phase)

		idx := len(blocks) / 2
		if span > 0 {
			normalized := (v - min) / span
			idx = int(math.Round(normalized * float64(len(blocks)-1)))
			if idx >= len(blocks) {
				idx = len(blocks) - 1
			}
		}

		crossover := math.Abs(sinVal) < 0.25

		if crossover {
			top.WriteRune('╲')
			bot.WriteRune('╱')
		} else if sinVal > 0 {
			top.WriteRune(blocks[idx])
			bot.WriteRune('·')
		} else {
			top.WriteRune('·')
			bot.WriteRune(blocks[idx])
		}
	}

	return [2]string{top.String(), bot.String()}
}
