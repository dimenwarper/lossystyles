package renderers

import (
	"fmt"
	"math"
	"strings"
)

// EvaBackground generates a tiling triangular wireframe pattern.
// Returns a slice of strings, one per row, for the full width x height area.
func EvaBackground(width, height int) []string {
	if width <= 0 || height <= 0 {
		return nil
	}

	// Repeating 4-row tile:
	//   ╱╲╱╲╱╲
	//   ╲╱╲╱╲╱
	//   ╱╲╱╲╱╲
	//   ╲╱╲╱╲╱
	// With occasional node markers ·

	tile := [4]string{
		"╱╲",
		"╲╱",
		"╱╲",
		"╲╱",
	}

	lines := make([]string, height)
	for row := 0; row < height; row++ {
		pattern := tile[row%4]
		var b strings.Builder
		col := 0
		for col < width {
			// Every 12 cols on alternating rows, place a node marker
			if col%12 == 0 && row%4 == 0 {
				b.WriteRune('·')
				col++
			} else {
				idx := col % len(pattern)
				b.WriteByte(pattern[idx])
				col++
			}
		}
		s := b.String()
		// Trim or pad to exact width
		runes := []rune(s)
		if len(runes) > width {
			runes = runes[:width]
		}
		for len(runes) < width {
			runes = append(runes, ' ')
		}
		lines[row] = string(runes)
	}
	return lines
}

// EvaScanline renders a horizontal scan line with NERV-style markers.
// Used as separator between sections.
func EvaScanline(width int) string {
	if width < 20 {
		return strings.Repeat("━", width)
	}

	var b strings.Builder
	b.WriteString("━━┫")
	mid := width - 6
	for i := 0; i < mid; i++ {
		if i%8 == 4 {
			b.WriteRune('┃')
		} else {
			b.WriteRune('━')
		}
	}
	b.WriteString("┣━━")
	return b.String()
}

// EvaLargeNumber renders a number in a large, prominent format
// using Unicode block characters for emphasis.
func EvaLargeNumber(val float64) string {
	formatted := FormatFloat(val)
	return formatted
}

// EvaProgressBar renders a NERV-style progress bar.
func EvaProgressBar(current, max float64, width int) string {
	if width < 4 {
		return ""
	}

	ratio := current / max
	if ratio > 1 {
		ratio = 1
	}
	if ratio < 0 {
		ratio = 0
	}

	filled := int(math.Round(ratio * float64(width-2)))
	empty := width - 2 - filled

	var b strings.Builder
	b.WriteRune('▐')
	b.WriteString(strings.Repeat("█", filled))
	b.WriteString(strings.Repeat("░", empty))
	b.WriteRune('▌')
	return b.String()
}

// EvaMAGIStatus returns the MAGI system status footer line.
func EvaMAGIStatus(width int, step int) string {
	statuses := []string{"MELCHIOR", "BALTHASAR", "CASPER"}

	// Cycle which MAGI node is "active" based on step
	activeIdx := (step / 10) % 3

	var parts []string
	for i, name := range statuses {
		marker := "◇"
		if i == activeIdx {
			marker = "◆"
		}
		parts = append(parts, fmt.Sprintf("%s %s-%d", marker, name, i+1))
	}

	inner := "MAGI ━━ " + strings.Join(parts, " ━ ") + " ━━ ACTIVE"

	// Center it
	pad := width - len([]rune(inner))
	if pad < 0 {
		pad = 0
	}
	left := pad / 2
	right := pad - left

	return strings.Repeat(" ", left) + inner + strings.Repeat(" ", right)
}

// EvaNERVLogo returns a compact NERV text mark.
func EvaNERVLogo() string {
	return "▸ NERV"
}
