// Package ui is the terminal interface: gauges for CPU/RAM and a process table
// with CPU, RAM and network columns that can be sorted either way.
// It only reads data through the api package.
package ui

import (
	"fmt"
	"slices"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"linux-system-monitor/api"
)

const (
	refreshEvery = 2 * time.Second
	tableRows    = 20

	pidWidth = 7  // PID column
	numWidth = 12 // each CPU / RAM / DOWN / UP column
	maxName  = 32 // widest NAME column
	minName  = 10 // narrowest NAME column on small terminals
)

type column struct {
	rt    api.ResourceType
	title string
}

// sortColumns are the sortable columns, left to right (also the tab order).
var sortColumns = []column{
	{api.CPU, "CPU"},
	{api.RAM, "RAM"},
	{api.NETDown, "DOWN"},
	{api.NETUp, "UP"},
}

var (
	titleStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39"))
	labelStyle  = lipgloss.NewStyle().Width(5)
	headerStyle = lipgloss.NewStyle().Bold(true).Reverse(true)
	helpStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	okStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	warnStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
	critStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
)

type startedMsg struct{}
type tickMsg struct{}

// Model is the Bubble Tea model.
type Model struct {
	snap          *api.Snapshot
	sortBy        api.ResourceType
	desc          bool // true: highest first
	width, height int
	ready         bool
}

func New() Model { return Model{sortBy: api.CPU, desc: true} }

// sortOn sorts by rt. Choosing the column that is already active flips the
// direction instead.
func (m Model) sortOn(rt api.ResourceType) Model {
	if m.sortBy == rt {
		m.desc = !m.desc
	} else {
		m.sortBy, m.desc = rt, true
	}
	return m
}

// cycleSort moves the sort column step places along sortColumns (wrapping),
// starting with the highest values first.
func (m Model) cycleSort(step int) Model {
	i := slices.IndexFunc(sortColumns, func(c column) bool { return c.rt == m.sortBy })
	n := len(sortColumns)
	m.sortBy, m.desc = sortColumns[(i+step+n)%n].rt, true
	return m
}

// Init starts the api's background collector. Its first update blocks for
// ~1s, so it runs as a Cmd. The monitor runs until the program exits.
func (m Model) Init() tea.Cmd {
	return func() tea.Msg {
		api.Start(time.Second)
		return startedMsg{}
	}
}

// tick schedules the next read of the latest snapshot.
func tick() tea.Cmd {
	return tea.Tick(refreshEvery, func(time.Time) tea.Msg { return tickMsg{} })
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
	case startedMsg:
		m.snap, m.ready = api.GetSnapshot(), true
		return m, tick()
	case tickMsg:
		m.snap = api.GetSnapshot() // cheap read; the api refreshes itself
		return m, tick()
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "tab":
			m = m.cycleSort(1)
		case "shift+tab":
			m = m.cycleSort(-1)
		case "c":
			m = m.sortOn(api.CPU)
		case "m":
			m = m.sortOn(api.RAM)
		case "d":
			m = m.sortOn(api.NETDown)
		case "u":
			m = m.sortOn(api.NETUp)
		case "r":
			m.desc = !m.desc
		}
	}
	return m, nil
}

func (m Model) View() string {
	if !m.ready {
		return "Loading...\n"
	}
	barW := 40
	if m.width > 0 && m.width-20 < barW {
		barW = max(10, m.width-20)
	}

	var b strings.Builder
	free := m.snap.RAMTotal * (1 - m.snap.RAMUsedPercent/100)
	fmt.Fprintf(&b, "%s\n\n", titleStyle.Render("Linux System Monitor"))
	fmt.Fprintf(&b, "%s%s\n", labelStyle.Render("CPU"), gauge(m.snap.CPUPercent, barW))
	fmt.Fprintf(&b, "%s%s%s\n\n", labelStyle.Render("RAM"), gauge(m.snap.RAMUsedPercent, barW),
		helpStyle.Render(fmt.Sprintf("  %s free", api.FormatRAM(free))))

	b.WriteString(m.table())
	fmt.Fprintf(&b, "\n%s\n",
		helpStyle.Render("sort: c cpu  m ram  d down  u up (again = reverse)  tab next  r reverse  q quit"))
	return b.String()
}

func (m Model) table() string {
	// NAME takes whatever width the other columns leave, within [minName, maxName].
	nameW := maxName
	if m.width > 0 {
		fixed := pidWidth + len(sortColumns)*numWidth + len(sortColumns) + 1 // columns + separators
		nameW = min(maxName, max(minName, m.width-fixed))
	}

	var header strings.Builder
	fmt.Fprintf(&header, "%-*s %-*s", pidWidth, "PID", nameW, "NAME")
	for _, c := range sortColumns {
		title := c.title
		if c.rt == m.sortBy {
			if m.desc {
				title += " ▼"
			} else {
				title += " ▲"
			}
		}
		fmt.Fprintf(&header, " %*s", numWidth, title)
	}

	var b strings.Builder
	fmt.Fprintf(&b, "%s\n", headerStyle.Render(header.String()))
	for _, p := range m.snap.Sorted(m.sortBy, m.desc)[:min(tableRows, len(m.snap.Processes))] {
		fmt.Fprintf(&b, "%-*d %-*s", pidWidth, p.PID, nameW, truncate(p.Name, nameW))
		for _, c := range sortColumns {
			fmt.Fprintf(&b, " %*s", numWidth, p.FormatConsumption(c.rt))
		}
		b.WriteByte('\n')
	}
	return b.String()
}

// gauge renders a percent bar coloured by load.
func gauge(percent float64, width int) string {
	percent = min(max(percent, 0), 100)
	filled := int(percent / 100 * float64(width))
	style := okStyle
	switch {
	case percent >= 85:
		style = critStyle
	case percent >= 60:
		style = warnStyle
	}
	bar := style.Render(strings.Repeat("█", filled)) + helpStyle.Render(strings.Repeat("░", width-filled))
	return fmt.Sprintf("%s %5.1f%%", bar, percent)
}

func truncate(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n-1]) + "…"
}
