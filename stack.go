package overlay

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/madicen/bubble-overlay/internal/layout"
)

type OverlayOnCloser interface {
	OnOverlayClose() tea.Cmd
}

// OverlayTitler lets a stack-pushed model (or a Window content provider)
// supply a dynamic tab title. When a model implements this interface, the
// renderer reads OverlayTitle() each frame and uses that string in the
// WindowChrome tab instead of the static cfg.WindowChrome.Title. Returning
// "" falls back to the static title so the interface is safe to opt into
// even when you only sometimes have a dynamic title.
type OverlayTitler interface {
	OverlayTitle() string
}

// OverlayResizedMsg is dispatched to a stack-pushed model when the user
// finishes a resize gesture through window chrome (mouse release on a
// resize edge, or an Alt+Shift arrow keypress). Models can react by
// adjusting viewports, re-wrapping content, etc.
//
// NewContentWidth / NewContentHeight are the body dims the chrome will
// render at; they exclude the chrome's own border and tab.
type OverlayResizedMsg struct {
	NewContentWidth  int
	NewContentHeight int
}

// OverlayResizer is the optional interface a stack-pushed model can
// implement if it prefers a direct callback over a OverlayResizedMsg in
// its Update. The stack invokes OnOverlayResize before delivering the
// OverlayResizedMsg; returning a tea.Cmd schedules follow-up work.
type OverlayResizer interface {
	OnOverlayResize(contentW, contentH int) tea.Cmd
}

type stackEntry struct {
	model tea.Model
	cfg   OverlayConfig
	layer LayerState
}

type OverlayStack struct {
	entries      []stackEntry
	lastW, lastH int
}

func (s *OverlayStack) Depth() int {
	if s == nil {
		return 0
	}
	return len(s.entries)
}

func (s *OverlayStack) StackDepth() int {
	return s.Depth()
}

func (s *OverlayStack) Push(m tea.Model, cfg OverlayConfig) tea.Cmd {
	if s == nil {
		return nil
	}
	if cfg.WindowChrome.Enabled {
		cfg.WindowChrome = cfg.WindowChrome.effective()
	}
	ent := stackEntry{model: m, cfg: cfg}
	InitLayerContentSize(&ent.layer, m.View(), cfg.WindowChrome.effective())
	s.entries = append(s.entries, ent)
	return m.Init()
}

func (s *OverlayStack) Pop() (popped tea.Model, cmd tea.Cmd) {
	if s == nil || len(s.entries) == 0 {
		return nil, nil
	}
	i := len(s.entries) - 1
	ent := s.entries[i]
	s.entries = s.entries[:i]
	if c, ok := ent.model.(OverlayOnCloser); ok {
		cmd = c.OnOverlayClose()
	}
	return ent.model, cmd
}

func (s *OverlayStack) Top() tea.Model {
	if s == nil || len(s.entries) == 0 {
		return nil
	}
	return s.entries[len(s.entries)-1].model
}

// TopChromeLayout returns the compositor origin and chrome hit regions for the top stack entry.
func (s *OverlayStack) TopChromeLayout(viewW, viewH int) (top, left int, reg ChromeRegions, ok bool) {
	if s == nil || len(s.entries) == 0 {
		return 0, 0, ChromeRegions{}, false
	}
	ent := &s.entries[len(s.entries)-1]
	cfg := ent.effectiveConfig()
	wc := cfg.WindowChrome.effective()
	modal := RenderEntryModal(ent.model.View(), cfg, &ent.layer)
	top, left = EntryClampedOrigin(cfg, &ent.layer, modal, viewW, viewH)
	cw, ch := ContentSizeForLayer(modal, wc, &ent.layer)
	reg = ComputeChromeRegions(wc, cw, ch)
	return top, left, reg, true
}

