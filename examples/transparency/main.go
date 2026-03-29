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
	showModal     bool
	width         int
	height        int
	selectedIndex int
}

type keyMap struct {
	Quit key.Binding
	Show key.Binding
	Up   key.Binding
	Down key.Binding
}

var keys = keyMap{
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
	Show: key.NewBinding(
		key.WithKeys(" "),
		key.WithHelp("space", "view details"),
	),
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "prev"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "next"),
	),
}

type dataItem struct {
	id    int
	name  string
	value int
	desc  string
}

var items = []dataItem{
	{1, "Alice", 42, "Senior Engineer"},
	{2, "Bob", 38, "Product Manager"},
	{3, "Carol", 35, "Designer"},
	{4, "David", 41, "Lead Architect"},
	{5, "Eve", 39, "DevOps Engineer"},
	{6, "Frank", 37, "QA Lead"},
	{7, "Grace", 40, "Developer"},
	{8, "Heidi", 33, "Security Specialist"},
	{9, "Ivan", 45, "Systems Architect"},
	{10, "Judy", 29, "Junior Developer"},
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
		case key.Matches(msg, keys.Up):
			if m.selectedIndex > 0 {
				m.selectedIndex--
			}
			return m, nil
		case key.Matches(msg, keys.Down):
			if m.selectedIndex < len(items)-1 {
				m.selectedIndex++
			}
			return m, nil
		}
	}
	return m, nil
}

func (m model) View() string {
	w, h := m.width, m.height
	if w == 0 || h == 0 {
		w, h = 80, 24
	}

	// Create a table in the main view showing all items
	mainView := m.createMainView(w, h)

	if !m.showModal {
		return mainView
	}

	// Create a detailed view modal for the selected item
	modal := m.createDetailModal()

	// Calculate 'top' so the mask line (first line of modal content)
	// aligns with the selected row in the table.
	// Header is index 0, Divider is index 1. Data starts at 2.
	// The modal border adds 1 row, so top = index + 1 makes the content hit index + 2.
	top := m.selectedIndex + 1
	left := 0

	// Use OverlayViewWithMask for pixel-perfect placement.
	return overlay.OverlayViewWithMask(mainView, modal, w, h, top, left, '░')
}

func (m model) createMainView(w, h int) string {
	var b strings.Builder

	// Header row - Always visible and spans full width to anchor the UI
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("15")).
		Background(lipgloss.Color("4")).
		Padding(0, 1).
		Width(33)
	header := headerStyle.Render(fmt.Sprintf("%-5s %-25s", "ID", "Name"))
	b.WriteString(header)
	b.WriteString("\n")

	// Full-width divider for visual consistency
	dividerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	b.WriteString(dividerStyle.Render(strings.Repeat("─", 32)))
	b.WriteString("\n")

	// Data rows
	rowStyle := lipgloss.NewStyle().Padding(0, 1)
	selectedRowStyle := rowStyle.
		Foreground(lipgloss.Color("15")).
		Background(lipgloss.Color("62"))

	for i, item := range items {
		row := fmt.Sprintf("%-5d %-25s", item.id, item.name)
		if i == m.selectedIndex {
			b.WriteString(selectedRowStyle.Render(row))
		} else {
			b.WriteString(rowStyle.Render(row))
		}
		b.WriteString("\n")
	}

	// Padding
	for i := len(items); i < h-3; i++ {
		b.WriteString("\n")
	}

	// Footer with help
	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")).
		Padding(0, 1)
	help := helpStyle.Render("↑↓: navigate  SPACE: details  Q: quit")
	b.WriteString(help)

	return lipgloss.NewStyle().MaxHeight(h).MaxWidth(w).Render(b.String())
}

func (m model) createDetailModal() string {
	item := items[m.selectedIndex]

	// Using '░' as our "green screen" character.
	mask := "░"

	// We've removed the explicit background color. Since the library is
	// configured to only treat the mask rune as transparent, the rest of
	// the modal (including spaces) will remain opaque.
	boxStyle := lipgloss.NewStyle().
		Width(32).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("99")).
		Padding(0, 0)

	// Indent the details manually since padding is 0 for the mask alignment.
	content := fmt.Sprintf(
		"%s\n\n  User Details\n  ────────────\n  ID:   %d\n  Name: %s\n  Role: %s\n  Age:  %d",
		strings.Repeat(mask, 30), item.id, item.name, item.desc, item.value,
	)

	return boxStyle.Render(content)
}

func main() {
	p := tea.NewProgram(model{
		selectedIndex: 0,
	}, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
