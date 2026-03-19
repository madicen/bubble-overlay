package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	overlay "github.com/madicen/bubble-overlay"
)

type model struct {
	mainView  string
	showModal bool
	spinner   spinner.Model
	done      bool
	width     int
	height    int
}

type keyMap struct {
	Quit key.Binding
	Show key.Binding
}

var keys = keyMap{
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c", "esc"),
		key.WithHelp("q", "quit"),
	),
	Show: key.NewBinding(
		key.WithKeys(" "),
		key.WithHelp("space", "show modal"),
	),
}

type taskFinishedMsg struct{}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, keys.Show):
			m.showModal = !m.showModal
			if m.showModal {
				m.done = false
				return m, tea.Batch(m.spinner.Tick, m.doTask())
			}
			return m, nil
		}
	case taskFinishedMsg:
		m.done = true
		return m, nil
	}

	if m.showModal && !m.done {
		m.spinner, cmd = m.spinner.Update(msg)
	}

	return m, cmd
}

func (m model) View() string {
	w, h := m.width, m.height
	if w == 0 || h == 0 {
		w, h = 80, 25
	}

	if !m.showModal {
		return m.mainView
	}

	var content string
	if m.done {
		content = "Task finished!"
	} else {
		content = fmt.Sprintf("%s Loading...", m.spinner.View())
	}

	modal := lipgloss.NewStyle().
		Width(30).
		Align(lipgloss.Center).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")).
		Render(content)

	return overlay.OverlayView(m.mainView, modal, w, h, 5, 10)
}

func (m *model) doTask() tea.Cmd {
	return tea.Tick(time.Second*2, func(t time.Time) tea.Msg {
		return taskFinishedMsg{}
	})
}

func main() {
	s := spinner.New()
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	const str = "This is the main view. Press the spacebar to show the modal."
	var lines []string
	for i := range 19 {
		lines = append(lines, fmt.Sprintf("%-4d %s", i, str))
	}
	m := model{
		mainView: strings.Join(lines, "\n"),
		spinner:  s,
	}
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
