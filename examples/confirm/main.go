package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	overlay "github.com/madicen/bubble-overlay"
)

type model struct {
	mainView  string
	showModal bool
	confirmed bool
	width     int
	height    int
}

type keyMap struct {
	Quit key.Binding
	Show key.Binding
	Yes  key.Binding
	No   key.Binding
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
	Yes: key.NewBinding(
		key.WithKeys("y"),
		key.WithHelp("y", "yes"),
	),
	No: key.NewBinding(
		key.WithKeys("n"),
		key.WithHelp("n", "no"),
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
		if m.showModal && !m.confirmed {
			switch {
			case key.Matches(msg, keys.Yes):
				m.confirmed = true
				return m, nil
			case key.Matches(msg, keys.No):
				m.confirmed = false
				m.showModal = false
				return m, nil
			}
		}
		switch {
		case key.Matches(msg, keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, keys.Show):
			m.showModal = !m.showModal
			if m.showModal {
				m.confirmed = false
			}
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

	var content string
	if m.confirmed {
		content = "You confirmed!"
	} else {
		question := "Are you sure?"
		yes := "Yes (y)"
		no := "No (n)"
		content = fmt.Sprintf("%s %s / %s", question, yes, no)
	}

	modal := lipgloss.NewStyle().
		Width(30).
		Align(lipgloss.Center).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")).
		Render(content)

	return overlay.OverlayViewInCenter(m.mainView, modal, w, h)
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
