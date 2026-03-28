package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	overlay "github.com/madicen/bubble-overlay"
)

type model struct {
	mainView  string
	showModal bool
	textInput textinput.Model
	submitted bool
	width     int
	height    int
}

type keyMap struct {
	Quit   key.Binding
	Show   key.Binding
	Submit key.Binding
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
	Submit: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "submit"),
	),
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		if m.showModal && !m.submitted {
			switch {
			case key.Matches(msg, keys.Submit):
				m.submitted = true
				return m, nil
			}
		}
		switch {
		case key.Matches(msg, keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, keys.Show):
			m.showModal = !m.showModal
			if m.showModal {
				m.submitted = false
				m.textInput.SetValue("")
				m.textInput.Focus()
			} else {
				m.textInput.Blur()
			}
			return m, nil
		}
	}

	if m.showModal && !m.submitted {
		m.textInput, cmd = m.textInput.Update(msg)
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
	if m.submitted {
		content = fmt.Sprintf("You submitted: %s", m.textInput.Value())
	} else {
		content = fmt.Sprintf("Enter something:%s%s", m.textInput.View(), "(enter to submit)")
	}

	modal := lipgloss.NewStyle().
		Width(40).
		Align(lipgloss.Center).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")).
		Render(content)

	return overlay.OverlayViewInCenter(m.mainView, modal, w, h)
}

func main() {
	ti := textinput.New()
	ti.Placeholder = "Hello, world!"
	ti.CharLimit = 156
	ti.Width = 30

	const s = "This is the main view. Press the spacebar to show the modal."
	var lines []string
	for i := range 19 {
		lines = append(lines, fmt.Sprintf("%-4d %s", i, s))
	}
	m := model{
		mainView:  strings.Join(lines, "\n"),
		textInput: ti,
	}
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
