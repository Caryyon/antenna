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
	colorVoid    = lipgloss.Color("#050505")
	colorPanel   = lipgloss.Color("#0a0a0a")
	colorSurface = lipgloss.Color("#111111")
	colorBorder  = lipgloss.Color("#1a1a1a")
	colorGreen   = lipgloss.Color("#00ff99")
	colorCyan    = lipgloss.Color("#00ffd5")
	colorPurple  = lipgloss.Color("#bf6fff")
	colorOrange  = lipgloss.Color("#ff8c4c")
	colorRed     = lipgloss.Color("#ff4477")
	colorDim     = lipgloss.Color("#555555")
	colorDimmer  = lipgloss.Color("#333333")
	colorFg      = lipgloss.Color("#bbbbbb")
	colorWhite   = lipgloss.Color("#ffffff")

	// Selection
	colorSelectBg = lipgloss.Color("#0a2a1a")
)

type view int

const (
	viewDashboard view = iota
	viewDetail
)

// Sections for navigation (matches web layout grid)
const (
	sectionActive = 0 // left top
	sectionIdle   = 1 // left bottom
	sectionSubs   = 2 // right top
	sectionCrons  = 3 // right bottom
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

	section    int    // focused section
	sectionCur [4]int // cursor per section
}

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

func (m model) sessionsForSection(sec int) []api.Session {
	active, idle, subs, crons := m.grouped()
	switch sec {
	case sectionActive:
		return active
	case sectionIdle:
		return idle
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
		section:   sectionActive,
	}
}

func tickCmd(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(t time.Time) tea.Msg { return tickMsg(t) })
}

func (m model) Init() tea.Cmd {
	return tickCmd(m.interval)
}

// Navigation grid:
//   [active]  [subs]
//   [idle]    [crons]
var navRight = [4]int{sectionSubs, sectionCrons, -1, -1}
var navLeft = [4]int{-1, -1, sectionActive, sectionIdle}
var navDown = [4]int{sectionIdle, -1, sectionCrons, -1}
var navUp = [4]int{-1, sectionActive, -1, sectionSubs}

func (m *model) moveSection(grid [4]int) {
	next := grid[m.section]
	if next >= 0 {
		m.section = next
	}
}

