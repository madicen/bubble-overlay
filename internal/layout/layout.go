package layout

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// ModalCellSize matches overlayViewInternal: strings.Split on "\n" (no TrimSuffix),
// height = line count, width = max lipgloss.Width per line.
func ModalCellSize(modal string) (w, h int) {
	lines := strings.Split(modal, "\n")
	if len(lines) == 0 {
		return 0, 0
	}
	h = len(lines)
	for _, line := range lines {
		if lw := lipgloss.Width(line); lw > w {
			w = lw
		}
	}
	return w, h
}

func CellInModal(x, y, top, left, mw, mh int) bool {
	if mw <= 0 || mh <= 0 {
		return true
	}
	return x >= left && x < left+mw && y >= top && y < top+mh
}
