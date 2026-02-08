package main

import (
	"fmt"
	"math"
	"os"
	"strings"
	"time"

	"github.com/Caryyon/antenna/internal/api"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Gmork theme colors (matching web frontend CSS variables)
var (
	colorGreen  = lipgloss.Color("#00ff99")
	colorOrange = lipgloss.Color("#ff8c4c")
	colorPurple = lipgloss.Color("#bf6fff")
	colorRed    = lipgloss.Color("#ff4477")
	colorCyan   = lipgloss.Color("#00ffd5")
	colorDim    = lipgloss.Color("#555555")
	colorFg     = lipgloss.Color("#bbbbbb")
	colorWhite  = lipgloss.Color("#ffffff")
	colorBorder = lipgloss.Color("#1a1a1a")
	colorPanel  = lipgloss.Color("#0a0a0a")
)

type view int

const (
	viewDashboard view = iota
	viewDetail
)

type tickMsg time.Time

type model struct {
	client    *api.Client
	dashboard api.DashboardData
	hourly    []api.HourlyBucket
	cursor    int
	view      view
	width     int
	height    int
	interval  time.Duration
	section   int // 0=active, 1=idle, 2=sub, 3=cron
	err       error
}

// grouped returns sessions split by kind
func (m model) grouped() (active, idle, subs, crons []api.Session) {
	for _, s := range m.dashboard.Sessions {
		switch s.Kind {
		case "cron":
			crons = append(crons, s)
		case "subagent":
			subs = append(subs, s)
		default:
			if s.IsActive {
				active = append(active, s)
			} else {
				idle = append(idle, s)
			}
		}
	}
	return
}

// flatList returns the session at the current cursor position across all groups
func (m model) selectedSession() (api.Session, bool) {
	active, idle, subs, crons := m.grouped()
	all := make([]api.Session, 0, len(m.dashboard.Sessions))
	all = append(all, active...)
	all = append(all, idle...)
	all = append(all, subs...)
	all = append(all, crons...)
	if m.cursor >= 0 && m.cursor < len(all) {
		return all[m.cursor], true
	}
	return api.Session{}, false
}

func (m model) totalItems() int {
	return len(m.dashboard.Sessions)
}

func initialModel() model {
	dir := os.Getenv("OPENCLAW_DIR")
	if dir == "" {
		home, _ := os.UserHomeDir()
		dir = home + "/.openclaw"
	}
	interval := 5 * time.Second
	if v := os.Getenv("ANTENNA_INTERVAL"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			interval = d
		}
	}
	c := api.NewClient(dir)
	return model{
		client:    c,
		dashboard: c.GetDashboard(),
		hourly:    c.GetHourlyActivity(),
		interval:  interval,
	}
}

func tickCmd(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(t time.Time) tea.Msg { return tickMsg(t) })
}

func (m model) Init() tea.Cmd {
	return tickCmd(m.interval)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			if m.view == viewDetail {
				m.view = viewDashboard
				return m, nil
			}
			return m, tea.Quit
		case "j", "down":
			if m.view == viewDashboard && m.cursor < m.totalItems()-1 {
				m.cursor++
			}
		case "k", "up":
			if m.view == viewDashboard && m.cursor > 0 {
				m.cursor--
			}
		case "enter":
			if m.view == viewDashboard && m.totalItems() > 0 {
				m.view = viewDetail
			}
		case "esc", "backspace":
			m.view = viewDashboard
		case "tab":
			if m.view == viewDashboard {
				m.view = viewDetail
			} else {
				m.view = viewDashboard
			}
		case "r":
			m.dashboard = m.client.GetDashboard()
			m.hourly = m.client.GetHourlyActivity()
		}
		return m, nil

	case tickMsg:
		m.dashboard = m.client.GetDashboard()
		m.hourly = m.client.GetHourlyActivity()
		if m.cursor >= m.totalItems() && m.totalItems() > 0 {
			m.cursor = m.totalItems() - 1
		}
		return m, tickCmd(m.interval)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	}
	return m, nil
}

func (m model) View() string {
	w := m.width
	if w == 0 {
		w = 80
	}
	h := m.height
	if h == 0 {
		h = 24
	}

	var b strings.Builder

	// ── Stats Bar ──
	b.WriteString(m.renderStatsBar(w))
	b.WriteString("\n")

	switch m.view {
	case viewDashboard:
		b.WriteString(m.renderDashboard(w, h))
	case viewDetail:
		b.WriteString(m.renderDetail(w, h))
	}

	return b.String()
}

