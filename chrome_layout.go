package overlay

import (
	tea "github.com/charmbracelet/bubbletea"
)

// ChromeLayout holds compositor origin, modal dimensions, content size, and hit regions
// for one paint frame. Pass a cached layout to HandleChromePointerLayout and
// HandleChromeKeyLayout to avoid re-measuring the modal string on every input event.
type ChromeLayout struct {
	Top, Left          int
	ModalW, ModalH     int
	ContentW, ContentH int
	Regions            ChromeRegions
}

// LayoutChrome computes placement and hit regions from a rendered modal string.
func LayoutChrome(cfg OverlayConfig, st *LayerState, modal string, viewW, viewH int) ChromeLayout {
	top, left := EntryClampedOrigin(cfg, st, modal, viewW, viewH)
	mw, mh := ModalCellSize(modal)
	wc := cfg.WindowChrome.effective()
	cw, ch := ContentSizeForLayer(modal, wc, st)
	return ChromeLayout{
		Top:      top,
		Left:     left,
		ModalW:   mw,
		ModalH:   mh,
		ContentW: cw,
		ContentH: ch,
		Regions:  ComputeChromeRegions(wc, cw, ch),
	}
}

// CellInModal reports whether terminal cell (x, y) lies inside the modal rectangle.
func (l ChromeLayout) CellInModal(x, y int) bool {
	return CellInModal(x, y, l.Top, l.Left, l.ModalW, l.ModalH)
}

// HandleChromeMouseLayout is like HandleChromeMouse but uses a precomputed ChromeLayout.
func HandleChromeMouseLayout(msg tea.MouseMsg, cfg OverlayConfig, st *LayerState, layout ChromeLayout, viewW, viewH int) ChromeMouseResult {
	var action ChromePointerAction
	switch msg.Action {
	case tea.MouseActionPress:
		action = ChromePointerPress
	case tea.MouseActionRelease:
		action = ChromePointerRelease
	case tea.MouseActionMotion:
		action = ChromePointerMotion
	default:
		return ChromeMouseResult{}
	}
	leftBtn := msg.Button == tea.MouseButtonLeft
	return HandleChromePointerLayout(action, leftBtn, msg.X, msg.Y, cfg, st, layout, viewW, viewH)
}

// HandleChromePointerLayout is the layout-cached chrome pointer handler.
func HandleChromePointerLayout(action ChromePointerAction, leftButton bool, x, y int, cfg OverlayConfig, st *LayerState, layout ChromeLayout, viewW, viewH int) ChromeMouseResult {
	wc := cfg.WindowChrome.effective()
	if !wc.Enabled || st == nil {
		return ChromeMouseResult{}
	}
	rx, ry := x-layout.Left, y-layout.Top
	return handleChromePointerCore(wc, st, action, leftButton, rx, ry, layout, viewW, viewH)
}

// HandleChromeKeyLayout is like HandleChromeKey but uses a precomputed ChromeLayout.
func HandleChromeKeyLayout(msg tea.KeyMsg, cfg OverlayConfig, st *LayerState, layout ChromeLayout, viewW, viewH int) ChromeKeyResult {
	return HandleChromeKeyStringLayout(msg.String(), cfg, st, layout, viewW, viewH)
}

// HandleChromeKeyStringLayout is the layout-cached key handler (Bubble Tea v2 KeyPressMsg).
func HandleChromeKeyStringLayout(key string, cfg OverlayConfig, st *LayerState, layout ChromeLayout, viewW, viewH int) ChromeKeyResult {
	wc := cfg.WindowChrome.effective()
	if !wc.keyboard() || st == nil {
		return ChromeKeyResult{}
	}
	step := wc.keyStep()
	switch classifyChromeKey(key) {
	case chromeKeyMoveRight:
		if wc.draggable() {
			nudgeOrigin(st, 0, step, layout.ModalW, layout.ModalH, viewW, viewH)
			return ChromeKeyResult{Consumed: true}
		}
	case chromeKeyMoveLeft:
		if wc.draggable() {
			nudgeOrigin(st, 0, -step, layout.ModalW, layout.ModalH, viewW, viewH)
			return ChromeKeyResult{Consumed: true}
		}
	case chromeKeyMoveDown:
		if wc.draggable() {
			nudgeOrigin(st, step, 0, layout.ModalW, layout.ModalH, viewW, viewH)
			return ChromeKeyResult{Consumed: true}
		}
	case chromeKeyMoveUp:
		if wc.draggable() {
			nudgeOrigin(st, -step, 0, layout.ModalW, layout.ModalH, viewW, viewH)
			return ChromeKeyResult{Consumed: true}
		}
	case chromeKeyGrowRight:
		if wc.resizable() {
			nudgeContentSize(st, wc, step, 0, layout.Top, layout.Left, viewW, viewH)
			return ChromeKeyResult{Consumed: true}
		}
	case chromeKeyGrowLeft:
		if wc.resizable() {
			nudgeContentSize(st, wc, -step, 0, layout.Top, layout.Left, viewW, viewH)
			return ChromeKeyResult{Consumed: true}
		}
	case chromeKeyGrowDown:
		if wc.resizable() {
			nudgeContentSize(st, wc, 0, step, layout.Top, layout.Left, viewW, viewH)
			return ChromeKeyResult{Consumed: true}
		}
	case chromeKeyGrowUp:
		if wc.resizable() {
			nudgeContentSize(st, wc, 0, -step, layout.Top, layout.Left, viewW, viewH)
			return ChromeKeyResult{Consumed: true}
		}
	}
	return ChromeKeyResult{}
}

