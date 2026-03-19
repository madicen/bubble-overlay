package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	overlay "github.com/diceguyd30/bubble-overlay"
)

type model struct {
	mainView  string
	showModal bool
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

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			return m, nil
		}
	}
	return m, nil
}

func (m model) View() string {
	w, h := m.width, m.height
	if w == 0 || h == 0 {
		w, h = 80, 25
	}

	if !m.showModal {
		return m.mainView
	}

	modal := lipgloss.NewStyle().
		Width(30).
		Height(10).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")).
		Render("Hello, from the modal! Press 'q' or 'esc' to quit.")

	return overlay.OverlayView(m.mainView, modal, w, h, 5, 10)
}

func main() {
	const s = "This is the main view. Press the spacebar to show the modal."
	var lines []string
	for i := range 19 {
		lines = append(lines, fmt.Sprintf("%-4d %s", i, s))
	}
	m := model{
		mainView: strings.Join(lines, "\n"),
	}
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
