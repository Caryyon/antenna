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

	// Highlight colors
	colorHighlightBg = lipgloss.Color("#1a3a1a") // dark green bg for selected row
	colorActiveBorder = lipgloss.Color("#00ff99") // bright green for focused section border
	colorFocusedHeader = lipgloss.Color("#00ffd5") // cyan for focused section header
)

type view int

const (
	viewDashboard view = iota
	viewDetail
)

// Sections for navigation
const (
	sectionChart    = 0
	sectionSessions = 1
	sectionSubs     = 2
	sectionCrons    = 3
)

type tickMsg time.Time

type model struct {
	client    *api.Client
	dashboard api.DashboardData
	hourly    []api.HourlyBucket
	view      view
	width     int
	height    int
	interval  time.Duration
	err       error

	// Section-based navigation
	section     int   // which section is focused
	sectionCur  [4]int // cursor within each section (chart has no cursor)
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

// sessionsForSection returns the list of sessions for a given section
func (m model) sessionsForSection(sec int) []api.Session {
	active, idle, subs, crons := m.grouped()
	switch sec {
	case sectionSessions:
		all := make([]api.Session, 0, len(active)+len(idle))
		all = append(all, active...)
		all = append(all, idle...)
		return all
	case sectionSubs:
		return subs
	case sectionCrons:
		return crons
	}
	return nil
}

func (m model) sectionLen(sec int) int {
	return len(m.sessionsForSection(sec))
}

func (m model) selectedSession() (api.Session, bool) {
	sessions := m.sessionsForSection(m.section)
	cur := m.sectionCur[m.section]
	if cur >= 0 && cur < len(sessions) {
		return sessions[cur], true
	}
	return api.Session{}, false
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
		section:   sectionSessions,
	}
}

func tickCmd(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(t time.Time) tea.Msg { return tickMsg(t) })
}

func (m model) Init() tea.Cmd {
	return tickCmd(m.interval)
}

// sectionGrid defines spatial layout for ctrl+hjkl navigation
// Layout:
//   [chart]    (top, full width)
//   [sessions] [subs]
//              [crons]
var sectionRight = map[int]int{
	sectionChart:    -1,
	sectionSessions: sectionSubs,
	sectionSubs:     -1,
	sectionCrons:    -1,
}
var sectionLeft = map[int]int{
	sectionChart:    -1,
	sectionSessions: -1,
	sectionSubs:     sectionSessions,
	sectionCrons:    sectionSessions,
}
var sectionDown = map[int]int{
	sectionChart:    sectionSessions,
	sectionSessions: -1,
	sectionSubs:     sectionCrons,
	sectionCrons:    -1,
}
var sectionUp = map[int]int{
	sectionChart:    -1,
	sectionSessions: sectionChart,
	sectionSubs:     sectionChart,
	sectionCrons:    sectionSubs,
}

func (m *model) moveSection(dir map[int]int) {
	next, ok := dir[m.section]
	if ok && next >= 0 {
		m.section = next
	}
}

