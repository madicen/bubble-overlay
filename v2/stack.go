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
	s.entries = append(s.entries, stackEntry{model: m, cfg: cfg})
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

func (s *Stack) CompositeView(base string, viewW, viewH int) tea.View {
	if s == nil || len(s.entries) == 0 {
		return tea.NewView(base)
	}
	frames := make([]FrameView, 0, len(s.entries))
	for i := range s.entries {
		modal := ViewString(s.entries[i].model.View())
		frames = append(frames, FrameView{Modal: modal, Cfg: s.entries[i].cfg})
	}
	return s.adapter().Adapt(base, frames, viewW, viewH)
}

func (s *Stack) topLayout(viewW, viewH int) (top, left, mw, mh int) {
	if s == nil || len(s.entries) == 0 {
		return 0, 0, 0, 0
	}
	ent := s.entries[len(s.entries)-1]
	modal := ViewString(ent.model.View())
	mw, mh = layout.ModalCellSize(modal)
	top, left = ent.cfg.Placement.Origin(mw, mh, viewW, viewH)
	return top, left, mw, mh
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
			var c tea.Cmd
			s.entries[i].model, c = s.entries[i].model.Update(msg)
			cmds = append(cmds, c)
		}
		return tea.Batch(cmds...)
	case tea.KeyPressMsg:
		top := len(s.entries) - 1
		cfg := s.entries[top].cfg
		if cfg.CloseOnEscape && isEscapeKey(msg) {
			_, c := s.Pop()
			return c
		}
		var c tea.Cmd
		s.entries[top].model, c = s.entries[top].model.Update(msg)
		return c
	case tea.MouseMsg:
		top := len(s.entries) - 1
		cfg := s.entries[top].cfg
		if cfg.CloseOnClickOutside {
			if mc, ok := msg.(tea.MouseClickMsg); ok && mc.Button == tea.MouseLeft {
				if s.lastW > 0 && s.lastH > 0 {
					m := mc.Mouse()
					t, l, mw, mh := s.topLayout(s.lastW, s.lastH)
					if !layout.CellInModal(m.X, m.Y, t, l, mw, mh) {
						_, c := s.Pop()
						return c
					}
				}
			}
		}
		var c tea.Cmd
		s.entries[top].model, c = s.entries[top].model.Update(msg)
		return c
	default:
		top := len(s.entries) - 1
		var c tea.Cmd
		s.entries[top].model, c = s.entries[top].model.Update(msg)
		return c
	}
}

func isEscapeKey(msg tea.KeyPressMsg) bool {
	k := msg.Key()
	if k.Code == tea.KeyEscape || k.Code == tea.KeyEsc {
		return true
	}
	return msg.String() == "esc"
}
