package overlay

import "github.com/madicen/bubble-overlay/internal/layout"

// ModalCellSize returns display-cell width (max per line) and line count for a modal string,
// using the same rules as OverlayView when measuring the modal for placement and clamping.
func ModalCellSize(modal string) (w, h int) {
	return layout.ModalCellSize(modal)
}

// CellInModal reports whether terminal cell coordinates (x, y) fall inside the modal rectangle
// [left, left+mw) × [top, top+mh). Use coordinates after ClampOverlayOrigin / Placement.ClampedOrigin
// so hit-testing matches painted overlay geometry.
//
// Bubble Tea v1 tea.MouseMsg uses zero-based X and Y for the terminal cell (column, row), matching
// top and left passed into OverlayView (also zero-based from the top-left of the view).
func CellInModal(x, y, top, left, mw, mh int) bool {
	return layout.CellInModal(x, y, top, left, mw, mh)
}