func (m model) renderStatsBar(w int) string {
	active, idle, subs, crons := m.grouped()

	sep := lipgloss.NewStyle().Foreground(colorBorder).Render(" │ ")

	parts := []string{
		lipgloss.NewStyle().Bold(true).Foreground(colorGreen).Render("📡 Antenna"),
	}

	// Total count
	parts = append(parts, sep)
	parts = append(parts, lipgloss.NewStyle().Bold(true).Foreground(colorWhite).Render(fmt.Sprintf("%d", m.dashboard.TotalCount))+
		lipgloss.NewStyle().Foreground(colorDim).Render(" sessions"))

	// Active
	parts = append(parts, sep)
	parts = append(parts, lipgloss.NewStyle().Foreground(colorGreen).Render("●")+
		lipgloss.NewStyle().Bold(true).Foreground(colorGreen).Render(fmt.Sprintf(" %d", len(active)))+
		lipgloss.NewStyle().Foreground(colorDim).Render(" active"))

	// Idle
	_ = idle

	// Sub-agents
	parts = append(parts, sep)
	parts = append(parts, lipgloss.NewStyle().Foreground(colorPurple).Render("●")+
		lipgloss.NewStyle().Bold(true).Foreground(colorPurple).Render(fmt.Sprintf(" %d", len(subs)))+
		lipgloss.NewStyle().Foreground(colorDim).Render(" sub"))

	// Cron
	parts = append(parts, sep)
	parts = append(parts, lipgloss.NewStyle().Foreground(colorOrange).Render("●")+
		lipgloss.NewStyle().Bold(true).Foreground(colorOrange).Render(fmt.Sprintf(" %d", len(crons)))+
		lipgloss.NewStyle().Foreground(colorDim).Render(" cron"))

	left := strings.Join(parts, "")

	// Cost info on right
	todayCost := lipgloss.NewStyle().Foreground(colorDim).Render("Today ") +
		lipgloss.NewStyle().Bold(true).Foreground(colorGreen).Render(fmt.Sprintf("$%.2f", m.dashboard.TodayCost))
	totalCost := lipgloss.NewStyle().Foreground(colorDim).Render("  Total ") +
		lipgloss.NewStyle().Bold(true).Foreground(colorWhite).Render(fmt.Sprintf("$%.2f", m.dashboard.TotalCost))
	right := todayCost + totalCost

	// Pad to fill width
	leftLen := lipgloss.Width(left)
	rightLen := lipgloss.Width(right)
	gap := w - leftLen - rightLen
	if gap < 1 {
		gap = 1
	}

	bar := left + strings.Repeat(" ", gap) + right
	divider := lipgloss.NewStyle().Foreground(colorBorder).Render(strings.Repeat("─", w))

	return bar + "\n" + divider
}

func (m model) renderDashboard(w, h int) string {
	var b strings.Builder

	// ── Activity Chart (full width) ──
	b.WriteString(m.renderActivityChart(w))
	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().Foreground(colorBorder).Render(strings.Repeat("─", w)))
	b.WriteString("\n")

	active, idle, subs, crons := m.grouped()

	// Calculate layout: if wide enough, use two columns (sessions left, sub/cron right)
	useColumns := w >= 100
	var leftW, rightW int
	if useColumns {
		rightW = clampInt(w*35/100, 30, 60)
		leftW = w - rightW - 3 // 3 for separator
	} else {
		leftW = w
	}

	// Remaining height for sessions
	usedRows := 4 // stats bar + divider + chart + divider
	footerRows := 2
	availRows := h - usedRows - footerRows
	if availRows < 5 {
		availRows = 5
	}

	// Left panel: Active + Idle sessions
	leftLines := m.renderSessionList(active, idle, leftW, availRows)

	if useColumns {
		// Right panel: Sub-agents + Cron
		rightLines := m.renderSidePanels(subs, crons, rightW, availRows)

		// Merge columns
		maxLines := maxInt(len(leftLines), len(rightLines))
		sep := lipgloss.NewStyle().Foreground(colorBorder).Render(" │ ")
		for i := 0; i < maxLines && i < availRows; i++ {
			left := ""
			if i < len(leftLines) {
				left = leftLines[i]
			}
			right := ""
			if i < len(rightLines) {
				right = rightLines[i]
			}
			// Pad left to exact width
			left = padRight(left, leftW)
			right = padRight(right, rightW)
			b.WriteString(left + sep + right + "\n")
		}
	} else {
		// Single column: all sections stacked
		for i, line := range leftLines {
			if i >= availRows {
				break
			}
			b.WriteString(line + "\n")
		}
		// Then sub/cron below
		sideLines := m.renderSidePanels(subs, crons, w, availRows-len(leftLines))
		for _, line := range sideLines {
			b.WriteString(line + "\n")
		}
	}

	// Footer
	b.WriteString("\n")
	help := lipgloss.NewStyle().Foreground(colorDim).Render("  j/k navigate  enter details  r refresh  tab toggle  q quit")
	b.WriteString(help)

	return b.String()
}

