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
	s.entries = append(s.entries, stackEntry{model: m, cfg: cfg})
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

func (s *OverlayStack) MainReceivesKeyMsg() bool {
	return s == nil || len(s.entries) == 0
}

func (s *OverlayStack) MainReceivesMouseMsg() bool {
	return s.MainReceivesKeyMsg()
}

func (s *OverlayStack) View(baseMain string, viewW, viewH int) string {
	if s == nil || len(s.entries) == 0 {
		return baseMain
	}
	cur := baseMain
	for i := range s.entries {
		cfg := s.entries[i].cfg
		if cfg.DimOpacity > 0 {
			cur = DimSurface(cur, cfg.DimOpacity)
		}
		modal := s.entries[i].model.View()
		mw, mh := layout.ModalCellSize(modal)
		top, left := cfg.Placement.Origin(mw, mh, viewW, viewH)
		cur = OverlayView(cur, modal, viewW, viewH, top, left)
	}
	return cur
}

func (s *OverlayStack) topLayout(viewW, viewH int) (top, left, mw, mh int) {
	if s == nil || len(s.entries) == 0 {
		return 0, 0, 0, 0
	}
	ent := s.entries[len(s.entries)-1]
	modal := ent.model.View()
	mw, mh = layout.ModalCellSize(modal)
	top, left = ent.cfg.Placement.Origin(mw, mh, viewW, viewH)
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
			var c tea.Cmd
			s.entries[i].model, c = s.entries[i].model.Update(msg)
			cmds = append(cmds, c)
		}
		return tea.Batch(cmds...)

	case tea.KeyMsg:
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
		if cfg.CloseOnClickOutside && isPrimaryPress(msg) {
			if s.lastW > 0 && s.lastH > 0 {
				t, l, mw, mh := s.topLayout(s.lastW, s.lastH)
				if !layout.CellInModal(msg.X, msg.Y, t, l, mw, mh) {
					_, c := s.Pop()
					return c
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

func isEscapeKey(msg tea.KeyMsg) bool {
	return msg.Type == tea.KeyEsc || msg.Type == tea.KeyEscape
}

func isPrimaryPress(msg tea.MouseMsg) bool {
	if msg.Action != tea.MouseActionPress {
		return false
	}
	return msg.Button == tea.MouseButtonLeft
}
