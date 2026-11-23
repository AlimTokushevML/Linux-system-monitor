// Package ui is the terminal interface: gauges for CPU/RAM and a process table.
// It only reads data through the api package.
package ui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"linux-system-monitor/api"
)

const (
	refreshEvery = 2 * time.Second
	tableRows    = 20
)

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
	width, height int
	ready         bool
}

func New() Model { return Model{sortBy: api.CPU} }

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
			m.sortBy = (m.sortBy + 1) % 2
		case "c":
			m.sortBy = api.CPU
		case "m":
			m.sortBy = api.RAM
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
	b.WriteString(titleStyle.Render("Linux System Monitor") + "\n\n")
	free := m.snap.RAMTotal * (1 - m.snap.RAMUsedPercent/100)
	b.WriteString(labelStyle.Render("CPU") + gauge(m.snap.CPUPercent, barW) + "\n")
	b.WriteString(labelStyle.Render("RAM") + gauge(m.snap.RAMUsedPercent, barW) +
		helpStyle.Render(fmt.Sprintf("  %s free", api.FormatRAM(free))) + "\n\n")

	b.WriteString(m.table())
	b.WriteString("\n" + helpStyle.Render("tab sort  c cpu  m ram  q quit") + "\n")
	return b.String()
}

func (m Model) table() string {
	arrow := func(rt api.ResourceType) string {
		if m.sortBy == rt {
			return " ▼"
		}
		return ""
	}
	var b strings.Builder
	b.WriteString(headerStyle.Render(fmt.Sprintf("%-8s %-32s %12s %12s",
		"PID", "NAME", "CPU"+arrow(api.CPU), "RAM"+arrow(api.RAM))) + "\n")
	for _, p := range m.snap.Top(tableRows, m.sortBy) {
		b.WriteString(fmt.Sprintf("%-8d %-32s %12s %12s\n",
			p.PID, truncate(p.Name, 32), p.FormatConsumption(api.CPU), p.FormatConsumption(api.RAM)))
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
