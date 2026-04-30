package main

import (
	"fmt"
	"os"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/lipgloss"
	bov "github.com/madicen/bubble-overlay"
	"github.com/madicen/bubble-overlay/v2"
)

type closeHelloMsg struct{}

type rootModel struct {
	mainView string
	stack    overlayv2.Stack
	width    int
	height   int
}

type helloModal struct{}

func (helloModal) Init() tea.Cmd { return nil }

func (m helloModal) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if msg.String() == "space" {
			return helloModal{}, func() tea.Msg { return closeHelloMsg{} }
		}
	}
	return m, nil
}

func (helloModal) View() tea.View {
	s := lipgloss.NewStyle().
		Width(34).
		Height(10).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")).
		Align(lipgloss.Center, lipgloss.Center).
		Render("overlayv2 + Bubble Tea v2.\n\nesc closes · space pops.")
	return tea.NewView(s)
}

type keyMap struct {
	Quit key.Binding
	Show key.Binding
}

var keys = keyMap{
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
	Show: key.NewBinding(
		key.WithKeys("space"),
		key.WithHelp("space", "open modal"),
	),
}

func (m rootModel) Init() tea.Cmd { return nil }

func (m rootModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case closeHelloMsg:
		var c tea.Cmd
		if m.stack.Depth() > 0 {
			_, c = m.stack.Pop()
		}
		return m, c
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, m.stack.Update(msg)
	case tea.KeyPressMsg:
		if key.Matches(msg, keys.Quit) {
			return m, tea.Quit
		}
		if !m.stack.MainReceivesKeys() {
			return m, m.stack.Update(msg)
		}
		if key.Matches(msg, keys.Show) {
			return m, m.stack.Push(helloModal{}, bov.DefaultOverlayConfig())
		}
		if msg.String() == "esc" {
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m rootModel) View() tea.View {
	w, h := m.width, m.height
	if w == 0 || h == 0 {
		w, h = 80, 25
	}
	v := m.stack.CompositeView(m.mainView, w, h)
	v.AltScreen = true
	return v
}

func main() {
	const s = "Bubble Tea v2 + overlayv2. Press space for overlay."
	var lines []string
	for i := range 19 {
		lines = append(lines, fmt.Sprintf("%-4d %s", i, s))
	}
	m := rootModel{mainView: strings.Join(lines, "\n")}
	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
