package overlay

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/madicen/bubble-overlay/internal/layout"
)

type OverlayOnCloser interface {
	OnOverlayClose() tea.Cmd
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
	cfg := ent.cfg
	wc := cfg.WindowChrome.effective()
	modal := RenderEntryModal(ent.model.View(), cfg, &ent.layer)
	top, left = EntryClampedOrigin(cfg, &ent.layer, modal, viewW, viewH)
	cw, ch := ContentSizeForLayer(modal, wc, &ent.layer)
	reg = ComputeChromeRegions(wc, cw, ch)
	return top, left, reg, true
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
	cfg := ent.cfg
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
	modal := RenderEntryModal(ent.model.View(), ent.cfg, &ent.layer)
	mw, mh = layout.ModalCellSize(modal)
	top, left = EntryClampedOrigin(ent.cfg, &ent.layer, modal, viewW, viewH)
	return top, left, mw, mh
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
			modal := RenderEntryModal(s.entries[i].model.View(), s.entries[i].cfg, &s.entries[i].layer)
			ReclampLayerOrigin(s.entries[i].cfg, &s.entries[i].layer, modal, msg.Width, msg.Height)
			var c tea.Cmd
			s.entries[i].model, c = s.entries[i].model.Update(msg)
			cmds = append(cmds, c)
		}
		return tea.Batch(cmds...)

	case tea.KeyMsg:
		top := len(s.entries) - 1
		ent := &s.entries[top]
		cfg := ent.cfg
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
			if res := HandleChromeKey(msg, cfg, &ent.layer, modal, t, l, mw, mh, viewW, viewH); res.Consumed {
				return nil
			}
		}
		var c tea.Cmd
		ent.model, c = ent.model.Update(msg)
		return c

	case tea.MouseMsg:
		top := len(s.entries) - 1
		ent := &s.entries[top]
		cfg := ent.cfg
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
			res := HandleChromeMouse(msg, cfg, &ent.layer, modal, t, l, mw, mh, viewW, viewH)
			if res.Pop {
				_, c := s.Pop()
				return c
			}
			if res.Consumed {
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