func handleChromePointerCore(wc WindowChrome, st *LayerState, action ChromePointerAction, leftButton bool, rx, ry int, layout ChromeLayout, viewW, viewH int) ChromeMouseResult {
	reg := layout.Regions
	top, left, mw, mh := layout.Top, layout.Left, layout.ModalW, layout.ModalH

	switch action {
	case ChromePointerPress:
		if !leftButton {
			return ChromeMouseResult{}
		}
		if wc.showClose() && cellInRect(rx, ry, reg.CloseX, reg.CloseY, reg.CloseW, reg.CloseH) {
			return ChromeMouseResult{Consumed: true, Pop: true}
		}
		if wc.resizable() {
			if cellInRect(rx, ry, reg.ResizeCornerX, reg.ResizeCornerY, reg.ResizeCornerW, reg.ResizeCornerH) {
				st.Resizing = true
				st.ResizeEdge = ResizeCorner
				st.ResizeStartX, st.ResizeStartY = rx+left, ry+top
				st.ResizeStartW, st.ResizeStartH = layout.ContentW, layout.ContentH
				return ChromeMouseResult{Consumed: true}
			}
			if cellInRect(rx, ry, reg.ResizeRightX, reg.ResizeRightY, reg.ResizeRightW, reg.ResizeRightH) {
				st.Resizing = true
				st.ResizeEdge = ResizeRight
				st.ResizeStartX, st.ResizeStartY = rx+left, ry+top
				st.ResizeStartW, st.ResizeStartH = layout.ContentW, layout.ContentH
				return ChromeMouseResult{Consumed: true}
			}
			if cellInRect(rx, ry, reg.ResizeBottomX, reg.ResizeBottomY, reg.ResizeBottomW, reg.ResizeBottomH) {
				st.Resizing = true
				st.ResizeEdge = ResizeBottom
				st.ResizeStartX, st.ResizeStartY = rx+left, ry+top
				st.ResizeStartW, st.ResizeStartH = layout.ContentW, layout.ContentH
				return ChromeMouseResult{Consumed: true}
			}
		}
		if wc.draggable() && cellInTabDrag(rx, ry, reg) {
			st.Dragging = true
			st.DragOffsetX = rx
			st.DragOffsetY = ry
			return ChromeMouseResult{Consumed: true}
		}
	case ChromePointerMotion:
		if st.Resizing && wc.resizable() {
			applyResize(st, wc, rx+left, ry+top, left, top, viewW, viewH)
			return ChromeMouseResult{Consumed: true}
		}
		if st.Dragging && wc.draggable() {
			st.OriginTop = ry + top - st.DragOffsetY
			st.OriginLeft = rx + left - st.DragOffsetX
			st.OriginTop, st.OriginLeft = ClampOverlayOrigin(mw, mh, viewW, viewH, st.OriginTop, st.OriginLeft)
			return ChromeMouseResult{Consumed: true}
		}
	case ChromePointerRelease:
		if leftButton && st.Resizing {
			st.Resizing = false
			st.ResizeEdge = ResizeNone
			return ChromeMouseResult{Consumed: true}
		}
		if leftButton && st.Dragging {
			st.Dragging = false
			return ChromeMouseResult{Consumed: true}
		}
	}
	return ChromeMouseResult{}
}
