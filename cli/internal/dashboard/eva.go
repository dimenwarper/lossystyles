package dashboard

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/lossystyles/cli/internal/renderers"
)

// renderEvaView renders the complete NERV HUD with rotating polytope background,
// triangular panel arrangement, and connecting lines.
func (m Model) renderEvaView() string {
	canvas := renderers.NewCanvas(m.width, m.height)

	// === 1. Draw background layers ===
	// Sine waves first (behind everything)
	drawEvaSineWaves(canvas, m.steps, m.width, m.height)
	// Polytope on top of waves
	renderers.DrawPolytope(canvas, m.steps, "#554422", "#887744")
	// "02" block text in top-right corner
	drawEva02(canvas, m.width, m.height)

	// === 2. Get metrics ===
	keys := m.sortedMetricKeys()

	// === 3. Panel dimensions — squarer panels, bottom anchored low ===
	headerH := 3
	footerH := 3

	panelH := m.height / 4
	if panelH < 8 {
		panelH = 8
	}
	if panelH > 14 {
		panelH = 14
	}

	// Top-center panel (main chart, same as before)
	topPanelW := panelH * 4
	if topPanelW > m.width*2/3 {
		topPanelW = m.width * 2 / 3
	}
	if topPanelW < 30 {
		topPanelW = 30
	}
	topX := (m.width - topPanelW) / 2
	topY := headerH + 2

	// Upper-left panel: new, narrower, same height as others
	ulPanelW := topX / 2
	if ulPanelW < 20 {
		ulPanelW = 20
	}
	ulX := 2
	ulY := headerH + 2

	// Bottom panels (same as before)
	botPanelW := topPanelW * 3 / 4
	if botPanelW < 24 {
		botPanelW = 24
	}
	bottomY := m.height - footerH - panelH - 1
	if bottomY < topY+panelH+3 {
		bottomY = topY + panelH + 3
	}
	gap := m.width - botPanelW*2
	blX := gap / 3
	brX := m.width - gap/3 - botPanelW

	// Ensure nothing overflows
	if bottomY+panelH > m.height-footerH {
		panelH = m.height - footerH - bottomY
		if panelH < 5 {
			panelH = 5
		}
	}

	// === 4. Header ===
	m.drawEvaHeaderCanvas(canvas)

	// === 5. Draw metric panels ===
	// Upper-left: new 4th metric (bigger panel)
	if len(keys) >= 4 {
		m.drawEvaPanelCanvas(canvas, keys[3], ulX, ulY, ulPanelW, panelH, true)
	}
	// Top-center: first metric (main chart, same position)
	if len(keys) >= 1 {
		m.drawEvaPanelCanvas(canvas, keys[0], topX, topY, topPanelW, panelH, true)
	}
	// Bottom-left: second metric
	if len(keys) >= 2 {
		m.drawEvaPanelCanvas(canvas, keys[1], blX, bottomY, botPanelW, panelH, false)
	}
	// Bottom-right: third metric
	if len(keys) >= 3 {
		m.drawEvaPanelCanvas(canvas, keys[2], brX, bottomY, botPanelW, panelH, false)
	}

	// === 7. MAGI footer ===
	m.drawEvaFooterCanvas(canvas)

	return canvas.Render()
}

// drawEvaHeaderCanvas draws the NERV header bar directly on the canvas.
func (m Model) drawEvaHeaderCanvas(canvas *renderers.Canvas) {
	elapsed := time.Since(m.startTime).Round(time.Second)

	borderColor := "#AA5500"
	canvas.DrawBox(0, 0, m.width, 3, borderColor)

	// NERV mark
	canvas.WriteText(2, 1, "▸ NERV", "#FF8800", true)
	canvas.WriteText(9, 1, "━━", "#AA5500", false)
	canvas.WriteText(12, 1, "lossystyles", "#CC6600", false)

	// Project name
	if m.project != "" {
		tag := "[" + m.project + "]"
		canvas.WriteText(24, 1, tag, "#FFAA00", true)
	}

	// Right side: step + time + pattern
	right := fmt.Sprintf("step %d ▸ %s", m.steps, elapsed)
	pattern := "PATTERN:BLUE"
	rightFull := right + "  " + pattern

	rx := m.width - len([]rune(rightFull)) - 2
	if rx < 30 {
		rx = 30
	}
	canvas.WriteText(rx, 1, right, "#553300", false)
	canvas.WriteText(rx+len([]rune(right))+2, 1, pattern, "#FF8800", true)
}

