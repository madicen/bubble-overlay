package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	ov "github.com/madicen/bubble-overlay"
)

type dismissConfirmMsg struct{}

type rootModel struct {
	mainView string
	stack    ov.OverlayStack
	width    int
	height   int
}

type confirmModal struct {
	confirmed bool
}

func (m confirmModal) Init() tea.Cmd { return nil }

func (m confirmModal) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Yes):
			if !m.confirmed {
				return confirmModal{confirmed: true}, nil
			}
		case key.Matches(msg, keys.No):
			return m, func() tea.Msg { return dismissConfirmMsg{} }
		case msg.Type == tea.KeyEsc || msg.Type == tea.KeyEscape:
			return m, func() tea.Msg { return dismissConfirmMsg{} }
		case key.Matches(msg, keys.Show):
			return m, func() tea.Msg { return dismissConfirmMsg{} }
		}
	}
	return m, nil
}

func (m confirmModal) View() string {
	var content string
	if m.confirmed {
		content = "You confirmed!\n\n(space to close)"
	} else {
		content = "Are you sure?\n\nYes (y) / No (n)"
	}
	return lipgloss.NewStyle().
		Width(30).
		Align(lipgloss.Center).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")).
		Render(content)
}

type keyMap struct {
	Quit key.Binding
	Show key.Binding
	Yes  key.Binding
	No   key.Binding
}

var keys = keyMap{
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
	Show: key.NewBinding(
		key.WithKeys(" "),
		key.WithHelp("space", "show / hide modal"),
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

func (m rootModel) Init() tea.Cmd { return nil }

func (m rootModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case dismissConfirmMsg:
		var c tea.Cmd
		if m.stack.Depth() > 0 {
			_, c = m.stack.Pop()
		}
		return m, c
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, m.stack.Update(msg)
	case tea.KeyMsg:
		if key.Matches(msg, keys.Quit) {
			return m, tea.Quit
		}
		if !m.stack.MainReceivesKeyMsg() {
			return m, m.stack.Update(msg)
		}
		if key.Matches(msg, keys.Show) {
			cfg := ov.DefaultOverlayConfig()
			cfg.CloseOnEscape = false
			return m, m.stack.Push(confirmModal{}, cfg)
		}
		if msg.Type == tea.KeyEsc || msg.Type == tea.KeyEscape {
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m rootModel) View() string {
	w, h := m.width, m.height
	if w == 0 || h == 0 {
		w, h = 80, 25
	}
	return m.stack.View(m.mainView, w, h)
}

func main() {
	const s = "Press space for the modal."
	var lines []string
	for i := range 19 {
		lines = append(lines, fmt.Sprintf("%-4d %s", i, s))
	}
	m := rootModel{mainView: strings.Join(lines, "\n")}
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
