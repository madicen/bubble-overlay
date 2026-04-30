package layout

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

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

func CellInModal(x, y, top, left, mw, mh int) bool {
	if mw <= 0 || mh <= 0 {
		return true
	}
	return x >= left && x < left+mw && y >= top && y < top+mh
}
