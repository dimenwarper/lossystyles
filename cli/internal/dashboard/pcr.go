package dashboard

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/lossystyles/cli/internal/renderers"
)

// renderPCRView renders a gel electrophoresis dashboard — vertical lanes,
// horizontal bands, marker ladders on the edges.
func (m Model) renderPCRView() string {
	canvas := renderers.NewCanvas(m.width, m.height)

	keys := m.pcrSortedKeys()

	// === Header ===
	m.drawPCRHeader(canvas)

	// === Gel panel ===
	gelX := 1
	gelY := 3
	gelW := m.width - 2
	gelH := m.height - 7 // room for header + footer
	if gelH < 10 {
		gelH = 10
	}

	m.drawGelPanel(canvas, keys, gelX, gelY, gelW, gelH)

	// === Footer with stats ===
	m.drawPCRFooter(canvas, keys)

	return canvas.Render()
}

// drawPCRHeader draws the bio theme header.
func (m Model) drawPCRHeader(canvas *renderers.Canvas) {
	elapsed := time.Since(m.startTime).Round(time.Second)
	borderColor := "#006633"

	canvas.DrawBox(0, 0, m.width, 3, borderColor)
	canvas.WriteText(2, 1, "🧬 lossystyles", "#00FF88", true)

	if m.project != "" {
		canvas.WriteText(18, 1, "["+m.project+"]", "#33FFAA", false)
	}

	right := fmt.Sprintf("step %d • %s", m.steps, elapsed)
	rx := m.width - len([]rune(right)) - 2
	if rx < 30 {
		rx = 30
	}
	canvas.WriteText(rx, 1, right, "#225533", false)
}

// drawGelPanel renders the full gel with marker lanes + metric lanes.
func (m Model) drawGelPanel(canvas *renderers.Canvas, keys []string, gelX, gelY, gelW, gelH int) {
	borderColor := "#004422"

	// Gel border
	canvas.DrawBox(gelX, gelY, gelW, gelH, borderColor)

	innerX := gelX + 1
	innerY := gelY + 2 // leave room for lane labels at top
	innerW := gelW - 2
	innerH := gelH - 3 // room for labels

	// Calculate lane layout: M | metric1 | metric2 | ... | M
	numMetrics := len(keys)
	numLanes := numMetrics + 2 // 2 marker lanes
	if numLanes == 0 {
		return
	}

	laneW := innerW / numLanes
	if laneW < 3 {
		laneW = 3
	}
	separatorW := 1 // gap between lanes
	actualLaneW := laneW - separatorW
	if actualLaneW < 2 {
		actualLaneW = 2
	}

	// Total width used
	totalUsed := numLanes * laneW
	startX := innerX + (innerW-totalUsed)/2 // center the lanes

	// Draw lane separators (thin vertical lines)
	for i := 0; i <= numLanes; i++ {
		sx := startX + i*laneW
		for row := 0; row < innerH; row++ {
			canvas.Set(sx, innerY+row, '│', "#003318")
		}
	}

	// Compute global min/max across all metrics for ladder reference
	globalMin, globalMax := math.MaxFloat64, -math.MaxFloat64
	for _, key := range keys {
		for _, v := range m.metrics[key] {
			if v < globalMin {
				globalMin = v
			}
			if v > globalMax {
				globalMax = v
			}
		}
	}
	if globalMin == math.MaxFloat64 {
		globalMin, globalMax = 0, 1
	}

	// Draw marker lane (left)
	drawMarkerLane(canvas, startX+1, innerY, actualLaneW, innerH, globalMin, globalMax)
	canvas.WriteText(startX+1, gelY+1, "M", "#006633", true)

	// Draw metric lanes — each with its own color
	for i, key := range keys {
		lx := startX + (i+1)*laneW + 1
		values := m.metrics[key]
		pal := laneColorForIndex(i)

		// Lane label at top, using the lane's brightest color
		label := strings.ToUpper(key)
		if len(label) > actualLaneW {
			label = label[:actualLaneW]
		}
		canvas.WriteText(lx, gelY+1, label, pal.Full, true)

		if len(values) > 0 {
			drawMetricLane(canvas, lx, innerY, actualLaneW, innerH, values, pal)
		}
	}

	// Draw marker lane (right) — with value labels
	mlx := startX + (numMetrics+1)*laneW + 1
	drawMarkerLane(canvas, mlx, innerY, actualLaneW, innerH, globalMin, globalMax)
	canvas.WriteText(mlx, gelY+1, "M", "#006633", true)
}