// effectiveConfig returns the entry's OverlayConfig with WindowChrome.Title
// overridden by the model's dynamic OverlayTitle when it implements the
// OverlayTitler interface and returns a non-empty string. The returned
// config is a copy so callers can't accidentally clobber the entry's stored
// configuration.
func (e *stackEntry) effectiveConfig() OverlayConfig {
	cfg := e.cfg
	if !cfg.WindowChrome.Enabled {
		return cfg
	}
	if t, ok := e.model.(OverlayTitler); ok {
		if dyn := t.OverlayTitle(); dyn != "" {
			cfg.WindowChrome.Title = dyn
		}
	}
	return cfg
}

// SetTopLayerOrigin sets the painted origin for the top entry when it uses draggable window chrome.
func (s *OverlayStack) SetTopLayerOrigin(top, left int) {
	if s == nil || len(s.entries) == 0 {
		return
	}
	ent := &s.entries[len(s.entries)-1]
	ent.layer.OriginTop = top
	ent.layer.OriginLeft = left
	ent.layer.OriginInitialized = true
}

func (s *OverlayStack) MainReceivesKeyMsg() bool {
	return s == nil || len(s.entries) == 0
}

func (s *OverlayStack) MainReceivesMouseMsg() bool {
	return s.MainReceivesKeyMsg()
}

func (s *OverlayStack) syncViewport(viewW, viewH int) {
	if s == nil || viewW <= 0 || viewH <= 0 {
		return
	}
	s.lastW, s.lastH = viewW, viewH
}

func (s *OverlayStack) View(baseMain string, viewW, viewH int) string {
	if s == nil || len(s.entries) == 0 {
		return baseMain
	}
	s.syncViewport(viewW, viewH)
	cur := baseMain
	for i := range s.entries {
		cur = s.composeEntry(cur, &s.entries[i], viewW, viewH)
	}
	return cur
}

func (s *OverlayStack) composeEntry(cur string, ent *stackEntry, viewW, viewH int) string {
	cfg := ent.effectiveConfig()
	if cfg.DimOpacity > 0 {
		cur = DimSurface(cur, cfg.DimOpacity)
	}
	modal := RenderEntryModal(ent.model.View(), cfg, &ent.layer)
	top, left := EntryClampedOrigin(cfg, &ent.layer, modal, viewW, viewH)
	return ComposeModalLayer(cur, modal, cfg, top, left, viewW, viewH)
}

func (s *OverlayStack) topLayout(viewW, viewH int) (top, left, mw, mh int) {
	if s == nil || len(s.entries) == 0 {
		return 0, 0, 0, 0
	}
	ent := &s.entries[len(s.entries)-1]
	cfg := ent.effectiveConfig()
	modal := RenderEntryModal(ent.model.View(), cfg, &ent.layer)
	mw, mh = layout.ModalCellSize(modal)
	top, left = EntryClampedOrigin(cfg, &ent.layer, modal, viewW, viewH)
	return top, left, mw, mh
}

// notifyResize calls the entry model's optional OnOverlayResize hook and
// returns a tea.Cmd that delivers OverlayResizedMsg back through the
// model's own Update. Both signals fire so consumers can pick whichever
// integration style suits them.
func (s *OverlayStack) notifyResize(ent *stackEntry, cw, ch int) tea.Cmd {
	var cmds []tea.Cmd
	if r, ok := ent.model.(OverlayResizer); ok {
		if c := r.OnOverlayResize(cw, ch); c != nil {
			cmds = append(cmds, c)
		}
	}
	msg := OverlayResizedMsg{NewContentWidth: cw, NewContentHeight: ch}
	// Deliver the message through the model's Update so the stack-pushed
	// model can react with the same code path it uses for any other tea
	// message. We invoke Update synchronously here (rather than batching
	// it through tea.Cmd) so the model sees the resize before any later
	// frame renders, which keeps subsequent View() calls in sync with the
	// new content dims.
	var c tea.Cmd
	ent.model, c = ent.model.Update(msg)
	if c != nil {
		cmds = append(cmds, c)
	}
	if len(cmds) == 0 {
		return nil
	}
	return tea.Batch(cmds...)
}

