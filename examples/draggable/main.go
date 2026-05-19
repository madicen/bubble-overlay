package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	ov "github.com/madicen/bubble-overlay"
)

type panelModal struct{}

func (panelModal) Init() tea.Cmd { return nil }

func (panelModal) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return panelModal{}, nil
}

func (panelModal) View() string {
	return "Drag the tab or use Alt+arrows to move.\n" +
		"Drag edges or Alt+Shift+arrows to resize.\n" +
		"Click [x] or press esc to close.\n\n" +
		"q quits when no window is open."
}

type rootModel struct {
	mainView string
	stack    ov.OverlayStack
	width    int
	height   int
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
		key.WithHelp("space", "open window"),
	),
}

func draggableConfig() ov.OverlayConfig {
	cfg := ov.DefaultOverlayConfig()
	cfg.WindowChrome = ov.EnableWindowChrome("Draggable window")
	cfg.WindowChrome.TabBorder = "63"
	cfg.WindowChrome.TabBackground = ov.MutedTabBackground(cfg.WindowChrome.TabBorder)
	cfg.WindowChrome.TabForeground = "252"
	cfg.WindowChrome.Resizable = true
	cfg.WindowChrome.Keyboard = true
	cfg.WindowChrome.CenterContent = true
	cfg.WindowChrome.ContentPadTop = 1
	return cfg
}

func (m *rootModel) Init() tea.Cmd { return nil }

func (m *rootModel) viewport() (w, h int) {
	w, h = m.width, m.height
	if w == 0 {
		w = 80
	}
	if h == 0 {
		h = 25
	}
	return w, h
}

func (m *rootModel) syncStackViewport() tea.Cmd {
	w, h := m.viewport()
	return m.stack.Update(tea.WindowSizeMsg{Width: w, Height: h})
}

func (m *rootModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, m.stack.Update(msg)

	case tea.KeyMsg:
		if key.Matches(msg, keys.Quit) {
			return m, tea.Quit
		}
		if m.stack.Depth() > 0 {
			return m, m.stack.Update(msg)
		}
		if key.Matches(msg, keys.Show) {
			return m, tea.Batch(
				m.stack.Push(panelModal{}, draggableConfig()),
				m.syncStackViewport(),
			)
		}
		if msg.Type == tea.KeyEsc || msg.Type == tea.KeyEscape {
			return m, tea.Quit
		}

	case tea.MouseMsg:
		if m.stack.Depth() > 0 {
			return m, m.stack.Update(msg)
		}
	}
	return m, nil
}

func (m *rootModel) View() string {
	w, h := m.viewport()
	return m.stack.View(m.mainView, w, h)
}

func main() {
	const s = "Main view. Press space for a draggable window."
	var lines []string
	for i := range 19 {
		lines = append(lines, fmt.Sprintf("%-4d %s", i, s))
	}
	m := &rootModel{
		mainView: strings.Join(lines, "\n"),
	}
	opts := []tea.ProgramOption{
		tea.WithAltScreen(),
		tea.WithMouseAllMotion(),
	}
	p := tea.NewProgram(m, opts...)
	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
