package overlay

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var dimStyle = lipgloss.NewStyle().Faint(true)

func DimSurface(s string, opacity float64) string {
	return dimSurface(s, opacity)
}

func dimSurface(s string, opacity float64) string {
	if opacity <= 0 || s == "" {
		return s
	}
	lines := strings.Split(s, "\n")
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		out = append(out, dimStyle.Render(line))
	}
	return strings.Join(out, "\n")
}