func (s *OverlayStack) Update(msg tea.Msg) tea.Cmd {
	if s == nil || len(s.entries) == 0 {
		return nil
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		s.lastW, s.lastH = msg.Width, msg.Height
		var cmds []tea.Cmd
		for i := range s.entries {
			cfg := s.entries[i].effectiveConfig()
			modal := RenderEntryModal(s.entries[i].model.View(), cfg, &s.entries[i].layer)
			ReclampLayerOrigin(cfg, &s.entries[i].layer, modal, msg.Width, msg.Height)
			var c tea.Cmd
			s.entries[i].model, c = s.entries[i].model.Update(msg)
			cmds = append(cmds, c)
		}
		return tea.Batch(cmds...)

	case tea.KeyMsg:
		top := len(s.entries) - 1
		ent := &s.entries[top]
		cfg := ent.effectiveConfig()
		if cfg.CloseOnEscape && isEscapeKey(msg) {
			_, c := s.Pop()
			return c
		}
		viewW, viewH := s.lastW, s.lastH
		if viewW <= 0 {
			viewW = 80
		}
		if viewH <= 0 {
			viewH = 25
		}
		if cfg.WindowChrome.effective().keyboard() {
			modal := RenderEntryModal(ent.model.View(), cfg, &ent.layer)
			t, l, mw, mh := s.topLayout(viewW, viewH)
			oldW, oldH := ent.layer.ContentWidth, ent.layer.ContentHeight
			if res := HandleChromeKey(msg, cfg, &ent.layer, modal, t, l, mw, mh, viewW, viewH); res.Consumed {
				if ent.layer.ContentSizeInitialized &&
					(ent.layer.ContentWidth != oldW || ent.layer.ContentHeight != oldH) {
					return s.notifyResize(ent, ent.layer.ContentWidth, ent.layer.ContentHeight)
				}
				return nil
			}
		}
		var c tea.Cmd
		ent.model, c = ent.model.Update(msg)
		return c

	case tea.MouseMsg:
		top := len(s.entries) - 1
		ent := &s.entries[top]
		cfg := ent.effectiveConfig()
		viewW, viewH := s.lastW, s.lastH
		if viewW <= 0 {
			viewW = 80
		}
		if viewH <= 0 {
			viewH = 25
		}
		t, l, mw, mh := s.topLayout(viewW, viewH)
		if cfg.WindowChrome.Enabled {
			modal := RenderEntryModal(ent.model.View(), cfg, &ent.layer)
			wasResizing := ent.layer.Resizing
			oldW, oldH := ent.layer.ContentWidth, ent.layer.ContentHeight
			res := HandleChromeMouse(msg, cfg, &ent.layer, modal, t, l, mw, mh, viewW, viewH)
			if res.Pop {
				_, c := s.Pop()
				return c
			}
			if res.Consumed {
				// Emit a resize signal once the resize gesture has
				// completed: either a release (was resizing, now isn't)
				// or any other consumed event that actually changed the
				// content dims (handles intermediate motion as well).
				justFinishedResize := wasResizing && !ent.layer.Resizing
				dimsChanged := ent.layer.ContentSizeInitialized &&
					(ent.layer.ContentWidth != oldW || ent.layer.ContentHeight != oldH)
				if justFinishedResize || dimsChanged {
					if c := s.notifyResize(ent, ent.layer.ContentWidth, ent.layer.ContentHeight); c != nil {
						return c
					}
				}
				return nil
			}
		}
		if cfg.CloseOnClickOutside && isPrimaryPress(msg) {
			if s.lastW > 0 && s.lastH > 0 {
				if !layout.CellInModal(msg.X, msg.Y, t, l, mw, mh) {
					_, c := s.Pop()
					return c
				}
			}
		}
		var c tea.Cmd
		ent.model, c = ent.model.Update(msg)
		return c

	default:
		top := len(s.entries) - 1
		var c tea.Cmd
		s.entries[top].model, c = s.entries[top].model.Update(msg)
		return c
	}
}

func isEscapeKey(msg tea.KeyMsg) bool {
	return msg.Type == tea.KeyEsc || msg.Type == tea.KeyEscape
}

func isPrimaryPress(msg tea.MouseMsg) bool {
	if msg.Action != tea.MouseActionPress {
		return false
	}
	return msg.Button == tea.MouseButtonLeft
}
