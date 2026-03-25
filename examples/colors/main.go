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

// Layout tuned so a fixed overlay rectangle cuts through styled cells on row 0.
const (
	graphPadCols  = 14
	modalOverlayW = 44
	modalOverlayH = 10
	// Start the modal this many columns from the left so it covers part of the
	// purple hash and the following cyan tail (SGR for both sits under the box).
	modalLeft = 8
)

type model struct {
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
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
	Show: key.NewBinding(
		key.WithKeys(" "),
		key.WithHelp("space", "toggle modal"),
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

func graphRow() string {
	pad := strings.Repeat(" ", graphPadCols)
	at := lipgloss.NewStyle().Foreground(lipgloss.Color("252")).Render("@ ")
	hash := lipgloss.NewStyle().
		Background(lipgloss.Color("#5b21b6")).
		Foreground(lipgloss.Color("#e9d5ff")).
		Render("573fd5c7 ")
	tail := lipgloss.NewStyle().Foreground(lipgloss.Color("81")).Render("Add feature flag [demo] · Evolog split (z)")
	return pad + at + hash + tail
}

// hintRow is one full-width styled line: a single background span so the modal can
// obscure the opening SGR; overlay must reapply the pen for the visible tail.
func hintRow(viewWidth int) string {
	msg := "Olive bar: full-width background — should continue to the right of the modal."
	return lipgloss.NewStyle().
		Background(lipgloss.Color("#3d4f2f")).
		Foreground(lipgloss.Color("#e8f5e0")).
		Width(viewWidth).
		Render(msg)
}

func (m model) View() string {
	w, h := m.width, m.height
	if w == 0 || h == 0 {
		w, h = 80, 24
	}

	dim := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	var b strings.Builder
	// Row 0: colored graph — modal uses top=0 so the box cuts through this line.
	b.WriteString(graphRow())
	b.WriteByte('\n')
	// Row 1: full-width background; modal body rows overlap this line too.
	b.WriteString(hintRow(w))
	b.WriteByte('\n')
	b.WriteString(dim.Render("Changed files"))
	b.WriteByte('\n')
	b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Render("M ") + "src/feature.go")
	b.WriteByte('\n')
	b.WriteByte('\n')
	b.WriteString(dim.Render("space: toggle · q: quit"))

	mainView := b.String()

	if !m.showModal {
		return mainView
	}

	left := modalLeft
	if left+modalOverlayW > w {
		left = max(0, w-modalOverlayW)
	}

	title := lipgloss.NewStyle().Foreground(lipgloss.Color("117")).Bold(true).Render("Evolog split (demo)")
	body := lipgloss.JoinVertical(lipgloss.Left,
		title,
		"",
		dim.Render("Cyan graph tail + olive bar should both resume past the box."),
		"",
		lipgloss.NewStyle().Foreground(lipgloss.Color("141")).Render("eb544e6d (empty)"),
		"",
		lipgloss.NewStyle().Reverse(true).Render(" Split here (Enter) ")+"  "+
			lipgloss.NewStyle().Background(lipgloss.Color("236")).Foreground(lipgloss.Color("252")).Render(" Cancel (Esc) "),
		dim.Render("j/k: move · Do not pick tip as base"),
	)

	modal := lipgloss.NewStyle().
		Width(modalOverlayW).
		Height(modalOverlayH).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(0, 1).
		Render(body)

	// top=0: first row of the modal stack aligns with graphRow(); left slices through purple + cyan.
	return overlay.OverlayView(mainView, modal, w, h, 0, left)
}

func main() {
	p := tea.NewProgram(model{}, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
