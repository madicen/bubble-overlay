package overlay

import (
	tea "github.com/charmbracelet/bubbletea"
)

// WindowResizedMsg is sent when a resizable window finishes a resize (mouse release or keyboard grow).
type WindowResizedMsg struct {
	NewContentWidth  int
	NewContentHeight int
}

// Window composes one chromed modal per frame for state-machine-driven apps (no OverlayStack).
// Pass a stable non-empty key per modal slot; a key change from the previous frame resets
// State so reopens recenter. An empty key clears State when no modal is open.
type Window struct {
	State LayerState

	// OnResize is optional; WindowResizedMsg is always emitted on resize end.
	OnResize func(contentW, contentH int) tea.Cmd

	lastKey string

	frameKey      string
	frameTitle    string
	frameContent  string
	frameViewW    int
	frameViewH    int
	frameModal    string
	frameLayout   ChromeLayout
	frameCfg      OverlayConfig
	frameValid    bool
}

// View wraps content in window chrome and composites it over base.
// No-op when key is empty, content is empty, or the viewport is zero-sized.
func (w *Window) View(base, content, title, key string, viewW, viewH int) string {
	if w == nil {
		return base
	}
	if !w.active(key, content, viewW, viewH) {
		w.syncKey("")
		return base
	}
	w.syncKey(key)
	cfg, _, layout, ok := w.ensureFrame(title, content, key, viewW, viewH)
	if !ok {
		return base
	}
	cur := base
	if cfg.DimOpacity > 0 {
		cur = DimSurface(cur, cfg.DimOpacity)
	}
	return ComposeModalLayer(cur, w.frameModal, cfg, layout.Top, layout.Left, viewW, viewH)
}

// Update routes input for the active chromed modal. closeCmd is returned when the user
// clicks [x] or when CloseOnEscape is enabled and Esc is pressed.
func (w *Window) Update(msg tea.Msg, content, title, key string, viewW, viewH int, closeCmd tea.Cmd) (consumed bool, cmd tea.Cmd) {
	if w == nil {
		return false, nil
	}
	if !w.active(key, content, viewW, viewH) {
		w.syncKey("")
		return false, nil
	}
	w.syncKey(key)
	cfg, st, layout, ok := w.ensureFrame(title, content, key, viewW, viewH)
	if !ok || st == nil {
		return false, nil
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		viewW, viewH = msg.Width, msg.Height
		w.frameViewW, w.frameViewH = viewW, viewH
		ReclampLayerOrigin(cfg, st, w.frameModal, viewW, viewH)
		w.invalidateFrame()
		return false, nil

	case tea.KeyMsg:
		if cfg.CloseOnEscape && isEscapeKey(msg) {
			w.syncKey("")
			return true, closeCmd
		}
		oldW, oldH := st.ContentWidth, st.ContentHeight
		if cfg.WindowChrome.effective().keyboard() {
			if res := HandleChromeKeyLayout(msg, cfg, st, layout, viewW, viewH); res.Consumed {
				w.invalidateFrame()
				return true, w.maybeResizeCmd(oldW, oldH)
			}
		}
		return false, nil

	case tea.MouseMsg:
		wasResizing := st.Resizing
		oldW, oldH := st.ContentWidth, st.ContentHeight
		if cfg.WindowChrome.Enabled {
			res := HandleChromeMouseLayout(msg, cfg, st, layout, viewW, viewH)
			if res.Pop {
				w.syncKey("")
				return true, closeCmd
			}
			if res.Consumed {
				w.invalidateFrame()
				var resizeCmd tea.Cmd
				if wasResizing && !st.Resizing {
					resizeCmd = w.maybeResizeCmd(oldW, oldH)
				}
				return true, resizeCmd
			}
		}
		if cfg.CloseOnClickOutside && isPrimaryPress(msg) {
			if !layout.CellInModal(msg.X, msg.Y) {
				w.syncKey("")
				return true, closeCmd
			}
		}
		return false, nil

	default:
		return false, nil
	}
}

// LastChromeLayout returns layout from the most recent View or Update on this frame.
func (w *Window) LastChromeLayout() (ChromeLayout, bool) {
	if w == nil || !w.frameValid {
		return ChromeLayout{}, false
	}
	return w.frameLayout, true
}

func (w *Window) active(key, content string, viewW, viewH int) bool {
	return key != "" && content != "" && viewW > 0 && viewH > 0
}

func (w *Window) syncKey(key string) {
	if key == "" {
		w.State.Reset()
		w.lastKey = ""
		w.invalidateFrame()
		return
	}
	if key != w.lastKey {
		w.State.Reset()
		w.lastKey = key
		w.invalidateFrame()
	}
}

func (w *Window) ensureFrame(title, content, key string, viewW, viewH int) (OverlayConfig, *LayerState, ChromeLayout, bool) {
	if w.frameValid &&
		w.frameKey == key &&
		w.frameTitle == title &&
		w.frameContent == content &&
		w.frameViewW == viewW &&
		w.frameViewH == viewH {
		return w.frameCfg, &w.State, w.frameLayout, true
	}
	cfg := w.configFor(title)
	if !w.State.ContentSizeInitialized && cfg.WindowChrome.resizable() {
		InitLayerContentSize(&w.State, content, cfg.WindowChrome)
	}
	w.frameModal = RenderEntryModal(content, cfg, &w.State)
	layout := LayoutChrome(cfg, &w.State, w.frameModal, viewW, viewH)
	w.frameKey = key
	w.frameTitle = title
	w.frameContent = content
	w.frameViewW = viewW
	w.frameViewH = viewH
	w.frameLayout = layout
	w.frameCfg = cfg
	w.frameValid = true
	return cfg, &w.State, layout, true
}

func (w *Window) configFor(title string) OverlayConfig {
	cfg := DefaultOverlayConfig()
	cfg.DimOpacity = 0
	cfg.CloseOnEscape = false
	cfg.CloseOnClickOutside = false
	cfg.WindowChrome = EnableWindowChrome(title)
	cfg.WindowChrome.Resizable = true
	cfg.WindowChrome.Keyboard = true
	cfg.WindowChrome.CenterContent = true
	cfg.WindowChrome.MinWidth = 32
	cfg.WindowChrome.MinHeight = 6
	cfg.WindowChrome = cfg.WindowChrome.effective()
	return cfg
}

func (w *Window) invalidateFrame() {
	w.frameValid = false
}

func (w *Window) maybeResizeCmd(oldW, oldH int) tea.Cmd {
	if !w.State.ContentSizeInitialized {
		return nil
	}
	cw, ch := w.State.ContentWidth, w.State.ContentHeight
	if cw == oldW && ch == oldH {
		return nil
	}
	var cmds []tea.Cmd
	cmds = append(cmds, func() tea.Msg {
		return WindowResizedMsg{NewContentWidth: cw, NewContentHeight: ch}
	})
	if w.OnResize != nil {
		cmds = append(cmds, w.OnResize(cw, ch))
	}
	return tea.Batch(cmds...)
}