func (m *model) clampCursors() {
	for i := 1; i <= 3; i++ {
		max := m.sectionLen(i)
		if max == 0 {
			m.sectionCur[i] = 0
		} else if m.sectionCur[i] >= max {
			m.sectionCur[i] = max - 1
		}
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		key := msg.String()
		switch key {
		case "q", "ctrl+c":
			if m.view == viewDetail {
				m.view = viewDashboard
				return m, nil
			}
			return m, tea.Quit

		// vim navigation within section
		case "j", "down":
			if m.view == viewDashboard && m.section >= 1 {
				max := m.sectionLen(m.section)
				if m.sectionCur[m.section] < max-1 {
					m.sectionCur[m.section]++
				}
			}
		case "k", "up":
			if m.view == viewDashboard && m.section >= 1 {
				if m.sectionCur[m.section] > 0 {
					m.sectionCur[m.section]--
				}
			}
		case "h":
			if m.view == viewDashboard {
				m.moveSection(sectionLeft)
			}
		case "l":
			if m.view == viewDashboard {
				m.moveSection(sectionRight)
			}

		// ctrl+hjkl for section navigation
		case "ctrl+h":
			if m.view == viewDashboard {
				m.moveSection(sectionLeft)
			}
		case "ctrl+l":
			if m.view == viewDashboard {
				m.moveSection(sectionRight)
			}
		case "ctrl+j":
			if m.view == viewDashboard {
				m.moveSection(sectionDown)
			}
		case "ctrl+k":
			if m.view == viewDashboard {
				m.moveSection(sectionUp)
			}

		case "enter":
			if m.view == viewDashboard && m.section >= 1 {
				if _, ok := m.selectedSession(); ok {
					m.view = viewDetail
				}
			}
		case "esc", "backspace":
			m.view = viewDashboard
		case "tab":
			if m.view == viewDashboard {
				// Cycle through sections: sessions -> subs -> crons -> sessions
				switch m.section {
				case sectionChart:
					m.section = sectionSessions
				case sectionSessions:
					m.section = sectionSubs
				case sectionSubs:
					m.section = sectionCrons
				case sectionCrons:
					m.section = sectionSessions
				}
			}
		case "r":
			m.dashboard = m.client.GetDashboard()
			m.hourly = m.client.GetHourlyActivity()
			m.clampCursors()
		}
		return m, nil

	case tickMsg:
		m.dashboard = m.client.GetDashboard()
		m.hourly = m.client.GetHourlyActivity()
		m.clampCursors()
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
	active, _, subs, crons := m.grouped()

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
	chartFocused := m.section == sectionChart
	b.WriteString(m.renderActivityChart(w, chartFocused))
	b.WriteString("\n")

	// Divider between chart and lists
	divColor := colorBorder
	b.WriteString(lipgloss.NewStyle().Foreground(divColor).Render(strings.Repeat("─", w)))
	b.WriteString("\n")

	active, idle, subs, crons := m.grouped()

	// Calculate layout
	useColumns := w >= 100
	var leftW, rightW int
	if useColumns {
		rightW = clampInt(w*35/100, 30, 60)
		leftW = w - rightW - 3
	} else {
		leftW = w
	}

	usedRows := 16 // stats bar(2) + chart(~12) + divider(1) + gap(1)
	footerRows := 3
	availRows := h - usedRows - footerRows
	if availRows < 5 {
		availRows = 5
	}

	// Left panel: Active + Idle sessions
	leftLines := m.renderSessionList(active, idle, leftW, availRows)

	if useColumns {
		// Right panel: Sub-agents + Cron
		rightLines := m.renderSidePanels(subs, crons, rightW, availRows)

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
			left = padRight(left, leftW)
			right = padRight(right, rightW)
			b.WriteString(left + sep + right + "\n")
		}
	} else {
		for i, line := range leftLines {
			if i >= availRows {
				break
			}
			b.WriteString(line + "\n")
		}
		sideLines := m.renderSidePanels(subs, crons, w, availRows-len(leftLines))
		for _, line := range sideLines {
			b.WriteString(line + "\n")
		}
	}

	// Footer with section indicator
	b.WriteString("\n")
	sectionNames := []string{"chart", "sessions", "sub-agents", "cron"}
	var sectionIndicators []string
	for i, name := range sectionNames {
		if i == m.section {
			sectionIndicators = append(sectionIndicators,
				lipgloss.NewStyle().Bold(true).Foreground(colorCyan).Render("["+name+"]"))
		} else {
			sectionIndicators = append(sectionIndicators,
				lipgloss.NewStyle().Foreground(colorDim).Render(" "+name+" "))
		}
	}
	help := lipgloss.NewStyle().Foreground(colorDim).Render("  hjkl move  ctrl+hjkl section  enter details  tab cycle  r refresh  q quit")
	b.WriteString("  " + strings.Join(sectionIndicators, lipgloss.NewStyle().Foreground(colorDim).Render("·")) + "\n")
	b.WriteString(help)

	return b.String()
}

