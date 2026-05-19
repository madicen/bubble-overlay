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

// CellInTitleBar reports whether (x, y) is in the title bar rows at the top of the modal,
// excluding the close-button region on the right when closeW > 0.
func CellInTitleBar(x, y, top, left, mw, titleBarH, closeW int) bool {
	if titleBarH <= 0 || mw <= 0 {
		return false
	}
	if y < top || y >= top+titleBarH {
		return false
	}
	if x < left || x >= left+mw {
		return false
	}
	if closeW > 0 && x >= left+mw-closeW {
		return false
	}
	return true
}

// CellInCloseButton reports whether (x, y) is in the close control on the title bar's right edge.
func CellInCloseButton(x, y, top, left, mw, titleBarH, closeW int) bool {
	if closeW <= 0 || titleBarH <= 0 || mw <= 0 {
		return false
	}
	if y < top || y >= top+titleBarH {
		return false
	}
	return x >= left+mw-closeW && x < left+mw
}
