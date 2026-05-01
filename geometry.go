package overlay

import "github.com/madicen/bubble-overlay/internal/layout"

// ClampOverlayOriginAtPoint runs ModalCellSize(modal) then ClampOverlayOrigin. Use for anchors such as a
// context menu at (top, left) (e.g. Bubble Tea mouse Y/X as row/column) so hit-testing with CellInModal
// matches OverlayView(modal, …, top, left).
func ClampOverlayOriginAtPoint(modal string, viewW, viewH, top, left int) (int, int) {
	mw, mh := ModalCellSize(modal)
	return ClampOverlayOrigin(mw, mh, viewW, viewH, top, left)
}

// ClampMenuOrigin is an alias for ClampOverlayOriginAtPoint: same implementation, for call sites that anchor
// a menu or popover at a cursor cell (typically tea.MouseMsg.Y as top, tea.MouseMsg.X as left).
func ClampMenuOrigin(modal string, viewW, viewH, anchorTop, anchorLeft int) (int, int) {
	return ClampOverlayOriginAtPoint(modal, viewW, viewH, anchorTop, anchorLeft)
}

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
