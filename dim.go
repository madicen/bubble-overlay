package overlay

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var dimStyle = lipgloss.NewStyle().Faint(true)

// DimSurface applies a terminal faint style to a full multiline view when opacity > 0.
// Values <= 0 leave s unchanged.
func DimSurface(s string, opacity float64) string {
	return dimSurface(s, opacity)
}

// dimSurface applies a terminal faint style to a full multiline view when opacity > 0.
// Values <= 0 leave s unchanged. True alpha blending is not available in plain TUI;
// higher DimOpacity values still use one faint pass (reserved for stronger dim later).
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
