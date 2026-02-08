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

// Gmork theme colors
var (
	colorGreen   = lipgloss.Color("#4EC990")
	colorOrange  = lipgloss.Color("#E5A84B")
	colorRed     = lipgloss.Color("#E55B5B")
	colorDim     = lipgloss.Color("#555555")
	colorFg      = lipgloss.Color("#CCCCCC")
	colorBg      = lipgloss.Color("#0A0A0A")
	colorAccent  = lipgloss.Color("#7EC8E3")
	colorCyan    = lipgloss.Color("#56B6C2")

	styleTitle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorGreen).
			Padding(0, 1)

	styleHeader = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorAccent)

	styleDim = lipgloss.NewStyle().
			Foreground(colorDim)

	styleActive = lipgloss.NewStyle().
			Foreground(colorGreen).
			Bold(true)

	styleSelected = lipgloss.NewStyle().
			Background(lipgloss.Color("#1A2A1A")).
			Foreground(colorFg).
			Bold(true)

	styleKindMain = lipgloss.NewStyle().
			Foreground(colorGreen)

	styleKindCron = lipgloss.NewStyle().
			Foreground(colorOrange)

	styleKindSub = lipgloss.NewStyle().
			Foreground(colorCyan)

	styleCost = lipgloss.NewStyle().
			Foreground(colorOrange)

	styleBar = lipgloss.NewStyle().
			Foreground(colorGreen)

	styleHelp = lipgloss.NewStyle().
			Foreground(colorDim).
			Padding(1, 1)
)

type view int

const (
	viewList view = iota
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
	err       error
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
				m.view = viewList
				return m, nil
			}
			return m, tea.Quit
		case "j", "down":
			if m.view == viewList && m.cursor < len(m.dashboard.Sessions)-1 {
				m.cursor++
			}
			return m, nil
		case "k", "up":
			if m.view == viewList && m.cursor > 0 {
				m.cursor--
			}
			return m, nil
		case "enter":
			if m.view == viewList && len(m.dashboard.Sessions) > 0 {
				m.view = viewDetail
			}
			return m, nil
		case "esc", "backspace":
			m.view = viewList
			return m, nil
		case "tab":
			if m.view == viewList {
				m.view = viewDetail
			} else {
				m.view = viewList
			}
			return m, nil
		case "r":
			m.dashboard = m.client.GetDashboard()
			m.hourly = m.client.GetHourlyActivity()
			return m, nil
		}

	case tickMsg:
		m.dashboard = m.client.GetDashboard()
		m.hourly = m.client.GetHourlyActivity()
		if m.cursor >= len(m.dashboard.Sessions) && len(m.dashboard.Sessions) > 0 {
			m.cursor = len(m.dashboard.Sessions) - 1
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

	// Header
	header := styleTitle.Render("📡 Antenna") +
		styleDim.Render(" — OpenClaw Monitor") +
		"  " +
		styleCost.Render(fmt.Sprintf("Today: $%.4f", m.dashboard.TodayCost)) +
		styleDim.Render(fmt.Sprintf("  Total: $%.4f  Sessions: %d", m.dashboard.TotalCost, m.dashboard.TotalCount))
	b.WriteString(header + "\n")
	b.WriteString(styleDim.Render(strings.Repeat("─", min(w, 120))) + "\n")

	switch m.view {
	case viewList:
		b.WriteString(m.renderList(w, h-5))
	case viewDetail:
		b.WriteString(m.renderDetail(w, h-5))
	}

	// Footer
	b.WriteString("\n")
	if m.view == viewList {
		b.WriteString(styleHelp.Render("j/k: navigate  enter: details  r: refresh  tab: toggle  q: quit"))
	} else {
		b.WriteString(styleHelp.Render("esc/q: back  r: refresh  tab: toggle"))
	}

	return b.String()
}

