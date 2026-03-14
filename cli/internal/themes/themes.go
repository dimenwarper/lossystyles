package themes

import (
	"fmt"
	"math"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// RenderStyle controls how renderers draw charts/sparklines — not just colors but the visual algorithm.
type RenderStyle string

const (
	StyleDefault RenderStyle = ""
	StyleRainbow RenderStyle = "rainbow"
	StylePCR     RenderStyle = "pcr"
	StyleEva     RenderStyle = "eva"
)

type Theme struct {
	Name        string
	RenderStyle RenderStyle
	Title       lipgloss.Style
	Border      lipgloss.Style
	MetricKey   lipgloss.Style
	MetricVal   lipgloss.Style
	Sparkline   lipgloss.Style
	Muted       lipgloss.Style
	Accent      lipgloss.Style
	Background  lipgloss.Style
	// BorderChar overrides the default "─" border character.
	BorderChar string
	// Rainbow indicates renderers should use per-character rainbow coloring.
	Rainbow bool
}

// RainbowColors returns a slice of hex color strings cycling through the spectrum.
func RainbowColors(n int) []string {
	colors := make([]string, n)
	for i := range n {
		hue := float64(i) / float64(n) * 360.0
		r, g, b := hslToRGB(hue, 1.0, 0.55)
		colors[i] = fmt.Sprintf("#%02X%02X%02X", r, g, b)
	}
	return colors
}

// Rainbowify applies per-character rainbow coloring to a string.
func Rainbowify(s string, offset int) string {
	runes := []rune(s)
	visible := 0
	for _, r := range runes {
		if r != ' ' {
			visible++
		}
	}
	if visible == 0 {
		return s
	}

	colors := RainbowColors(visible)
	var b strings.Builder
	ci := 0
	for _, r := range runes {
		if r == ' ' {
			b.WriteRune(' ')
		} else {
			idx := (ci + offset) % len(colors)
			style := lipgloss.NewStyle().Foreground(lipgloss.Color(colors[idx]))
			b.WriteString(style.Render(string(r)))
			ci++
		}
	}
	return b.String()
}

func hslToRGB(h, s, l float64) (uint8, uint8, uint8) {
	c := (1 - math.Abs(2*l-1)) * s
	x := c * (1 - math.Abs(math.Mod(h/60, 2)-1))
	m := l - c/2

	var r, g, b float64
	switch {
	case h < 60:
		r, g, b = c, x, 0
	case h < 120:
		r, g, b = x, c, 0
	case h < 180:
		r, g, b = 0, c, x
	case h < 240:
		r, g, b = 0, x, c
	case h < 300:
		r, g, b = x, 0, c
	default:
		r, g, b = c, 0, x
	}

	return uint8((r + m) * 255), uint8((g + m) * 255), uint8((b + m) * 255)
}

var Dark = Theme{
	Name:       "dark",
	Title:      lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FF79C6")),
	Border:     lipgloss.NewStyle().Foreground(lipgloss.Color("#6272A4")),
	MetricKey:  lipgloss.NewStyle().Foreground(lipgloss.Color("#8BE9FD")),
	MetricVal:  lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#F8F8F2")),
	Sparkline:  lipgloss.NewStyle().Foreground(lipgloss.Color("#50FA7B")),
	Muted:      lipgloss.NewStyle().Foreground(lipgloss.Color("#6272A4")),
	Accent:     lipgloss.NewStyle().Foreground(lipgloss.Color("#FFB86C")),
	Background: lipgloss.NewStyle(),
}

var Neon = Theme{
	Name:       "neon",
	Title:      lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FF00FF")),
	Border:     lipgloss.NewStyle().Foreground(lipgloss.Color("#00FFFF")),
	MetricKey:  lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF00")),
	MetricVal:  lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF")),
	Sparkline:  lipgloss.NewStyle().Foreground(lipgloss.Color("#FF00FF")),
	Muted:      lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")),
	Accent:     lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFF00")),
	Background: lipgloss.NewStyle(),
}

var Retro = Theme{
	Name:       "retro",
	Title:      lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00FF00")),
	Border:     lipgloss.NewStyle().Foreground(lipgloss.Color("#008800")),
	MetricKey:  lipgloss.NewStyle().Foreground(lipgloss.Color("#00CC00")),
	MetricVal:  lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00FF00")),
	Sparkline:  lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF00")),
	Muted:      lipgloss.NewStyle().Foreground(lipgloss.Color("#005500")),
	Accent:     lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF88")),
	Background: lipgloss.NewStyle(),
}

var Minimal = Theme{
	Name:       "minimal",
	Title:      lipgloss.NewStyle().Bold(true),
	Border:     lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")),
	MetricKey:  lipgloss.NewStyle().Foreground(lipgloss.Color("#AAAAAA")),
	MetricVal:  lipgloss.NewStyle().Bold(true),
	Sparkline:  lipgloss.NewStyle(),
	Muted:      lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")),
	Accent:     lipgloss.NewStyle().Bold(true),
	Background: lipgloss.NewStyle(),
}

var Rainbow = Theme{
	Name:        "rainbow",
	RenderStyle: StyleRainbow,
	Title:       lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FF0000")),
	Border:      lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")),
	MetricKey:   lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00DDFF")),
	MetricVal:   lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF")),
	Sparkline:   lipgloss.NewStyle(),
	Muted:       lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")),
	Accent:      lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFAA00")),
	Background:  lipgloss.NewStyle(),
	Rainbow:     true,
}

// PCR theme — UV-illuminated gel electrophoresis aesthetic.
// Phosphor-green palette with smeared band charts and DNA helix sparklines.
var PCR = Theme{
	Name:        "pcr",
	RenderStyle: StylePCR,
	Title:       lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00FF88")),
	Border:      lipgloss.NewStyle().Foreground(lipgloss.Color("#004422")),
	MetricKey:   lipgloss.NewStyle().Foreground(lipgloss.Color("#33FFAA")),
	MetricVal:   lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#CCFFEE")),
	Sparkline:   lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF66")),
	Muted:       lipgloss.NewStyle().Foreground(lipgloss.Color("#225533")),
	Accent:      lipgloss.NewStyle().Foreground(lipgloss.Color("#88FFCC")),
	Background:  lipgloss.NewStyle(),
	BorderChar:  "~",
}

// Eva theme — NERV HUD aesthetic from Neon Genesis Evangelion.
// Orange/amber on dark, heavy borders, geometric background, MAGI footer.
var Eva = Theme{
	Name:        "eva",
	RenderStyle: StyleEva,
	Title:       lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FF8800")),
	Border:      lipgloss.NewStyle().Foreground(lipgloss.Color("#AA5500")),
	MetricKey:   lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFAA00")),
	MetricVal:   lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FF8800")),
	Sparkline:   lipgloss.NewStyle().Foreground(lipgloss.Color("#CC6600")),
	Muted:       lipgloss.NewStyle().Foreground(lipgloss.Color("#553300")),
	Accent:      lipgloss.NewStyle().Foreground(lipgloss.Color("#FFCC44")),
	Background:  lipgloss.NewStyle().Foreground(lipgloss.Color("#1A0800")),
	BorderChar:  "━",
}

// EvaBorder is the heavy box-drawing border used for Eva panels.
var EvaBorder = lipgloss.Border{
	Top:         "━",
	Bottom:      "━",
	Left:        "┃",
	Right:       "┃",
	TopLeft:     "┏",
	TopRight:    "┓",
	BottomLeft:  "┗",
	BottomRight: "┛",
}

// EvaChartStyles are pre-allocated styles for per-character Eva chart coloring.
var EvaChartStyles = map[rune]lipgloss.Style{
	'█': lipgloss.NewStyle().Foreground(lipgloss.Color("#FF8800")),
	'▓': lipgloss.NewStyle().Foreground(lipgloss.Color("#CC6600")),
	'▒': lipgloss.NewStyle().Foreground(lipgloss.Color("#884400")),
	'░': lipgloss.NewStyle().Foreground(lipgloss.Color("#553300")),
	'•': lipgloss.NewStyle().Foreground(lipgloss.Color("#FF8800")),
	'·': lipgloss.NewStyle().Foreground(lipgloss.Color("#553300")),
	'╱': lipgloss.NewStyle().Foreground(lipgloss.Color("#331800")),
	'╲': lipgloss.NewStyle().Foreground(lipgloss.Color("#331800")),
}

// StylizeEva applies per-character Eva amber coloring to a string.
func StylizeEva(s string) string {
	var b strings.Builder
	for _, r := range s {
		if style, ok := EvaChartStyles[r]; ok {
			b.WriteString(style.Render(string(r)))
		} else if r == ' ' {
			b.WriteRune(' ')
		} else {
			// Default to muted amber for unknown chars
			b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#663300")).Render(string(r)))
		}
	}
	return b.String()
}

// StylizeEvaBg applies very dark amber coloring for background pattern chars.
func StylizeEvaBg(s string) string {
	style := lipgloss.NewStyle().Foreground(lipgloss.Color("#1A0800"))
	return style.Render(s)
}

var Themes = map[string]Theme{
	"dark":    Dark,
	"neon":    Neon,
	"retro":   Retro,
	"minimal": Minimal,
	"rainbow": Rainbow,
	"pcr":     PCR,
	"eva":     Eva,
}

// PCRChartStyles are pre-allocated styles for per-character bio chart coloring.
var PCRChartStyles = map[rune]lipgloss.Style{
	'█': lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF66")),
	'▓': lipgloss.NewStyle().Foreground(lipgloss.Color("#00DD55")),
	'▒': lipgloss.NewStyle().Foreground(lipgloss.Color("#00AA33")),
	'░': lipgloss.NewStyle().Foreground(lipgloss.Color("#006622")),
	'╳': lipgloss.NewStyle().Foreground(lipgloss.Color("#33FFAA")),
	'╲': lipgloss.NewStyle().Foreground(lipgloss.Color("#22CC88")),
	'╱': lipgloss.NewStyle().Foreground(lipgloss.Color("#22CC88")),
	'·': lipgloss.NewStyle().Foreground(lipgloss.Color("#004422")),
}

// StylizePCR applies per-character bio coloring to a string.
func StylizePCR(s string) string {
	var b strings.Builder
	for _, r := range s {
		if style, ok := PCRChartStyles[r]; ok {
			b.WriteString(style.Render(string(r)))
		} else if r == ' ' {
			b.WriteRune(' ')
		} else {
			b.WriteString(PCRChartStyles['·'].Render(string(r)))
		}
	}
	return b.String()
}

// StylizeEvaMAGI applies Eva coloring to the MAGI footer, highlighting the active node.
func StylizeEvaMAGI(s string, step int) string {
	activeIdx := (step / 10) % 3
	activeNames := []string{"MELCHIOR", "BALTHASAR", "CASPER"}

	activeName := activeNames[activeIdx]
	activeStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FF8800"))
	inactiveStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#663300"))
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#AA5500"))
	symbolStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFAA00"))

	// Rebuild with styling
	result := labelStyle.Render("MAGI") + " " + labelStyle.Render("━━") + " "
	for i, name := range activeNames {
		marker := "◇"
		style := inactiveStyle
		if name == activeName {
			marker = "◆"
			style = activeStyle
		}
		result += symbolStyle.Render(marker) + " " + style.Render(fmt.Sprintf("%s-%d", name, i+1))
		if i < 2 {
			result += " " + labelStyle.Render("━") + " "
		}
	}
	result += " " + labelStyle.Render("━━") + " " + activeStyle.Render("ACTIVE")
	return result
}

func Get(name string) Theme {
	if t, ok := Themes[name]; ok {
		return t
	}
	return Dark
}