func (m model) renderActivityChart(w int, focused bool) string {
	headerColor := lipgloss.Color("#777777")
	if focused {
		headerColor = colorFocusedHeader
	}
	indicator := "  "
	if focused {
		indicator = "▶ "
	}
	header := lipgloss.NewStyle().Bold(true).Foreground(headerColor).Render(indicator + "24H ACTIVITY")

	return header + "\n" + m.renderBarChart(w)
}

func (m model) renderBarChart(w int) string {
	chartHeight := 8
	blocks := []rune(" ▁▂▃▄▅▆▇█")

	// Determine data: use hourly buckets, map to w columns
	data := m.hourly
	if len(data) == 0 {
		return lipgloss.NewStyle().Foreground(colorDim).Render("  no activity data\n")
	}

	// Reserve left margin for Y-axis labels (e.g. "  42 ") and 1 right margin
	yLabelW := 5
	barAreaW := w - yLabelW - 1
	if barAreaW < 10 {
		barAreaW = 10
	}

	// Map data buckets to bar columns. Each column represents one or more buckets.
	numBuckets := len(data)
	values := make([]float64, barAreaW)
	counts := make([]int, barAreaW)
	for i, b := range data {
		col := i * barAreaW / numBuckets
		if col >= barAreaW {
			col = barAreaW - 1
		}
		values[col] += float64(b.Messages)
		counts[col]++
	}
	// Average where multiple buckets map to same column
	for i := range values {
		if counts[i] > 1 {
			values[i] /= float64(counts[i])
		}
	}

	maxVal := 0.0
	for _, v := range values {
		if v > maxVal {
			maxVal = v
		}
	}
	if maxVal == 0 {
		maxVal = 1
	}

	// Y-axis: compute label values for top and mid rows
	topLabel := int(math.Ceil(maxVal))
	midLabel := topLabel / 2

	var sb strings.Builder
	dimStyle := lipgloss.NewStyle().Foreground(colorDim)
	barStyle := lipgloss.NewStyle().Foreground(colorGreen)
	zeroStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#1a3a1a"))

	for row := chartHeight; row >= 1; row-- {
		// Y-axis label
		threshold := float64(row) / float64(chartHeight)
		label := ""
		if row == chartHeight {
			label = fmt.Sprintf("%4d", topLabel)
		} else if row == chartHeight/2 {
			label = fmt.Sprintf("%4d", midLabel)
		} else {
			label = "    "
		}
		sb.WriteString(dimStyle.Render(label + " "))

		// Bar characters for this row
		for col := 0; col < barAreaW; col++ {
			ratio := values[col] / maxVal
			// How much of this row is filled (each row = 1/chartHeight of full scale)
			rowBottom := float64(row-1) / float64(chartHeight)
			rowTop := float64(row) / float64(chartHeight)
			_ = threshold

			if ratio >= rowTop {
				// Full block
				sb.WriteString(barStyle.Render("█"))
			} else if ratio > rowBottom {
				// Partial block
				frac := (ratio - rowBottom) / (rowTop - rowBottom)
				idx := int(math.Round(frac * 8))
				if idx < 1 {
					idx = 1
				}
				if idx > 8 {
					idx = 8
				}
				sb.WriteString(barStyle.Render(string(blocks[idx])))
			} else if values[col] > 0 && row == 1 {
				// Show minimum bar for non-zero values
				sb.WriteString(zeroStyle.Render("▁"))
			} else {
				sb.WriteString(" ")
			}
		}
		sb.WriteString("\n")
	}

	// X-axis line
	sb.WriteString(dimStyle.Render(strings.Repeat(" ", yLabelW) + strings.Repeat("─", barAreaW)))
	sb.WriteString("\n")

	// Time labels along the bottom
	xLabels := strings.Repeat(" ", yLabelW)
	if numBuckets > 0 {
		// Place hour labels at evenly spaced positions
		labelBuf := make([]byte, barAreaW)
		for i := range labelBuf {
			labelBuf[i] = ' '
		}
		// Every 3-6 hours depending on space
		step := 6
		if barAreaW > 80 {
			step = 3
		} else if barAreaW > 40 {
			step = 4
		}
		for i := 0; i < numBuckets; i++ {
			hourStr := data[i].Hour // e.g. "15:00"
			hourNum := 0
			fmt.Sscanf(hourStr, "%d", &hourNum)
			if hourNum%step == 0 {
				col := i * barAreaW / numBuckets
				lbl := fmt.Sprintf("%02d", hourNum)
				if col+2 <= barAreaW {
					labelBuf[col] = lbl[0]
					if col+1 < barAreaW {
						labelBuf[col+1] = lbl[1]
					}
				}
			}
		}
		xLabels += string(labelBuf)
	}
	sb.WriteString(dimStyle.Render(xLabels))

	return sb.String()
}

