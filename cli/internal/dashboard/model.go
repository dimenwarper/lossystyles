package dashboard

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lossystyles/cli/internal/protocol"
	"github.com/lossystyles/cli/internal/renderers"
	"github.com/lossystyles/cli/internal/themes"
)

// MetricMsg is sent when new metrics arrive from the SDK.
type MetricMsg protocol.Message

// TickMsg triggers a dashboard refresh.
type TickMsg time.Time

const refreshInterval = 250 * time.Millisecond

type Model struct {
	theme     themes.Theme
	project   string
	runID     string
	config    map[string]interface{}
	metrics   map[string][]float64 // key -> history
	steps     int
	startTime time.Time
	width     int
	height    int
	spinner   spinner.Model
	messages  <-chan protocol.Message
	quitting  bool
}

func New(msgs <-chan protocol.Message, themeName string) Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF79C6"))

	return Model{
		theme:     themes.Get(themeName),
		metrics:   make(map[string][]float64),
		startTime: time.Now(),
		spinner:   s,
		messages:  msgs,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		waitForMessage(m.messages),
		tickCmd(),
	)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case MetricMsg:
		pmsg := protocol.Message(msg)
		switch pmsg.Type {
		case protocol.MsgInit:
			m.project = pmsg.Project
			m.runID = pmsg.RunID
			m.config = pmsg.Config
			m.startTime = time.Now()
		case protocol.MsgLog:
			m.steps = pmsg.Step
			for k, v := range pmsg.Metrics {
				m.metrics[k] = append(m.metrics[k], v)
			}
		case protocol.MsgFinish:
			m.quitting = true
			return m, tea.Quit
		}
		return m, waitForMessage(m.messages)

	case TickMsg:
		return m, tickCmd()
	}

	return m, nil
}