// drawEvaPanelCanvas draws a metric panel directly on the canvas.
// All content is clipped to the panel interior (x+2 to x+w-3).
func (m Model) drawEvaPanelCanvas(canvas *renderers.Canvas, key string, x, y, w, h int, isMain bool) {
	borderColor := "#AA5500"
	labelColor := "#FFAA00"
	valueColor := "#00FF66"
	chartColor := "#CC6600"
	mutedColor := "#553300"

	// Clear panel area (overwrite polytope)
	canvas.FillRect(x, y, w, h, ' ', "")

	// Border
	canvas.DrawBox(x, y, w, h, borderColor)

	// Interior bounds for clipping
	innerX := x + 2        // left content margin
	innerW := w - 4         // content width (2 margin + 1 border each side)
	if innerW < 4 {
		innerW = 4
	}
	maxX := x + w - 2       // rightmost content column (exclusive)

	// Label
	label := strings.ToUpper(key)
	writeClipped(canvas, innerX, y+1, label, innerW, labelColor, true)

	values := m.metrics[key]
	if len(values) == 0 {
		writeClipped(canvas, innerX, y+3, "AWAITING SIGNAL...", innerW, mutedColor, false)
		return
	}

	// Current value
	current := renderers.FormatFloat(values[len(values)-1])
	writeClipped(canvas, innerX, y+3, current, innerW, valueColor, true)

	if isMain && h > 8 {
		// Full chart for the main/top panel
		chartH := h - 7
		if chartH < 3 {
			chartH = 3
		}
		chart := renderers.BrailleChart(values, innerW, chartH)
		for i, line := range chart {
			row := y + 5 + i
			if row < y+h-2 {
				writeEvaChartLineClipped(canvas, innerX, row, line, innerW, maxX, chartColor, mutedColor)
			}
		}
	} else {
		// Progress bar for side panels
		barW := innerW
		if barW > 4 && len(values) > 1 {
			summary := renderers.Summarize(key, values, 0)
			ratio := 0.5
			span := summary.Max - summary.Min
			if span > 0 {
				ratio = (summary.Current - summary.Min) / span
			}
			bar := renderers.EvaProgressBar(ratio, 1.0, barW)
			writeEvaChartLineClipped(canvas, innerX, y+5, bar, innerW, maxX, chartColor, mutedColor)
		}

		// Sparkline
		if innerW > 0 && y+7 < y+h-2 {
			spark := renderers.Sparkline(values, innerW)
			writeEvaChartLineClipped(canvas, innerX, y+7, spark, innerW, maxX, chartColor, mutedColor)
		}
	}

	// Min/Max stats at bottom of panel
	if len(values) > 0 && y+h-2 > y+5 {
		summary := renderers.Summarize(key, values, 0)
		stats := fmt.Sprintf("min %s  max %s  %s",
			renderers.FormatFloat(summary.Min),
			renderers.FormatFloat(summary.Max),
			summary.Trend)
		writeClipped(canvas, innerX, y+h-2, stats, innerW, mutedColor, false)
	}
}

// writeClipped writes text clipped to maxW characters.
func writeClipped(canvas *renderers.Canvas, x, y int, text string, maxW int, color string, bold bool) {
	runes := []rune(text)
	if len(runes) > maxW {
		runes = runes[:maxW]
	}
	canvas.WriteText(x, y, string(runes), color, bold)
}

// writeEvaChartLineClipped writes a chart/sparkline with per-character coloring, clipped to panel bounds.
func writeEvaChartLineClipped(canvas *renderers.Canvas, x, y int, line string, maxW, maxX int, brightColor, dimColor string) {
	runes := []rune(line)
	if len(runes) > maxW {
		runes = runes[:maxW]
	}
	for i, ch := range runes {
		cx := x + i
		if cx >= maxX {
			break
		}
		var color string
		switch ch {
		case '█', '▇', '▆', '•':
			color = brightColor
		case '▅', '▄', '▃':
			color = "#AA5500"
		case '▂', '▁', '▐', '▌':
			color = "#884400"
		case '░':
			color = "#664400"
		case '▒':
			color = "#884400"
		case '▓':
			color = "#AA5500"
		case '·':
			color = dimColor
		default:
			color = dimColor
		}
		canvas.Set(cx, y, ch, color)
	}
}

// drawEvaFooterCanvas draws the MAGI system footer.
func (m Model) drawEvaFooterCanvas(canvas *renderers.Canvas) {
	y := m.height - 3
	borderColor := "#AA5500"

	canvas.DrawBox(0, y, m.width, 3, borderColor)

	// MAGI nodes — cycle active based on step
	activeIdx := (m.steps / 10) % 3
	nodes := []string{"MELCHIOR-1", "BALTHASAR-2", "CASPER-3"}

	// Build MAGI line
	canvas.WriteText(2, y+1, "MAGI", "#AA5500", true)
	canvas.WriteText(7, y+1, "━━", "#664400", false)

	col := 10
	for i, name := range nodes {
		if col >= m.width-2 {
			break
		}
		if i == activeIdx {
			canvas.SetBold(col, y+1, '◆', "#FF8800")
		} else {
			canvas.Set(col, y+1, '◇', "#664400")
		}
		col += 2

		color := "#553300"
		bold := false
		if i == activeIdx {
			color = "#FF8800"
			bold = true
		}
		writeClipped(canvas, col, y+1, name, m.width-col-2, color, bold)
		col += len(name)

		if i < 2 {
			canvas.WriteText(col, y+1, " ━ ", "#664400", false)
			col += 3
		}
	}

	// ACTIVE status on the right
	if col+7 < m.width-2 {
		canvas.WriteText(col+2, y+1, "━━", "#664400", false)
		canvas.WriteText(col+5, y+1, "ACTIVE", "#FF8800", true)
	}
}