func (m model) renderSessionList(active, idle []api.Session, w, maxRows int) []string {
	var lines []string
	focused := m.section == sectionSessions
	cursor := m.sectionCur[sectionSessions]

	// Active section header
	headerStyle := lipgloss.NewStyle().Foreground(colorGreen).Bold(true)
	if focused {
		headerStyle = headerStyle.Foreground(colorCyan)
	}

	if len(active) > 0 {
		indicator := "  "
		if focused {
			indicator = "▶ "
		}
		header := headerStyle.Render(indicator + "● ACTIVE")
		lines = append(lines, header)
		for i, s := range active {
			line := m.renderSessionRow(s, w, focused && cursor == i, focused)
			lines = append(lines, line)
		}
		lines = append(lines, "")
	}

	// Idle section
	idleOffset := len(active)
	idleHeaderColor := colorDim
	if focused {
		idleHeaderColor = lipgloss.Color("#888888")
	}
	header := lipgloss.NewStyle().Foreground(idleHeaderColor).Render("  ○ IDLE") +
		lipgloss.NewStyle().Foreground(colorDim).Render(fmt.Sprintf(" %d", len(idle)))
	lines = append(lines, header)

	for i, s := range idle {
		globalIdx := idleOffset + i
		selected := focused && cursor == globalIdx
		if selected {
			line := m.renderSessionRow(s, w, true, true)
			lines = append(lines, line)
		} else {
			line := m.renderSessionRowDim(s, w)
			lines = append(lines, line)
		}
	}

	return lines
}

func (m model) renderSessionRow(s api.Session, w int, selected bool, sectionFocused bool) string {
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

	prefix := "  "
	if selected {
		prefix = " ▶"
	}

	// Dim text when section not focused
	nameColor := colorWhite
	msgColor := colorFg
	if !sectionFocused {
		nameColor = colorFg
		msgColor = colorDim
	}

	var line string
	if w >= 120 {
		line = fmt.Sprintf("%s %s %s %s %s %s %s  %s",
			prefix,
			dot,
			lipgloss.NewStyle().Foreground(nameColor).Render(name),
			lipgloss.NewStyle().Foreground(colorDim).Render(mdl),
			lipgloss.NewStyle().Foreground(msgColor).Render(msgs),
			lipgloss.NewStyle().Foreground(colorGreen).Render(fmt.Sprintf("%8s", today)),
			lipgloss.NewStyle().Foreground(msgColor).Render(fmt.Sprintf("%8s", total)),
			lipgloss.NewStyle().Foreground(colorDim).Render(ago),
		)
	} else if w >= 80 {
		line = fmt.Sprintf("%s %s %s %s %s %s",
			prefix,
			dot,
			lipgloss.NewStyle().Foreground(nameColor).Render(name),
			lipgloss.NewStyle().Foreground(msgColor).Render(msgs),
			lipgloss.NewStyle().Foreground(colorGreen).Render(fmt.Sprintf("%8s", today)),
			lipgloss.NewStyle().Foreground(colorDim).Render(ago),
		)
	} else {
		line = fmt.Sprintf("%s %s %s %s",
			prefix,
			dot,
			lipgloss.NewStyle().Foreground(nameColor).Render(truncate(s.Name, 20)),
			lipgloss.NewStyle().Foreground(colorGreen).Render(today),
		)
	}

	if selected {
		line = lipgloss.NewStyle().
			Background(colorHighlightBg).
			Foreground(colorWhite).
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
		return fmt.Sprintf("   %s %s %s %s  %s",
			dim.Render("○"),
			dim.Render(name),
			dim.Render(msgs),
			dim.Render(fmt.Sprintf("%8s", total)),
			dim.Render(ago),
		)
	}
	return fmt.Sprintf("   %s %s %s",
		dim.Render("○"),
		dim.Render(truncate(s.Name, 20)),
		dim.Render(total),
	)
}