func (m model) renderList(w, maxRows int) string {
	var b strings.Builder

	// Activity sparkline
	b.WriteString(styleHeader.Render("  24h Activity ") + " ")
	b.WriteString(m.renderSparkline(min(w-18, 48)) + "\n\n")

	// Column headers
	hdr := fmt.Sprintf("  %-3s %-30s %-10s %-8s %-10s %-10s %s",
		"", "NAME", "KIND", "MSGS", "TODAY", "TOTAL", "UPDATED")
	b.WriteString(styleDim.Render(hdr) + "\n")

	for i, s := range m.dashboard.Sessions {
		if i >= maxRows-4 {
			b.WriteString(styleDim.Render(fmt.Sprintf("  ... and %d more", len(m.dashboard.Sessions)-i)) + "\n")
			break
		}

		// Status indicator
		var dot string
		if s.IsActive {
			dot = styleActive.Render("●")
		} else {
			dot = styleDim.Render("○")
		}

		// Kind badge
		var kind string
		switch s.Kind {
		case "main":
			kind = styleKindMain.Render("main   ")
		case "cron":
			kind = styleKindCron.Render("cron   ")
		case "subagent":
			kind = styleKindSub.Render("sub    ")
		default:
			kind = styleDim.Render(fmt.Sprintf("%-7s", s.Kind))
		}

		name := s.Name
		if len(name) > 28 {
			name = name[:27] + "…"
		}

		ago := timeAgo(s.UpdatedAt)

		line := fmt.Sprintf("  %s %-30s %s %-8d %-10s %-10s %s",
			dot,
			name,
			kind,
			s.MessageCount,
			styleCost.Render(fmt.Sprintf("$%.4f", s.TodayCost)),
			styleCost.Render(fmt.Sprintf("$%.4f", s.TotalCost)),
			styleDim.Render(ago),
		)

		if i == m.cursor {
			line = styleSelected.Render(line)
		}
		b.WriteString(line + "\n")
	}

	if len(m.dashboard.Sessions) == 0 {
		b.WriteString(styleDim.Render("  No sessions found. Is OpenClaw running?\n"))
	}

	return b.String()
}

func (m model) renderDetail(w, maxRows int) string {
	if m.cursor >= len(m.dashboard.Sessions) {
		return styleDim.Render("  No session selected\n")
	}

	s := m.dashboard.Sessions[m.cursor]
	var b strings.Builder

	// Status
	var status string
	if s.IsActive {
		status = styleActive.Render("● Active")
	} else {
		status = styleDim.Render("○ Inactive")
	}

	b.WriteString(styleHeader.Render("  Session Detail") + "\n\n")
	b.WriteString(fmt.Sprintf("  Name:      %s\n", styleTitle.Render(s.Name)))
	b.WriteString(fmt.Sprintf("  Status:    %s\n", status))
	b.WriteString(fmt.Sprintf("  Kind:      %s\n", s.Kind))
	b.WriteString(fmt.Sprintf("  Model:     %s\n", modelDisplay(s.Model)))
	b.WriteString(fmt.Sprintf("  Messages:  %d\n", s.MessageCount))
	b.WriteString(fmt.Sprintf("  Today:     %s\n", styleCost.Render(fmt.Sprintf("$%.4f", s.TodayCost))))
	b.WriteString(fmt.Sprintf("  Total:     %s\n", styleCost.Render(fmt.Sprintf("$%.4f", s.TotalCost))))
	b.WriteString(fmt.Sprintf("  Updated:   %s (%s)\n",
		time.UnixMilli(s.UpdatedAt).Format("2006-01-02 15:04:05"),
		timeAgo(s.UpdatedAt)))
	b.WriteString(fmt.Sprintf("  Session:   %s\n", styleDim.Render(s.SessionID)))

	b.WriteString("\n")
	b.WriteString(styleHeader.Render("  24h Activity") + "\n")
	b.WriteString("  " + m.renderSparkline(min(w-4, 48)) + "\n")

	return b.String()
}

func (m model) renderSparkline(width int) string {
	if len(m.hourly) == 0 {
		return styleDim.Render("no data")
	}

	blocks := []rune("▁▂▃▄▅▆▇█")

	// Find max
	maxVal := 0
	for _, h := range m.hourly {
		if h.Messages > maxVal {
			maxVal = h.Messages
		}
	}
	if maxVal == 0 {
		return styleDim.Render(strings.Repeat("▁", min(len(m.hourly), width)))
	}

	var sb strings.Builder
	count := min(len(m.hourly), width)
	for i := len(m.hourly) - count; i < len(m.hourly); i++ {
		ratio := float64(m.hourly[i].Messages) / float64(maxVal)
		idx := int(math.Round(ratio * float64(len(blocks)-1)))
		if idx < 0 {
			idx = 0
		}
		if m.hourly[i].Messages > 0 {
			sb.WriteString(styleBar.Render(string(blocks[idx])))
		} else {
			sb.WriteString(styleDim.Render(string(blocks[0])))
		}
	}

	return sb.String()
}

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
		return styleDim.Render("unknown")
	}
	// Shorten common prefixes
	m = strings.TrimPrefix(m, "anthropic/")
	m = strings.TrimPrefix(m, "openai/")
	return m
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