func (m model) renderActivityChart(w int) string {
	label := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#777777")).Render("  24H ACTIVITY ")
	chartW := w - lipgloss.Width(label) - 2
	if chartW < 10 {
		chartW = 10
	}

	spark := m.renderSparkline(chartW)
	return label + spark
}

func (m model) renderSessionList(active, idle []api.Session, w, maxRows int) []string {
	var lines []string
	cursor := m.cursor

	// Active section
	if len(active) > 0 {
		header := lipgloss.NewStyle().Foreground(colorGreen).Bold(true).Render("  ● ACTIVE")
		lines = append(lines, header)
		for i, s := range active {
			line := m.renderSessionRow(s, w, cursor == i)
			lines = append(lines, line)
		}
		lines = append(lines, "") // spacer
	}

	// Idle section
	idleOffset := len(active)
	header := lipgloss.NewStyle().Foreground(colorDim).Render("  ○ IDLE") +
		lipgloss.NewStyle().Foreground(colorDim).Render(fmt.Sprintf(" %d", len(idle)))
	lines = append(lines, header)

	for i, s := range idle {
		globalIdx := idleOffset + i
		line := m.renderSessionRow(s, w, cursor == globalIdx)
		// Dim idle rows
		if cursor != globalIdx {
			line = lipgloss.NewStyle().Foreground(colorDim).Render(stripAnsi(line))
			line = m.renderSessionRowDim(s, w)
		}
		lines = append(lines, line)
	}

	return lines
}

func (m model) renderSessionRow(s api.Session, w int, selected bool) string {
	// Adaptive column widths based on terminal width
	nameW := clampInt(w*30/100, 15, 40)
	modelW := clampInt(w*15/100, 8, 25)

	dot := lipgloss.NewStyle().Foreground(colorGreen).Render("●")
	if !s.IsActive {
		dot = lipgloss.NewStyle().Foreground(colorDim).Render("○")
	}

	name := truncate(s.Name, nameW)
	name = fmt.Sprintf("%-*s", nameW, name)

	mdl := truncate(modelDisplay(s.Model), modelW)
	mdl = fmt.Sprintf("%-*s", modelW, mdl)

	msgs := fmt.Sprintf("%4d", s.MessageCount)
	today := fmt.Sprintf("$%.2f", s.TodayCost)
	total := fmt.Sprintf("$%.2f", s.TotalCost)
	ago := timeAgo(s.UpdatedAt)

	var line string
	if w >= 120 {
		line = fmt.Sprintf("  %s %s %s %s %s %s  %s",
			dot,
			lipgloss.NewStyle().Foreground(colorWhite).Render(name),
			lipgloss.NewStyle().Foreground(colorDim).Render(mdl),
			lipgloss.NewStyle().Foreground(colorFg).Render(msgs),
			lipgloss.NewStyle().Foreground(colorGreen).Render(fmt.Sprintf("%8s", today)),
			lipgloss.NewStyle().Foreground(colorFg).Render(fmt.Sprintf("%8s", total)),
			lipgloss.NewStyle().Foreground(colorDim).Render(ago),
		)
	} else if w >= 80 {
		line = fmt.Sprintf("  %s %s %s %s %s",
			dot,
			lipgloss.NewStyle().Foreground(colorWhite).Render(name),
			lipgloss.NewStyle().Foreground(colorFg).Render(msgs),
			lipgloss.NewStyle().Foreground(colorGreen).Render(fmt.Sprintf("%8s", today)),
			lipgloss.NewStyle().Foreground(colorDim).Render(ago),
		)
	} else {
		line = fmt.Sprintf("  %s %s %s",
			dot,
			lipgloss.NewStyle().Foreground(colorWhite).Render(truncate(s.Name, 20)),
			lipgloss.NewStyle().Foreground(colorGreen).Render(today),
		)
	}

	if selected {
		line = lipgloss.NewStyle().
			Background(lipgloss.Color("#1a2a1a")).
			Bold(true).
			Render(padRight(line, w))
	}

	return line
}