func (m model) renderSidePanels(subs, crons []api.Session, w, maxRows int) []string {
	var lines []string

	// Sub-agents
	subsFocused := m.section == sectionSubs
	subHeaderColor := colorPurple
	if subsFocused {
		subHeaderColor = colorCyan
	}
	indicator := "  "
	if subsFocused {
		indicator = "▶ "
	}
	header := lipgloss.NewStyle().Foreground(subHeaderColor).Bold(true).Render(indicator+"⚡ SUB-AGENTS") +
		lipgloss.NewStyle().Foreground(colorPurple).Render(fmt.Sprintf(" %d", len(subs)))
	lines = append(lines, header)

	if len(subs) == 0 {
		lines = append(lines, lipgloss.NewStyle().Foreground(colorDim).Render("    None"))
	} else {
		for i, s := range subs {
			selected := subsFocused && m.sectionCur[sectionSubs] == i
			lines = append(lines, m.renderCard(s, w, colorPurple, selected, subsFocused))
		}
	}

	lines = append(lines, "")

	// Cron jobs
	cronsFocused := m.section == sectionCrons
	cronHeaderColor := colorOrange
	if cronsFocused {
		cronHeaderColor = colorCyan
	}
	indicator = "  "
	if cronsFocused {
		indicator = "▶ "
	}
	header = lipgloss.NewStyle().Foreground(cronHeaderColor).Bold(true).Render(indicator+"⏱ CRON") +
		lipgloss.NewStyle().Foreground(colorOrange).Render(fmt.Sprintf(" %d", len(crons)))
	lines = append(lines, header)

	if len(crons) == 0 {
		lines = append(lines, lipgloss.NewStyle().Foreground(colorDim).Render("    None"))
	} else {
		for i, s := range crons {
			selected := cronsFocused && m.sectionCur[sectionCrons] == i
			lines = append(lines, m.renderCard(s, w, colorOrange, selected, cronsFocused))
		}
	}

	return lines
}

func (m model) renderCard(s api.Session, w int, accent lipgloss.Color, selected bool, sectionFocused bool) string {
	nameW := clampInt(w-20, 10, 40)
	name := truncate(s.Name, nameW)

	activeDot := ""
	if s.IsActive {
		activeDot = lipgloss.NewStyle().Foreground(colorGreen).Render(" ●")
	}

	meta := fmt.Sprintf("%d msgs  $%.2f", s.MessageCount, s.TotalCost)

	prefix := "   "
	if selected {
		prefix = " ▶ "
	}

	nameColor := colorFg
	if selected {
		nameColor = colorWhite
	} else if !sectionFocused {
		nameColor = colorDim
	}

	line1 := fmt.Sprintf("%s%s%s",
		prefix,
		lipgloss.NewStyle().Foreground(nameColor).Render(name),
		activeDot,
	)

	if w >= 50 {
		gap := w - lipgloss.Width(line1) - lipgloss.Width(meta) - 4
		if gap < 2 {
			gap = 2
		}
		metaColor := colorDim
		result := line1 + strings.Repeat(" ", gap) + lipgloss.NewStyle().Foreground(metaColor).Render(meta)
		if selected {
			result = lipgloss.NewStyle().
				Background(colorHighlightBg).
				Bold(true).
				Render(padRight(result, w))
		}
		return result
	}

	line2 := fmt.Sprintf("    %s",
		lipgloss.NewStyle().Foreground(colorDim).Render(meta),
	)
	result := line1 + "\n" + line2
	if selected {
		result = lipgloss.NewStyle().
			Background(colorHighlightBg).
			Bold(true).
			Render(padRight(line1, w)) + "\n" +
			lipgloss.NewStyle().
				Background(colorHighlightBg).
				Render(padRight(line2, w))
	}
	return result
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