func (m *model) clampCursors() {
	for i := 0; i < 4; i++ {
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

		case "j", "down":
			if m.view == viewDashboard {
				max := m.sectionLen(m.section)
				if m.sectionCur[m.section] < max-1 {
					m.sectionCur[m.section]++
				}
			}
		case "k", "up":
			if m.view == viewDashboard {
				if m.sectionCur[m.section] > 0 {
					m.sectionCur[m.section]--
				}
			}
		case "h":
			if m.view == viewDashboard {
				m.moveSection(navLeft)
			}
		case "l":
			if m.view == viewDashboard {
				m.moveSection(navRight)
			}
		case "ctrl+j":
			if m.view == viewDashboard {
				m.moveSection(navDown)
			}
		case "ctrl+k":
			if m.view == viewDashboard {
				m.moveSection(navUp)
			}
		case "ctrl+h":
			if m.view == viewDashboard {
				m.moveSection(navLeft)
			}
		case "ctrl+l":
			if m.view == viewDashboard {
				m.moveSection(navRight)
			}
		case "enter":
			if m.view == viewDashboard {
				if _, ok := m.selectedSession(); ok {
					m.view = viewDetail
				}
			}
		case "esc", "backspace":
			m.view = viewDashboard
		case "tab":
			if m.view == viewDashboard {
				order := []int{sectionActive, sectionSubs, sectionIdle, sectionCrons}
				for i, s := range order {
					if s == m.section {
						m.section = order[(i+1)%len(order)]
						break
					}
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
		w = 120
	}
	h := m.height
	if h == 0 {
		h = 40
	}

	var b strings.Builder

	switch m.view {
	case viewDashboard:
		b.WriteString(m.renderStatsBar(w))
		b.WriteString("\n")
		b.WriteString(m.renderDashboard(w, h))
	case viewDetail:
		b.WriteString(m.renderStatsBar(w))
		b.WriteString("\n")
		b.WriteString(m.renderDetail(w, h))
	}

	return b.String()
}

// ── Stats Bar ──
func (m model) renderStatsBar(w int) string {
	active, _, subs, crons := m.grouped()

	sep := lipgloss.NewStyle().Foreground(colorDimmer).Render(" │ ")

	// Pulsing green dot + "Live"
	live := lipgloss.NewStyle().Foreground(colorGreen).Render("● ") +
		lipgloss.NewStyle().Foreground(colorGreen).Bold(true).Render("Live")

	// Big session count
	count := lipgloss.NewStyle().Bold(true).Foreground(colorWhite).Render(fmt.Sprintf("%d", m.dashboard.TotalCount)) +
		lipgloss.NewStyle().Foreground(colorDim).Render(" sessions")

	// Colored counts
	activeCount := lipgloss.NewStyle().Bold(true).Foreground(colorGreen).Render(fmt.Sprintf("%d", len(active))) +
		lipgloss.NewStyle().Foreground(colorDim).Render(" active")
	subCount := lipgloss.NewStyle().Bold(true).Foreground(colorPurple).Render(fmt.Sprintf("%d", len(subs))) +
		lipgloss.NewStyle().Foreground(colorDim).Render(" sub")
	cronCount := lipgloss.NewStyle().Bold(true).Foreground(colorOrange).Render(fmt.Sprintf("%d", len(crons))) +
		lipgloss.NewStyle().Foreground(colorDim).Render(" cron")

	left := live + sep + count + sep + activeCount + sep + subCount + sep + cronCount

	// Right: costs
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

// ── Dashboard ──
func (m model) renderDashboard(w, h int) string {
	var b strings.Builder

	// Activity chart (full width)
	b.WriteString(m.renderActivityChart(w))
	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().Foreground(colorBorder).Render(strings.Repeat("─", w)))
	b.WriteString("\n")

	active, idle, subs, crons := m.grouped()

	// Two-column layout
	useColumns := w >= 90
	var leftW, rightW int
	if useColumns {
		rightW = clampInt(w/3, 28, 55)
		leftW = w - rightW - 1 // 1 for separator
	} else {
		leftW = w
		rightW = w
	}

	// Calculate available rows
	chartRows := 12 // approx: header + 8 bars + axis + labels + divider
	statsRows := 2
	footerRows := 2
	availRows := h - statsRows - chartRows - footerRows - 1
	if availRows < 8 {
		availRows = 8
	}

	// Build left panel: active + idle
	leftLines := m.renderLeftPanel(active, idle, leftW, availRows)

	if useColumns {
		// Build right panel: subs + crons
		rightLines := m.renderRightPanel(subs, crons, rightW, availRows)

		maxLines := maxInt(len(leftLines), len(rightLines))
		sep := lipgloss.NewStyle().Foreground(colorDimmer).Render("│")

		for i := 0; i < maxLines && i < availRows; i++ {
			left := ""
			if i < len(leftLines) {
				left = leftLines[i]
			}
			right := ""
			if i < len(rightLines) {
				right = rightLines[i]
			}
			b.WriteString(padRight(left, leftW) + sep + padRight(right, rightW) + "\n")
		}
	} else {
		for i, line := range leftLines {
			if i >= availRows/2 {
				break
			}
			b.WriteString(line + "\n")
		}
		rightLines := m.renderRightPanel(subs, crons, rightW, availRows/2)
		for _, line := range rightLines {
			b.WriteString(line + "\n")
		}
	}

	// Footer
	b.WriteString("\n")
	footerDim := lipgloss.NewStyle().Foreground(colorDimmer)
	footerKey := lipgloss.NewStyle().Foreground(colorDim)
	b.WriteString(footerDim.Render(" ") +
		footerKey.Render("j/k") + footerDim.Render(" move  ") +
		footerKey.Render("h/l") + footerDim.Render(" column  ") +
		footerKey.Render("ctrl+j/k") + footerDim.Render(" section  ") +
		footerKey.Render("enter") + footerDim.Render(" detail  ") +
		footerKey.Render("tab") + footerDim.Render(" cycle  ") +
		footerKey.Render("r") + footerDim.Render(" refresh  ") +
		footerKey.Render("q") + footerDim.Render(" quit"))

	return b.String()
}

// ── Section Header ──
// Renders an uppercase header with colored left border glow effect
func sectionHeader(title string, count int, accent lipgloss.Color, focused bool) string {
	borderChar := "┃"
	glowChar := "░"

	borderColor := accent
	if !focused {
		borderColor = lipgloss.Color("#333333")
	}

	border := lipgloss.NewStyle().Foreground(borderColor).Render(borderChar)
	glow := lipgloss.NewStyle().Foreground(borderColor).Render(glowChar)

	titleStyle := lipgloss.NewStyle().
		Foreground(accent).
		Bold(true)

	if !focused {
		titleStyle = titleStyle.Foreground(colorDim)
	}

	countStr := lipgloss.NewStyle().Foreground(accent).Render(fmt.Sprintf(" %d", count))
	if !focused {
		countStr = lipgloss.NewStyle().Foreground(colorDim).Render(fmt.Sprintf(" %d", count))
	}

	return border + glow + " " + titleStyle.Render(title) + countStr
}

// ── Left Panel: Active + Idle ──
func (m model) renderLeftPanel(active, idle []api.Session, w, maxRows int) []string {
	var lines []string
	activeFocused := m.section == sectionActive
	idleFocused := m.section == sectionIdle

	// Active header
	lines = append(lines, sectionHeader("● ACTIVE SESSIONS", len(active), colorGreen, activeFocused))

	if len(active) == 0 {
		lines = append(lines, renderBorderedLine("    "+lipgloss.NewStyle().Foreground(colorDim).Render("No active sessions"), colorGreen, activeFocused))
	} else {
		for i, s := range active {
			selected := activeFocused && m.sectionCur[sectionActive] == i
			lines = append(lines, m.renderSessionRow(s, w, selected, activeFocused, colorGreen))
		}
	}

	lines = append(lines, "")

	// Idle header
	lines = append(lines, sectionHeader("○ IDLE", len(idle), lipgloss.Color("#666666"), idleFocused))

	if len(idle) == 0 {
		lines = append(lines, renderBorderedLine("    "+lipgloss.NewStyle().Foreground(colorDim).Render("No idle sessions"), lipgloss.Color("#666666"), idleFocused))
	} else {
		for i, s := range idle {
			selected := idleFocused && m.sectionCur[sectionIdle] == i
			if selected {
				lines = append(lines, m.renderSessionRow(s, w, true, true, lipgloss.Color("#666666")))
			} else {
				lines = append(lines, m.renderSessionRowDim(s, w, idleFocused))
			}
		}
	}

	return lines
}

// renderBorderedLine renders a line with left border accent
func renderBorderedLine(content string, accent lipgloss.Color, focused bool) string {
	borderColor := accent
	if !focused {
		borderColor = lipgloss.Color("#333333")
	}
	border := lipgloss.NewStyle().Foreground(borderColor).Render("┃")
	return border + " " + content
}

// ── Right Panel: Sub-agents + Cron ──
func (m model) renderRightPanel(subs, crons []api.Session, w, maxRows int) []string {
	var lines []string
	subsFocused := m.section == sectionSubs
	cronsFocused := m.section == sectionCrons

	// Sub-agents header
	lines = append(lines, sectionHeader("⚡ SUB-AGENTS", len(subs), colorPurple, subsFocused))

	if len(subs) == 0 {
		lines = append(lines, renderBorderedLine("   "+lipgloss.NewStyle().Foreground(colorDim).Render("None"), colorPurple, subsFocused))
	} else {
		for i, s := range subs {
			selected := subsFocused && m.sectionCur[sectionSubs] == i
			lines = append(lines, m.renderCard(s, w, colorPurple, selected, subsFocused))
		}
	}

	lines = append(lines, "")

	// Cron header
	lines = append(lines, sectionHeader("⏱  CRON JOBS", len(crons), colorOrange, cronsFocused))

	if len(crons) == 0 {
		lines = append(lines, renderBorderedLine("   "+lipgloss.NewStyle().Foreground(colorDim).Render("None"), colorOrange, cronsFocused))
	} else {
		for i, s := range crons {
			selected := cronsFocused && m.sectionCur[sectionCrons] == i
			lines = append(lines, m.renderCard(s, w, colorOrange, selected, cronsFocused))
		}
	}

	return lines
}

// ── Activity Chart ──
func (m model) renderActivityChart(w int) string {
	headerStyle := lipgloss.NewStyle().Foreground(colorDim).Bold(true)
	header := headerStyle.Render("  ▌ 24H ACTIVITY")
	return header + "\n" + m.renderBarChart(w)
}

func (m model) renderBarChart(w int) string {
	chartHeight := 8
	blocks := []rune(" ▁▂▃▄▅▆▇█")

	data := m.hourly
	if len(data) == 0 {
		return lipgloss.NewStyle().Foreground(colorDim).Render("  no activity data\n")
	}

	yLabelW := 6
	barAreaW := w - yLabelW - 1
	if barAreaW < 10 {
		barAreaW = 10
	}

	numBuckets := len(data)
	values := make([]float64, barAreaW)
	costs := make([]float64, barAreaW)
	counts := make([]int, barAreaW)
	for i, b := range data {
		col := i * barAreaW / numBuckets
		if col >= barAreaW {
			col = barAreaW - 1
		}
		values[col] += float64(b.Messages)
		costs[col] += b.Cost
		counts[col]++
	}
	for i := range values {
		if counts[i] > 1 {
			values[i] /= float64(counts[i])
			costs[i] /= float64(counts[i])
		}
	}

	maxVal := 0.0
	maxCost := 0.0
	for i, v := range values {
		if v > maxVal {
			maxVal = v
		}
		if costs[i] > maxCost {
			maxCost = costs[i]
		}
	}
	if maxVal == 0 {
		maxVal = 1
	}
	if maxCost == 0 {
		maxCost = 1
	}

	topLabel := int(math.Ceil(maxVal))
	midLabel := topLabel / 2

	var sb strings.Builder
	dimStyle := lipgloss.NewStyle().Foreground(colorDimmer)
	barStyle := lipgloss.NewStyle().Foreground(colorGreen)
	costStyle := lipgloss.NewStyle().Foreground(colorPurple)
	zeroStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#0a3a1a"))

	for row := chartHeight; row >= 1; row-- {
		// Y-axis label
		label := "     "
		if row == chartHeight {
			label = fmt.Sprintf("%5d", topLabel)
		} else if row == chartHeight/2 {
			label = fmt.Sprintf("%5d", midLabel)
		}
		sb.WriteString(dimStyle.Render(label + " "))

		rowBottom := float64(row-1) / float64(chartHeight)
		rowTop := float64(row) / float64(chartHeight)

		for col := 0; col < barAreaW; col++ {
			ratio := values[col] / maxVal
			// Cost line overlay: determine if the purple cost dot should appear at this row
			costRatio := costs[col] / maxCost
			costRow := int(math.Round(costRatio * float64(chartHeight)))

			if costRow == row && costs[col] > 0 {
				// Purple cost line takes precedence
				sb.WriteString(costStyle.Render("━"))
			} else if ratio >= rowTop {
				sb.WriteString(barStyle.Render("█"))
			} else if ratio > rowBottom {
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
				sb.WriteString(zeroStyle.Render("▁"))
			} else {
				sb.WriteString(" ")
			}
		}
		sb.WriteString("\n")
	}

	// X-axis
	sb.WriteString(dimStyle.Render(strings.Repeat(" ", yLabelW) + strings.Repeat("─", barAreaW)))
	sb.WriteString("\n")

	// Time labels
	labelBuf := make([]byte, barAreaW)
	for i := range labelBuf {
		labelBuf[i] = ' '
	}
	step := 6
	if barAreaW > 80 {
		step = 3
	} else if barAreaW > 40 {
		step = 4
	}
	for i := 0; i < numBuckets; i++ {
		hourNum := 0
		fmt.Sscanf(data[i].Hour, "%d", &hourNum)
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
	sb.WriteString(dimStyle.Render(strings.Repeat(" ", yLabelW) + string(labelBuf)))

	return sb.String()
}

// ── Session Row ──
func (m model) renderSessionRow(s api.Session, w int, selected bool, sectionFocused bool, accent lipgloss.Color) string {
	nameW := clampInt(w*25/100, 12, 35)
	modelW := clampInt(w*15/100, 8, 22)

	dot := lipgloss.NewStyle().Foreground(colorGreen).Render("●")
	if !s.IsActive {
		dot = lipgloss.NewStyle().Foreground(colorDim).Render("○")
	}

	name := truncate(s.Name, nameW)
	name = fmt.Sprintf("%-*s", nameW, name)

	mdl := truncate(modelDisplay(s.Model), modelW)
	mdl = fmt.Sprintf("%-*s", modelW, mdl)

	msgs := fmt.Sprintf("%3d", s.MessageCount)
	today := fmt.Sprintf("$%.2f", s.TodayCost)
	total := fmt.Sprintf("$%.2f", s.TotalCost)
	ago := timeAgo(s.UpdatedAt)

	cursor := "  "
	if selected {
		cursor = " ▸"
	}

	// Border glow
	borderColor := accent
	if !sectionFocused {
		borderColor = lipgloss.Color("#333333")
	}
	border := lipgloss.NewStyle().Foreground(borderColor).Render("┃")

	nameColor := colorWhite
	if !sectionFocused {
		nameColor = colorFg
	}

	var line string
	if w >= 110 {
		line = fmt.Sprintf("%s%s %s %s %s %s %s %s  %s",
			border, cursor,
			dot,
			lipgloss.NewStyle().Foreground(nameColor).Render(name),
			lipgloss.NewStyle().Foreground(colorDim).Render(mdl),
			lipgloss.NewStyle().Foreground(colorFg).Render(msgs),
			lipgloss.NewStyle().Foreground(colorGreen).Render(fmt.Sprintf("%7s", today)),
			lipgloss.NewStyle().Foreground(colorDim).Render(fmt.Sprintf("%7s", total)),
			lipgloss.NewStyle().Foreground(colorDimmer).Render(ago),
		)
	} else if w >= 70 {
		line = fmt.Sprintf("%s%s %s %s %s %s  %s",
			border, cursor,
			dot,
			lipgloss.NewStyle().Foreground(nameColor).Render(name),
			lipgloss.NewStyle().Foreground(colorFg).Render(msgs),
			lipgloss.NewStyle().Foreground(colorGreen).Render(fmt.Sprintf("%7s", today)),
			lipgloss.NewStyle().Foreground(colorDimmer).Render(ago),
		)
	} else {
		line = fmt.Sprintf("%s%s %s %s %s",
			border, cursor,
			dot,
			lipgloss.NewStyle().Foreground(nameColor).Render(truncate(s.Name, 18)),
			lipgloss.NewStyle().Foreground(colorGreen).Render(today),
		)
	}

	if selected {
		// Highlight: use colored background
		lineContent := padRight(line, w)
		lineContent = lipgloss.NewStyle().
			Background(colorSelectBg).
			Bold(true).
			Render(lineContent)
		return lineContent
	}

	return line
}

func (m model) renderSessionRowDim(s api.Session, w int, sectionFocused bool) string {
	nameW := clampInt(w*25/100, 12, 35)

	name := truncate(s.Name, nameW)
	name = fmt.Sprintf("%-*s", nameW, name)

	msgs := fmt.Sprintf("%3d", s.MessageCount)
	total := fmt.Sprintf("$%.2f", s.TotalCost)
	ago := timeAgo(s.UpdatedAt)

	dim := lipgloss.NewStyle().Foreground(colorDimmer)
	dimFg := lipgloss.NewStyle().Foreground(colorDim)

	borderColor := lipgloss.Color("#222222")
	if sectionFocused {
		borderColor = lipgloss.Color("#444444")
	}
	border := lipgloss.NewStyle().Foreground(borderColor).Render("┃")

	if w >= 70 {
		return fmt.Sprintf("%s   %s %s %s %s  %s",
			border,
			dim.Render("○"),
			dimFg.Render(name),
			dim.Render(msgs),
			dim.Render(fmt.Sprintf("%7s", total)),
			dim.Render(ago),
		)
	}
	return fmt.Sprintf("%s   %s %s %s",
		border,
		dim.Render("○"),
		dimFg.Render(truncate(s.Name, 18)),
		dim.Render(total),
	)
}

// ── Card (Sub-agent / Cron) ──
func (m model) renderCard(s api.Session, w int, accent lipgloss.Color, selected bool, sectionFocused bool) string {
	borderColor := accent
	if !sectionFocused {
		borderColor = lipgloss.Color("#333333")
	}
	border := lipgloss.NewStyle().Foreground(borderColor).Render("┃")

	nameW := clampInt(w-20, 8, 35)
	name := truncate(s.Name, nameW)

	activeDot := ""
	if s.IsActive {
		activeDot = " " + lipgloss.NewStyle().Foreground(colorGreen).Render("●")
	}

	cursor := "  "
	if selected {
		cursor = " ▸"
	}

	nameColor := colorFg
	if selected {
		nameColor = colorWhite
	} else if !sectionFocused {
		nameColor = colorDim
	}

	meta := lipgloss.NewStyle().Foreground(colorDim).Render(
		fmt.Sprintf("%d msgs", s.MessageCount)) +
		"  " +
		lipgloss.NewStyle().Foreground(colorGreen).Render(
			fmt.Sprintf("$%.2f", s.TodayCost))

	line := fmt.Sprintf("%s%s %s%s  %s",
		border, cursor,
		lipgloss.NewStyle().Foreground(nameColor).Render(name),
		activeDot,
		meta,
	)

	if selected {
		line = lipgloss.NewStyle().
			Background(colorSelectBg).
			Bold(true).
			Render(padRight(line, w))
	}

	return line
}

// ── Detail View ──
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
		labelStyle.Render("Session") + "  " + lipgloss.NewStyle().Foreground(colorDimmer).Render(s.SessionID),
		"",
		lipgloss.NewStyle().Bold(true).Foreground(colorDim).Render("24H ACTIVITY"),
		m.renderSparkline(clampInt(cardW-6, 20, 80)),
	}, "\n")

	card := border.Render(content)

	return "\n" + lipgloss.NewStyle().Width(w).Align(lipgloss.Center).Render(card) + "\n\n" +
		lipgloss.NewStyle().Foreground(colorDim).Render("  esc back  r refresh  q quit")
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
			sb.WriteString(lipgloss.NewStyle().Foreground(colorDimmer).Render(string(blocks[0])))
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
	if max <= 0 {
		return ""
	}
	if len(s) > max {
		if max <= 1 {
			return "…"
		}
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