func (m model) renderSessionRowDim(s api.Session, w int) string {
	nameW := clampInt(w*30/100, 15, 40)

	name := truncate(s.Name, nameW)
	name = fmt.Sprintf("%-*s", nameW, name)

	msgs := fmt.Sprintf("%4d", s.MessageCount)
	total := fmt.Sprintf("$%.2f", s.TotalCost)
	ago := timeAgo(s.UpdatedAt)

	dim := lipgloss.NewStyle().Foreground(colorDim)

	if w >= 80 {
		return fmt.Sprintf("  %s %s %s %s  %s",
			dim.Render("○"),
			dim.Render(name),
			dim.Render(msgs),
			dim.Render(fmt.Sprintf("%8s", total)),
			dim.Render(ago),
		)
	}
	return fmt.Sprintf("  %s %s %s",
		dim.Render("○"),
		dim.Render(truncate(s.Name, 20)),
		dim.Render(total),
	)
}

func (m model) renderSidePanels(subs, crons []api.Session, w, maxRows int) []string {
	var lines []string

	// Sub-agents
	header := lipgloss.NewStyle().Foreground(colorPurple).Bold(true).Render("  ⚡ SUB-AGENTS") +
		lipgloss.NewStyle().Foreground(colorPurple).Render(fmt.Sprintf(" %d", len(subs)))
	lines = append(lines, header)

	if len(subs) == 0 {
		lines = append(lines, lipgloss.NewStyle().Foreground(colorDim).Render("    None"))
	} else {
		for _, s := range subs {
			lines = append(lines, m.renderCard(s, w, colorPurple))
		}
	}

	lines = append(lines, "") // spacer

	// Cron jobs
	header = lipgloss.NewStyle().Foreground(colorOrange).Bold(true).Render("  ⏱ CRON") +
		lipgloss.NewStyle().Foreground(colorOrange).Render(fmt.Sprintf(" %d", len(crons)))
	lines = append(lines, header)

	if len(crons) == 0 {
		lines = append(lines, lipgloss.NewStyle().Foreground(colorDim).Render("    None"))
	} else {
		for _, s := range crons {
			lines = append(lines, m.renderCard(s, w, colorOrange))
		}
	}

	return lines
}

func (m model) renderCard(s api.Session, w int, accent lipgloss.Color) string {
	nameW := clampInt(w-20, 10, 40)
	name := truncate(s.Name, nameW)

	activeDot := ""
	if s.IsActive {
		activeDot = lipgloss.NewStyle().Foreground(colorGreen).Render(" ●")
	}

	meta := fmt.Sprintf("%d msgs  $%.2f", s.MessageCount, s.TotalCost)

	line1 := fmt.Sprintf("    %s%s",
		lipgloss.NewStyle().Foreground(colorFg).Render(name),
		activeDot,
	)
	line2 := fmt.Sprintf("    %s",
		lipgloss.NewStyle().Foreground(colorDim).Render(meta),
	)

	// For wider terminals, put on one line
	if w >= 50 {
		gap := w - lipgloss.Width(line1) - lipgloss.Width(meta) - 6
		if gap < 2 {
			gap = 2
		}
		return line1 + strings.Repeat(" ", gap) + lipgloss.NewStyle().Foreground(colorDim).Render(meta)
	}
	return line1 + "\n" + line2
}