// drawMarkerLane draws a DNA ladder with labeled reference bands at nice round values.
// High values map to top (low row), low values map to bottom (high row).
func drawMarkerLane(canvas *renderers.Canvas, x, y, w, h int, minVal, maxVal float64) {
	span := maxVal - minVal
	if span <= 0 {
		span = 1
	}

	// Pick nice round tick values
	ticks := niceTickValues(minVal, maxVal, 6)

	for _, tick := range ticks {
		// Map value to row: high values = top, low = bottom
		normalized := (tick - minVal) / span
		row := int(float64(h-1) - normalized*float64(h-1))
		if row < 0 || row >= h {
			continue
		}

		// Draw the band blob — gaussian smear for realistic look
		sigma := 0.8
		for dr := -2; dr <= 2; dr++ {
			r := row + dr
			if r < 0 || r >= h {
				continue
			}
			dist := math.Abs(float64(dr))
			intensity := math.Exp(-(dist * dist) / (2 * sigma * sigma))

			ch, _ := bandCharColor(intensity)
			if ch == ' ' {
				continue
			}
			// Horizontal gaussian: brighter at center
			hCenter := float64(w-1) / 2.0
			hSigma := float64(w) / 2.5
			for col := 0; col < w; col++ {
				hDist := math.Abs(float64(col) - hCenter)
				hFalloff := math.Exp(-(hDist * hDist) / (2 * hSigma * hSigma))
				pch, pcolor := bandCharColor(intensity * hFalloff)
				if pch != ' ' {
					canvas.Set(x+col, y+r, pch, pcolor)
				}
			}
		}

		// Draw value label to the right of the band (if space allows)
		label := renderers.FormatFloat(tick)
		if len(label) <= w {
			canvas.WriteText(x, y+row, label, "#006633", false)
		}
	}
}

// niceTickValues returns ~n nice round tick values spanning [min, max].
func niceTickValues(min, max float64, n int) []float64 {
	span := max - min
	if span <= 0 {
		return []float64{min}
	}

	// Find a nice step size
	rawStep := span / float64(n)
	magnitude := math.Pow(10, math.Floor(math.Log10(rawStep)))
	normalized := rawStep / magnitude

	var niceStep float64
	switch {
	case normalized <= 1.5:
		niceStep = 1 * magnitude
	case normalized <= 3.5:
		niceStep = 2 * magnitude
	case normalized <= 7.5:
		niceStep = 5 * magnitude
	default:
		niceStep = 10 * magnitude
	}

	// Generate ticks from rounded start
	start := math.Ceil(min/niceStep) * niceStep
	var ticks []float64
	for v := start; v <= max; v += niceStep {
		ticks = append(ticks, v)
	}
	return ticks
}