// drawBlockDigit draws a block-character digit pattern onto the canvas,
// only placing █ characters and leaving everything else untouched.
func drawBlockDigit(canvas *renderers.Canvas, pattern []string, x, y int, color string) {
	for row, line := range pattern {
		col := 0
		for _, ch := range line {
			if ch == '█' {
				canvas.Set(x+col, y+row, '█', color)
			}
			col++
		}
	}
}

// sortedMetricKeys returns metric keys in stable sorted order.
func (m Model) sortedMetricKeys() []string {
	keys := make([]string, 0, len(m.metrics))
	for k := range m.metrics {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// drawEvaSineWaves draws layered sine waves across the background,
// inspired by the Eva-02 oscilloscope/waveform aesthetic.
func drawEvaSineWaves(canvas *renderers.Canvas, step, width, height int) {
	// Multiple sine waves with different frequencies, amplitudes, and phases
	// They scroll horizontally based on step count
	type wave struct {
		freq      float64 // frequency multiplier
		amp       float64 // amplitude as fraction of height
		phase     float64 // phase offset
		speed     float64 // scroll speed
		color     string
		ch        rune
	}

	waves := []wave{
		{freq: 0.04, amp: 0.15, phase: 0.0, speed: 0.08, color: "#FF4400", ch: '~'},
		{freq: 0.06, amp: 0.12, phase: 1.0, speed: 0.06, color: "#DD2200", ch: '~'},
		{freq: 0.03, amp: 0.18, phase: 2.0, speed: 0.10, color: "#BB1800", ch: '~'},
		{freq: 0.08, amp: 0.08, phase: 3.5, speed: 0.12, color: "#FF4400", ch: '~'},
		{freq: 0.05, amp: 0.10, phase: 5.0, speed: 0.04, color: "#DD2200", ch: '~'},
		{freq: 0.07, amp: 0.14, phase: 4.0, speed: 0.07, color: "#CC2200", ch: '~'},
		{freq: 0.09, amp: 0.06, phase: 2.5, speed: 0.09, color: "#BB1800", ch: '~'},
	}

	centerY := float64(height) * 0.55

	for _, w := range waves {
		t := float64(step) * w.speed
		for x := 0; x < width; x++ {
			y := centerY + w.amp*float64(height)*math.Sin(w.freq*float64(x)+w.phase+t)
			iy := int(math.Round(y))
			// Double-stroke: draw main line and a second stroke offset by 1
			if iy >= 0 && iy < height {
				canvas.Set(x, iy, w.ch, w.color)
			}
			if iy+1 >= 0 && iy+1 < height {
				canvas.Set(x, iy+1, '∼', w.color)
			}
		}
	}
}

// drawEva02 draws a large block-character "02" in the top-right area
// with hazard stripes underneath, matching the Eva-02 aesthetic.
func drawEva02(canvas *renderers.Canvas, width, height int) {
	// Large block digits — 12 wide x 7 tall each, outlined style
	zero := []string{
		"████████████",
		"██        ██",
		"██  ████  ██",
		"██  ████  ██",
		"██  ████  ██",
		"██        ██",
		"████████████",
	}
	two := []string{
		"████████████",
		"          ██",
		"          ██",
		"████████████",
		"██          ",
		"██          ",
		"████████████",
	}

	digitColor := "#AA2200"
	stripeColor := "#882200"

	// Total width: 12 + 2 gap + 12 = 26
	totalW := 26
	startX := width - totalW - 3
	startY := 4
	if startX < width/3 {
		return
	}
	if startY+10 > height {
		return
	}

	// Draw "0" — only place █ chars, skip spaces
	drawBlockDigit(canvas, zero, startX, startY, digitColor)
	// Draw "2" — offset by 14 (12 digit + 2 gap)
	drawBlockDigit(canvas, two, startX+14, startY, digitColor)

	// Hazard stripes underneath: alternating block stripes
	stripeY := startY + 8
	for col := 0; col < totalW; col++ {
		if col%2 == 0 {
			canvas.Set(startX+col, stripeY, '█', stripeColor)
			canvas.Set(startX+col, stripeY+1, '▐', stripeColor)
		} else {
			canvas.Set(startX+col, stripeY, '▌', stripeColor)
			canvas.Set(startX+col, stripeY+1, '█', stripeColor)
		}
	}
}