func (m model) renderDetail(w, h int) string {
	s, ok := m.selectedSession()
	if !ok {
		return lipgloss.NewStyle().Foreground(colorDim).Render("  No session selected\n")
	}

	cardW := clampInt(w-4, 40, 100)

	border := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorBorder).
		Padding(1, 2).
		Width(cardW)

	var status string
	if s.IsActive {
		status = lipgloss.NewStyle().Foreground(colorGreen).Bold(true).Render("● Active")
	} else {
		status = lipgloss.NewStyle().Foreground(colorDim).Render("○ Inactive")
	}

	// Kind badge color
	kindColor := colorGreen
	switch s.Kind {
	case "cron":
		kindColor = colorOrange
	case "subagent":
		kindColor = colorPurple
	}

	labelStyle := lipgloss.NewStyle().Foreground(colorDim).Width(12)
	valStyle := lipgloss.NewStyle().Foreground(colorFg)

	content := strings.Join([]string{
		lipgloss.NewStyle().Bold(true).Foreground(colorWhite).Render(s.Name),
		"",
		labelStyle.Render("Status") + "  " + status,
		labelStyle.Render("Kind") + "  " + lipgloss.NewStyle().Foreground(kindColor).Render(s.Kind),
		labelStyle.Render("Model") + "  " + valStyle.Render(modelDisplay(s.Model)),
		labelStyle.Render("Messages") + "  " + valStyle.Render(fmt.Sprintf("%d", s.MessageCount)),
		labelStyle.Render("Today") + "  " + lipgloss.NewStyle().Foreground(colorGreen).Render(fmt.Sprintf("$%.4f", s.TodayCost)),
		labelStyle.Render("Total") + "  " + valStyle.Render(fmt.Sprintf("$%.4f", s.TotalCost)),
		labelStyle.Render("Updated") + "  " + valStyle.Render(
			time.UnixMilli(s.UpdatedAt).Format("2006-01-02 15:04:05")+
				" ("+timeAgo(s.UpdatedAt)+")"),
		labelStyle.Render("Session") + "  " + lipgloss.NewStyle().Foreground(colorDim).Render(s.SessionID),
		"",
		lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#777777")).Render("24H ACTIVITY"),
		m.renderSparkline(clampInt(cardW-6, 20, 80)),
	}, "\n")

	card := border.Render(content)

	// Center the card
	return "\n" + lipgloss.NewStyle().Width(w).Align(lipgloss.Center).Render(card) + "\n\n" +
		lipgloss.NewStyle().Foreground(colorDim).Render("  esc/q back  r refresh  tab toggle")
}

func (m model) renderSparkline(width int) string {
	if len(m.hourly) == 0 {
		return lipgloss.NewStyle().Foreground(colorDim).Render("no data")
	}

	blocks := []rune("▁▂▃▄▅▆▇█")

	maxVal := 0
	for _, h := range m.hourly {
		if h.Messages > maxVal {
			maxVal = h.Messages
		}
	}
	if maxVal == 0 {
		return lipgloss.NewStyle().Foreground(colorDim).Render(strings.Repeat("▁", minInt(len(m.hourly), width)))
	}

	var sb strings.Builder
	count := minInt(len(m.hourly), width)
	for i := len(m.hourly) - count; i < len(m.hourly); i++ {
		ratio := float64(m.hourly[i].Messages) / float64(maxVal)
		idx := int(math.Round(ratio * float64(len(blocks)-1)))
		if idx < 0 {
			idx = 0
		}
		if m.hourly[i].Messages > 0 {
			sb.WriteString(lipgloss.NewStyle().Foreground(colorGreen).Render(string(blocks[idx])))
		} else {
			sb.WriteString(lipgloss.NewStyle().Foreground(colorDim).Render(string(blocks[0])))
		}
	}
	return sb.String()
}

// ── Helpers ──

func timeAgo(ms int64) string {
	d := time.Since(time.UnixMilli(ms))
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	default:
		return fmt.Sprintf("%dd ago", int(d.Hours()/24))
	}
}

func modelDisplay(m string) string {
	if m == "" {
		return "unknown"
	}
	m = strings.TrimPrefix(m, "anthropic/")
	m = strings.TrimPrefix(m, "openai/")
	return m
}

func truncate(s string, max int) string {
	if len(s) > max {
		return s[:max-1] + "…"
	}
	return s
}

func padRight(s string, w int) string {
	cur := lipgloss.Width(s)
	if cur >= w {
		return s
	}
	return s + strings.Repeat(" ", w-cur)
}

func stripAnsi(s string) string {
	// Simple approach - not used in hot path
	return s
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func clampInt(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
