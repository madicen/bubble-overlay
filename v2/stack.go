package overlayv2

import (
	tea "charm.land/bubbletea/v2"
	bov "github.com/madicen/bubble-overlay"
	"github.com/madicen/bubble-overlay/internal/layout"
)

type OverlayOnCloser interface {
	OnOverlayClose() tea.Cmd
}

type stackEntry struct {
	model tea.Model
	cfg   bov.OverlayConfig
	layer bov.LayerState
}

type Stack struct {
	entries      []stackEntry
	lastW, lastH int
	Adapter      ViewAdapter
}

func (s *Stack) Depth() int {
	if s == nil {
		return 0
	}
	return len(s.entries)
}

func (s *Stack) StackDepth() int { return s.Depth() }

func (s *Stack) Push(m tea.Model, cfg bov.OverlayConfig) tea.Cmd {
	if s == nil {
		return nil
	}
	if cfg.WindowChrome.Enabled {
		cfg.WindowChrome = cfg.WindowChrome.Effective()
	}
	ent := stackEntry{model: m, cfg: cfg}
	bov.InitLayerContentSize(&ent.layer, ViewString(m.View()), cfg.WindowChrome.Effective())
	s.entries = append(s.entries, ent)
	return m.Init()
}

func (s *Stack) Pop() (popped tea.Model, cmd tea.Cmd) {
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

func (s *Stack) Top() tea.Model {
	if s == nil || len(s.entries) == 0 {
		return nil
	}
	return s.entries[len(s.entries)-1].model
}

func (s *Stack) MainReceivesKeys() bool {
	return s == nil || len(s.entries) == 0
}

func (s *Stack) MainReceivesMouseMsg() bool { return s.MainReceivesKeys() }
func (s *Stack) MainReceivesKeyMsg() bool   { return s.MainReceivesKeys() }

func (s *Stack) adapter() ViewAdapter {
	if s == nil || s.Adapter == nil {
		return StringPipelineAdapter{}
	}
	return s.Adapter
}

func (s *Stack) syncViewport(viewW, viewH int) {
	if s == nil || viewW <= 0 || viewH <= 0 {
		return
	}
	s.lastW, s.lastH = viewW, viewH
}

func (s *Stack) CompositeView(base string, viewW, viewH int) tea.View {
	if s == nil || len(s.entries) == 0 {
		return tea.NewView(base)
	}
	s.syncViewport(viewW, viewH)
	frames := make([]FrameView, 0, len(s.entries))
	for i := range s.entries {
		modelView := ViewString(s.entries[i].model.View())
		frames = append(frames, FrameView{
			ModelView: modelView,
			Cfg:       s.entries[i].cfg,
			Layer:     &s.entries[i].layer,
		})
	}
	return s.adapter().Adapt(base, frames, viewW, viewH)
}

func (s *Stack) topLayout(viewW, viewH int) (top, left, mw, mh int) {
	if s == nil || len(s.entries) == 0 {
		return 0, 0, 0, 0
	}
	ent := &s.entries[len(s.entries)-1]
	modal := bov.RenderEntryModal(ViewString(ent.model.View()), ent.cfg, &ent.layer)
	mw, mh = layout.ModalCellSize(modal)
	top, left = bov.EntryClampedOrigin(ent.cfg, &ent.layer, modal, viewW, viewH)
	return top, left, mw, mh
}

// SetTopLayerOrigin sets the painted origin for the top entry when it uses draggable window chrome.
func (s *Stack) SetTopLayerOrigin(top, left int) {
	if s == nil || len(s.entries) == 0 {
		return
	}
	ent := &s.entries[len(s.entries)-1]
	ent.layer.OriginTop = top
	ent.layer.OriginLeft = left
	ent.layer.OriginInitialized = true
}

func (s *Stack) Update(msg tea.Msg) tea.Cmd {
	if s == nil || len(s.entries) == 0 {
		return nil
	}
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		s.lastW, s.lastH = msg.Width, msg.Height
		var cmds []tea.Cmd
		for i := range s.entries {
			modal := bov.RenderEntryModal(ViewString(s.entries[i].model.View()), s.entries[i].cfg, &s.entries[i].layer)
			bov.ReclampLayerOrigin(s.entries[i].cfg, &s.entries[i].layer, modal, msg.Width, msg.Height)
			var c tea.Cmd
			s.entries[i].model, c = s.entries[i].model.Update(msg)
			cmds = append(cmds, c)
		}
		return tea.Batch(cmds...)
	case tea.KeyPressMsg:
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
		if cfg.WindowChrome.Effective().Keyboard {
			modal := bov.RenderEntryModal(ViewString(ent.model.View()), cfg, &ent.layer)
			t, l, mw, mh := s.topLayout(viewW, viewH)
			if res := bov.HandleChromeKeyString(msg.String(), cfg, &ent.layer, modal, t, l, mw, mh, viewW, viewH); res.Consumed {
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
			modal := bov.RenderEntryModal(ViewString(ent.model.View()), cfg, &ent.layer)
			if res := s.handleChromeMouse(msg, cfg, ent, modal, t, l, mw, mh, viewW, viewH); res.Pop {
				_, c := s.Pop()
				return c
			} else if res.Consumed {
				return nil
			}
		}
		if cfg.CloseOnClickOutside {
			if mc, ok := msg.(tea.MouseClickMsg); ok && mc.Button == tea.MouseLeft {
				if s.lastW > 0 && s.lastH > 0 {
					m := mc.Mouse()
					if !layout.CellInModal(m.X, m.Y, t, l, mw, mh) {
						_, c := s.Pop()
						return c
					}
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

func (s *Stack) handleChromeMouse(msg tea.MouseMsg, cfg bov.OverlayConfig, ent *stackEntry, modal string, top, left, mw, mh, viewW, viewH int) bov.ChromeMouseResult {
	st := &ent.layer
	switch msg := msg.(type) {
	case tea.MouseClickMsg:
		return bov.HandleChromePointer(bov.ChromePointerPress, msg.Button == tea.MouseLeft, msg.X, msg.Y, cfg, st, modal, top, left, mw, mh, viewW, viewH)
	case tea.MouseReleaseMsg:
		return bov.HandleChromePointer(bov.ChromePointerRelease, msg.Button == tea.MouseLeft, msg.X, msg.Y, cfg, st, modal, top, left, mw, mh, viewW, viewH)
	case tea.MouseMotionMsg:
		return bov.HandleChromePointer(bov.ChromePointerMotion, msg.Button == tea.MouseLeft, msg.X, msg.Y, cfg, st, modal, top, left, mw, mh, viewW, viewH)
	default:
		return bov.ChromeMouseResult{}
	}
}

func isEscapeKey(msg tea.KeyPressMsg) bool {
	k := msg.Key()
	if k.Code == tea.KeyEscape || k.Code == tea.KeyEsc {
		return true
	}
	return msg.String() == "esc"
}