func (m Model) View() string {
	if m.width == 0 {
		return "Initializing..."
	}

	// Themes with completely different layouts — dispatch early
	switch m.theme.RenderStyle {
	case themes.StyleEva:
		return m.renderEvaView()
	case themes.StylePCR:
		return m.renderPCRView()
	}

	var sections []string

	// Header
	sections = append(sections, m.renderHeader())

	// Chart area
	if len(m.metrics) > 0 {
		sections = append(sections, m.renderCharts())
		sections = append(sections, m.renderMetricsTable())
	} else {
		sections = append(sections, m.renderWaiting())
	}

	// Footer
	sections = append(sections, m.renderFooter())

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

// borderChar returns the theme's border character, defaulting to "─".
func (m Model) borderChar() string {
	if m.theme.BorderChar != "" {
		return m.theme.BorderChar
	}
	return "─"
}

// styleBorder renders a full-width border line using the theme's style.
func (m Model) styleBorder(offset int) string {
	raw := strings.Repeat(m.borderChar(), m.width)
	switch m.theme.RenderStyle {
	case themes.StyleRainbow:
		return themes.Rainbowify(raw, m.steps+offset)
	case themes.StylePCR:
		return themes.StylizePCR(raw)
	default:
		return m.theme.Border.Render(raw)
	}
}

// styleTitle renders text using the theme's title approach.
func (m Model) styleTitle(text string) string {
	switch m.theme.RenderStyle {
	case themes.StyleRainbow:
		return themes.Rainbowify(text, m.steps)
	case themes.StylePCR:
		return m.theme.Title.Render("🧬 " + text)
	default:
		return m.theme.Title.Render(text)
	}
}

func (m Model) renderHeader() string {
	elapsed := time.Since(m.startTime).Round(time.Second)

	title := m.styleTitle("lossystyles")

	project := ""
	if m.project != "" {
		switch m.theme.RenderStyle {
		case themes.StyleRainbow:
			project = themes.Rainbowify(m.project, m.steps+15)
		case themes.StylePCR:
			project = m.theme.Accent.Render("[" + m.project + "]")
		default:
			project = m.theme.Accent.Render(m.project)
		}
	}

	right := m.theme.Muted.Render(fmt.Sprintf("step %d • %s", m.steps, elapsed))

	gap := m.width - lipgloss.Width(title) - lipgloss.Width(project) - lipgloss.Width(right) - 4
	if gap < 1 {
		gap = 1
	}

	header := title + "  " + project + strings.Repeat(" ", gap) + right
	return header + "\n" + m.styleBorder(0)
}

func (m Model) renderCharts() string {
	// Sort metric keys for stable ordering
	keys := make([]string, 0, len(m.metrics))
	for k := range m.metrics {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	chartHeight := 8
	chartWidth := m.width - 4
	if chartWidth < 20 {
		chartWidth = 20
	}

	var charts []string
	for ki, key := range keys {
		values := m.metrics[key]
		if len(values) == 0 {
			continue
		}

		// Select chart renderer based on style
		var chart []string
		switch m.theme.RenderStyle {
		case themes.StylePCR:
			chart = renderers.ElectrophoresisChart(values, chartWidth, chartHeight)
		default:
			chart = renderers.BrailleChart(values, chartWidth, chartHeight)
		}

		// Style chart lines based on theme
		styled := make([]string, len(chart))
		for i, line := range chart {
			switch m.theme.RenderStyle {
			case themes.StyleRainbow:
				styled[i] = themes.Rainbowify(line, m.steps+i*3)
			case themes.StylePCR:
				styled[i] = themes.StylizePCR(line)
			default:
				styled[i] = m.theme.Sparkline.Render(line)
			}
		}

		// Label
		var label string
		switch m.theme.RenderStyle {
		case themes.StylePCR:
			label = m.theme.MetricKey.Render(fmt.Sprintf(" Lane %d: %s ", ki+1, key))
		default:
			label = m.theme.MetricKey.Render(fmt.Sprintf(" %s ", key))
		}

		box := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(m.theme.Border.GetForeground()).
			Padding(0, 1).
			Width(m.width - 2).
			Render(label + "\n" + strings.Join(styled, "\n"))

		charts = append(charts, box)
	}

	return strings.Join(charts, "\n")
}

func (m Model) renderMetricsTable() string {
	keys := make([]string, 0, len(m.metrics))
	for k := range m.metrics {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	sparkWidth := 20

	// Table header
	var histLabel string
	switch m.theme.RenderStyle {
	case themes.StylePCR:
		histLabel = "helix"
	default:
		histLabel = "history"
	}
	header := m.theme.Muted.Render(
		fmt.Sprintf("  %-16s %12s %12s %12s %3s  %-*s",
			"metric", "current", "min", "max", "Δ", sparkWidth, histLabel))
	border := m.theme.Border.Render("  " + strings.Repeat(m.borderChar(), m.width-4))

	rows := []string{header, border}
	for ki, key := range keys {
		summary := renderers.Summarize(key, m.metrics[key], sparkWidth)

		// Select sparkline style
		var spark string
		switch m.theme.RenderStyle {
		case themes.StyleRainbow:
			spark = themes.Rainbowify(summary.Spark, m.steps+ki*7)
		case themes.StylePCR:
			helixSpark := renderers.HelixSparkline(m.metrics[key], sparkWidth)
			spark = themes.StylizePCR(helixSpark)
		default:
			spark = m.theme.Sparkline.Render(summary.Spark)
		}

		row := fmt.Sprintf("  %s %s %s %s %s  %s",
			m.theme.MetricKey.Render(fmt.Sprintf("%-16s", summary.Key)),
			m.theme.MetricVal.Render(fmt.Sprintf("%12s", renderers.FormatFloat(summary.Current))),
			m.theme.Muted.Render(fmt.Sprintf("%12s", renderers.FormatFloat(summary.Min))),
			m.theme.Muted.Render(fmt.Sprintf("%12s", renderers.FormatFloat(summary.Max))),
			m.theme.Accent.Render(fmt.Sprintf("%3s", summary.Trend)),
			spark,
		)
		rows = append(rows, row)
	}

	return strings.Join(rows, "\n")
}

func (m Model) renderWaiting() string {
	return fmt.Sprintf("\n  %s Waiting for metrics...\n", m.spinner.View())
}

func (m Model) renderFooter() string {
	help := m.theme.Muted.Render("  q quit")
	return "\n" + m.styleBorder(50) + "\n" + help
}

// waitForMessage returns a command that reads from the message channel.
func waitForMessage(msgs <-chan protocol.Message) tea.Cmd {
	return func() tea.Msg {
		msg, ok := <-msgs
		if !ok {
			return MetricMsg(protocol.Message{Type: protocol.MsgFinish})
		}
		return MetricMsg(msg)
	}
}

func tickCmd() tea.Cmd {
	return tea.Tick(refreshInterval, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}
