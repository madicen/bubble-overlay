// Package layout holds cell-layout helpers shared by the root overlay stack and v2.
package layout

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// ModalCellSize returns display width and line count for a multiline modal string.
func ModalCellSize(modal string) (w, h int) {
	modal = strings.TrimSuffix(modal, "\n")
	lines := strings.Split(modal, "\n")
	h = len(lines)
	if h == 0 {
		h = 1
	}
	for _, line := range lines {
		if lw := lipgloss.Width(line); lw > w {
			w = lw
		}
	}
	if w == 0 {
		w = 1
	}
	return w, h
}

// CellInModal reports whether cell (x,y) lies inside the modal rectangle.
func CellInModal(x, y, top, left, mw, mh int) bool {
	if mw <= 0 || mh <= 0 {
		return true
	}
	return x >= left && x < left+mw && y >= top && y < top+mh
}
