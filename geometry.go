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

// CellInTitleBar reports whether terminal cell coordinates (x, y) fall on the title bar
// at the top of a modal, excluding the close-button strip on the right when closeW > 0.
// Use the same top, left, mw, and titleBarH values as compositing and chrome hit-testing
// (for example from ComputeChromeRegions or WindowChrome titleBarHeight) so results match
// painted geometry.
//
// Bubble Tea v1 tea.MouseMsg uses zero-based X and Y for the terminal cell (column, row),
// matching top and left passed into OverlayView.
//
// @param x zero-based column of the cell to test.
// @param y zero-based row of the cell to test.
// @param top compositor row of the modal's top-left corner.
// @param left compositor column of the modal's top-left corner.
// @param mw modal width in cells (same as ModalCellSize width).
// @param titleBarH number of title-bar rows from top (0 means no title bar).
// @param closeW width in cells of the close control on the right (0 if absent).
//
// @return true when (x, y) lies in [left, left+mw) × [top, top+titleBarH) and not in the close region.
func CellInTitleBar(x, y, top, left, mw, titleBarH, closeW int) bool {
	return layout.CellInTitleBar(x, y, top, left, mw, titleBarH, closeW)
}

// CellInCloseButton reports whether terminal cell coordinates (x, y) fall on the close
// control in the title bar's rightmost closeW columns. Use the same placement and chrome
// dimensions as CellInTitleBar and CellInModal so click handling matches OverlayStack.
//
// @param x zero-based column of the cell to test.
// @param y zero-based row of the cell to test.
// @param top compositor row of the modal's top-left corner.
// @param left compositor column of the modal's top-left corner.
// @param mw modal width in cells (same as ModalCellSize width).
// @param titleBarH number of title-bar rows from top (must be > 0 for a hit).
// @param closeW width in cells of the close control (must be > 0 for a hit).
//
// @return true when (x, y) lies in the close region [left+mw-closeW, left+mw) × [top, top+titleBarH).
func CellInCloseButton(x, y, top, left, mw, titleBarH, closeW int) bool {
	return layout.CellInCloseButton(x, y, top, left, mw, titleBarH, closeW)
}
