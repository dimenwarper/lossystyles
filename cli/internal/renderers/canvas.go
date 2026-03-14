package renderers

import (
	"math"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Cell represents a single character on the canvas with styling.
type Cell struct {
	Ch    rune
	Color string // hex foreground color
	Bold  bool
}

// Canvas is a character-based framebuffer for compositing styled terminal output.
type Canvas struct {
	cells  [][]Cell
	Width  int
	Height int
}

// NewCanvas creates a canvas filled with spaces.
func NewCanvas(width, height int) *Canvas {
	cells := make([][]Cell, height)
	for y := range cells {
		cells[y] = make([]Cell, width)
		for x := range cells[y] {
			cells[y][x] = Cell{Ch: ' '}
		}
	}
	return &Canvas{cells: cells, Width: width, Height: height}
}

// InBounds checks if coordinates are within the canvas.
func (c *Canvas) InBounds(x, y int) bool {
	return x >= 0 && x < c.Width && y >= 0 && y < c.Height
}

// Set places a character at the given position.
func (c *Canvas) Set(x, y int, ch rune, color string) {
	if c.InBounds(x, y) {
		c.cells[y][x] = Cell{Ch: ch, Color: color}
	}
}

// SetBold places a bold character at the given position.
func (c *Canvas) SetBold(x, y int, ch rune, color string) {
	if c.InBounds(x, y) {
		c.cells[y][x] = Cell{Ch: ch, Color: color, Bold: true}
	}
}

// FillRect fills a rectangular area with a character.
func (c *Canvas) FillRect(x, y, w, h int, ch rune, color string) {
	for dy := 0; dy < h; dy++ {
		for dx := 0; dx < w; dx++ {
			c.Set(x+dx, y+dy, ch, color)
		}
	}
}

// DrawLine draws a line using a fixed character (DDA algorithm).
func (c *Canvas) DrawLine(x1, y1, x2, y2 int, ch rune, color string) {
	dx := x2 - x1
	dy := y2 - y1
	steps := intAbs(dx)
	if intAbs(dy) > steps {
		steps = intAbs(dy)
	}
	if steps == 0 {
		c.Set(x1, y1, ch, color)
		return
	}
	xInc := float64(dx) / float64(steps)
	yInc := float64(dy) / float64(steps)
	x, y := float64(x1), float64(y1)
	for i := 0; i <= steps; i++ {
		c.Set(int(math.Round(x)), int(math.Round(y)), ch, color)
		x += xInc
		y += yInc
	}
}

// DrawLineDirectional draws a line using direction-appropriate box-drawing characters.
func (c *Canvas) DrawLineDirectional(x1, y1, x2, y2 int, color string) {
	dx := x2 - x1
	dy := y2 - y1
	steps := intAbs(dx)
	if intAbs(dy) > steps {
		steps = intAbs(dy)
	}
	if steps == 0 {
		c.Set(x1, y1, '·', color)
		return
	}
	xInc := float64(dx) / float64(steps)
	yInc := float64(dy) / float64(steps)
	x, y := float64(x1), float64(y1)
	for i := 0; i <= steps; i++ {
		ix, iy := int(math.Round(x)), int(math.Round(y))
		ch := lineChar(xInc, yInc)
		c.Set(ix, iy, ch, color)
		x += xInc
		y += yInc
	}
}

// DrawBox draws a heavy-line box border.
func (c *Canvas) DrawBox(x, y, w, h int, color string) {
	if w < 2 || h < 2 {
		return
	}
	// Corners
	c.Set(x, y, '┏', color)
	c.Set(x+w-1, y, '┓', color)
	c.Set(x, y+h-1, '┗', color)
	c.Set(x+w-1, y+h-1, '┛', color)
	// Horizontal edges
	for i := 1; i < w-1; i++ {
		c.Set(x+i, y, '━', color)
		c.Set(x+i, y+h-1, '━', color)
	}
	// Vertical edges
	for i := 1; i < h-1; i++ {
		c.Set(x, y+i, '┃', color)
		c.Set(x+w-1, y+i, '┃', color)
	}
}

// WriteText writes a string horizontally starting at (x, y).
func (c *Canvas) WriteText(x, y int, text string, color string, bold bool) {
	for i, ch := range text {
		if bold {
			c.SetBold(x+i, y, ch, color)
		} else {
			c.Set(x+i, y, ch, color)
		}
	}
}

// Render converts the canvas to a styled terminal string.
// Groups adjacent cells with the same style for efficiency.
func (c *Canvas) Render() string {
	// Pre-allocate a style cache to avoid creating styles per-cell
	styleCache := make(map[Cell]lipgloss.Style)
	getStyle := func(color string, bold bool) lipgloss.Style {
		key := Cell{Color: color, Bold: bold}
		if s, ok := styleCache[key]; ok {
			return s
		}
		s := lipgloss.NewStyle()
		if color != "" {
			s = s.Foreground(lipgloss.Color(color))
		}
		if bold {
			s = s.Bold(true)
		}
		styleCache[key] = s
		return s
	}

	var lines []string
	for y := 0; y < c.Height; y++ {
		var b strings.Builder
		var run []rune
		curColor := ""
		curBold := false

		flush := func() {
			if len(run) == 0 {
				return
			}
			s := getStyle(curColor, curBold)
			b.WriteString(s.Render(string(run)))
			run = run[:0]
		}

		for x := 0; x < c.Width; x++ {
			cell := c.cells[y][x]
			if cell.Color != curColor || cell.Bold != curBold {
				flush()
				curColor = cell.Color
				curBold = cell.Bold
			}
			run = append(run, cell.Ch)
		}
		flush()
		lines = append(lines, b.String())
	}
	return strings.Join(lines, "\n")
}

// lineChar picks a box-drawing character based on line direction.
func lineChar(dx, dy float64) rune {
	absDx := math.Abs(dx)
	absDy := math.Abs(dy)

	if absDx < 0.001 && absDy < 0.001 {
		return '·'
	}
	if absDx < 0.001 {
		return '│'
	}

	ratio := absDy / absDx
	switch {
	case ratio < 0.4:
		return '─'
	case ratio > 2.5:
		return '│'
	default:
		if (dx > 0 && dy > 0) || (dx < 0 && dy < 0) {
			return '╲'
		}
		return '╱'
	}
}

func intAbs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
