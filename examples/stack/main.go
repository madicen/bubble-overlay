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

// nestedOpenMsg is sent by the outer overlay so the root can push the inner dialog.
type nestedOpenMsg struct{}

type rootModel struct {
	mainView string
	stack    ov.OverlayStack
	width    int
	height   int
	devMode  bool
}

type outerModal struct{}

func (outerModal) Init() tea.Cmd { return nil }

func (outerModal) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if k, ok := msg.(tea.KeyMsg); ok && k.String() == "i" {
		return outerModal{}, func() tea.Msg { return nestedOpenMsg{} }
	}
	return outerModal{}, nil
}

func (outerModal) View() string {
	return lipgloss.NewStyle().
		Width(40).
		Height(12).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("39")).
		Padding(0, 1).
		Render("Outer overlay\n\nPress i to open a nested dialog\n(right drawer).\n\nesc closes one layer at a time.")
}

type innerModal struct{}

func (innerModal) Init() tea.Cmd { return nil }

func (innerModal) Update(tea.Msg) (tea.Model, tea.Cmd) { return innerModal{}, nil }

func (innerModal) View() string {
	return lipgloss.NewStyle().
		Width(26).
		Height(8).
		Border(lipgloss.DoubleBorder()).
		BorderForeground(lipgloss.Color("205")).
		Padding(0, 1).
		Render("Inner (nested)\n\nesc pops this layer first.")
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
		key.WithKeys(" "),
		key.WithHelp("space", "open outer overlay"),
	),
}

func (m rootModel) Init() tea.Cmd { return nil }

func (m rootModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case nestedOpenMsg:
		innerCfg := ov.DefaultOverlayConfig()
		innerCfg.Placement = ov.RightDrawer()
		innerCfg.DimOpacity = 0.45
		return m, m.stack.Push(innerModal{}, innerCfg)

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
			outerCfg := ov.DefaultOverlayConfig()
			outerCfg.DimOpacity = 0.3
			return m, m.stack.Push(outerModal{}, outerCfg)
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
	// Bubble Tea expects View() to occupy exactly h rows. Appending a footer line
	// after stack.View(..., h) produced h+1 rows and caused a one-line scroll/jitter.
	if m.devMode && h > 1 {
		core := m.stack.View(m.mainView, w, h-1)
		foot := lipgloss.NewStyle().
			Width(w).
			Faint(true).
			Render(ov.DevStackDepthFooter(m.stack.StackDepth(), true))
		return lipgloss.JoinVertical(lipgloss.Left, core, foot)
	}
	return m.stack.View(m.mainView, w, h)
}

func main() {
	dev := os.Getenv("OVERLAY_DEV") == "1"
	const s = "space: outer overlay · i (inside outer): inner · esc: pop layer · q: quit"
	var lines []string
	for i := range 17 {
		lines = append(lines, fmt.Sprintf("%-4d %s", i, s))
	}
	m := rootModel{
		mainView: strings.Join(lines, "\n"),
		devMode:  dev,
	}
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