// drawMetricLane draws bands in a gel lane that simulate electrophoresis separation.
// Early values appear as a diffuse smear; as training progresses, values separate
// into distinct sharp bands — mimicking how DNA fragments resolve over time.
func drawMetricLane(canvas *renderers.Canvas, x, y, w, h int, values []float64, pal laneColorPalette) {
	if len(values) == 0 || w <= 0 || h <= 0 {
		return
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

	// Build density map: for each row, accumulate gaussian contributions.
	// Split values into time segments with decreasing sigma (smear → sharp bands).
	density := make([]float64, h)
	span := max - min

	numSegments := 5
	segSize := len(values) / numSegments
	if segSize < 1 {
		segSize = 1
	}

	for vi, v := range values {
		// Which time segment is this value in? (0 = earliest, numSegments-1 = latest)
		seg := vi / segSize
		if seg >= numSegments {
			seg = numSegments - 1
		}

		// Sigma decreases from wide (diffuse smear) to narrow (sharp bands)
		// Early: sigma=3.0 (blurry), Late: sigma=0.6 (crisp)
		progress := float64(seg) / float64(numSegments-1)
		sigma := 3.0 - 2.4*progress // 3.0 → 0.6

		// More recent values contribute more (fade older ones)
		recency := float64(vi+1) / float64(len(values))
		weight := 0.2 + 0.8*recency

		// Map value to row (high values = top, low = bottom)
		centerRow := float64(h-1) * 0.5
		if span > 0 {
			normalized := (v - min) / span
			centerRow = float64(h-1) - normalized*float64(h-1)
		}

		for row := 0; row < h; row++ {
			dist := math.Abs(float64(row) - centerRow)
			density[row] += weight * math.Exp(-(dist*dist)/(2*sigma*sigma))
		}
	}

	// Find max density for normalization
	maxDensity := 0.0
	for _, d := range density {
		if d > maxDensity {
			maxDensity = d
		}
	}
	if maxDensity == 0 {
		return
	}

	// Render each row
	for row := 0; row < h; row++ {
		intensity := density[row] / maxDensity
		if intensity < 0.03 {
			continue
		}

		// Horizontal gaussian profile: brighter at center, dimmer at edges
		hCenter := float64(w-1) / 2.0
		hSigma := float64(w) / 3.0

		for col := 0; col < w; col++ {
			hDist := math.Abs(float64(col) - hCenter)
			hFalloff := math.Exp(-(hDist * hDist) / (2 * hSigma * hSigma))
			pixel := intensity * hFalloff

			ch, color := bandCharColorPalette(pixel, pal)
			if ch != ' ' {
				canvas.Set(x+col, y+row, ch, color)
			}
		}
	}
}

// laneColorPalette defines the color gradient for a single lane (brightest to dimmest).
type laneColorPalette struct {
	Full, High, Mid, Low, Faint string
}

// laneColors are the per-lane color palettes: white, green, light blue, orange.
// Wraps around if there are more than 4 metrics.
var laneColors = []laneColorPalette{
	{"#FFFFFF", "#DDDDDD", "#AAAAAA", "#777777", "#444444"}, // white
	{"#00FF66", "#00DD55", "#00AA33", "#007722", "#004411"}, // green
	{"#66DDFF", "#55BBDD", "#4499BB", "#337799", "#224455"}, // light blue
	{"#FFAA33", "#DD8822", "#BB6611", "#884400", "#553300"}, // orange
}

// laneColorForIndex returns the palette for lane index i.
func laneColorForIndex(i int) laneColorPalette {
	return laneColors[i%len(laneColors)]
}

// bandCharColorPalette returns the character and color for a given intensity using a palette.
func bandCharColorPalette(intensity float64, pal laneColorPalette) (rune, string) {
	switch {
	case intensity >= 0.8:
		return '█', pal.Full
	case intensity >= 0.6:
		return '▓', pal.High
	case intensity >= 0.4:
		return '▒', pal.Mid
	case intensity >= 0.2:
		return '░', pal.Low
	case intensity >= 0.08:
		return '·', pal.Faint
	default:
		return ' ', ""
	}
}

// bandCharColor returns the character and green color for a given intensity (used by marker lanes).
func bandCharColor(intensity float64) (rune, string) {
	return bandCharColorPalette(intensity, laneColors[1]) // green palette
}

// drawPCRFooter draws a stats line at the bottom.
func (m Model) drawPCRFooter(canvas *renderers.Canvas, keys []string) {
	y := m.height - 4
	borderColor := "#004422"
	canvas.DrawBox(0, y, m.width, 4, borderColor)

	col := 2
	for i, key := range keys {
		values := m.metrics[key]
		if len(values) == 0 {
			continue
		}
		pal := laneColorForIndex(i)
		summary := renderers.Summarize(key, values, 0)

		label := strings.ToUpper(key)
		canvas.WriteText(col, y+1, label, pal.Full, true)
		val := renderers.FormatFloat(summary.Current)
		canvas.WriteText(col, y+2, val+" "+summary.Trend, pal.High, false)
		col += len(label) + 2
		if col+10 > m.width-2 {
			break
		}

		minmax := fmt.Sprintf("[%s..%s]", renderers.FormatFloat(summary.Min), renderers.FormatFloat(summary.Max))
		canvas.WriteText(col, y+2, minmax, pal.Low, false)
		col += len(minmax) + 3
	}
}

// pcrSortedKeys returns metric keys sorted for the bio layout.
func (m Model) pcrSortedKeys() []string {
	keys := make([]string, 0, len(m.metrics))
	for k := range m.metrics {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
