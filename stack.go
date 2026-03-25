package overlay

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/madicen/bubble-overlay/internal/layout"
)

// OverlayOnCloser is implemented by overlay models that need a cleanup cmd when
// the frame is popped (Escape, click-outside, or manual Pop).
type OverlayOnCloser interface {
	OnOverlayClose() tea.Cmd
}

type stackEntry struct {
	model tea.Model
	cfg   OverlayConfig
}

// OverlayStack manages nested tea.Model overlays with a single composited View,
// focus-trapped input routing, and optional Escape / click-outside dismissal.
type OverlayStack struct {
	entries      []stackEntry
	lastW, lastH int // set from WindowSizeMsg for click-outside layout
}

// Depth returns the number of overlays on the stack (0 = no overlay).
func (s *OverlayStack) Depth() int {
	if s == nil {
		return 0
	}
	return len(s.entries)
}

// StackDepth is an alias for Depth for dev / assert ergonomics.
func (s *OverlayStack) StackDepth() int {
	return s.Depth()
}

// Push adds a model above the current stack. Returns Init() cmd from the model.
func (s *OverlayStack) Push(m tea.Model, cfg OverlayConfig) tea.Cmd {
	if s == nil {
		return nil
	}
	s.entries = append(s.entries, stackEntry{model: m, cfg: cfg})
	return m.Init()
}

// Pop removes the top overlay. It returns the popped model and any OnOverlayClose cmd.
// If the stack is empty, popped is nil.
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

// Top returns the top overlay model or nil if empty.
func (s *OverlayStack) Top() tea.Model {
	if s == nil || len(s.entries) == 0 {
		return nil
	}
	return s.entries[len(s.entries)-1].model
}

// MainReceivesKeyMsg is false when at least one overlay is shown and keys are trapped
// to the stack (callers should not run their base model's key handling in that case).
func (s *OverlayStack) MainReceivesKeyMsg() bool {
	return s == nil || len(s.entries) == 0
}

// MainReceivesMouseMsg is false when overlays trap mouse to the top frame (same as keys).
func (s *OverlayStack) MainReceivesMouseMsg() bool {
	return s.MainReceivesKeyMsg()
}

// View composites baseMain with each stacked model's View, outermost last.
// viewW and viewH are the program viewport size in cells.
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

// topLayout returns the same top, left, mw, mh used for View for the top entry.
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

// Update routes messages to stacked models. Policy:
//   - tea.WindowSizeMsg: every overlay receives it in stack order (bottom to top).
//   - tea.KeyMsg / tea.MouseMsg: only the top overlay when non-empty; Escape and
//     click-outside may Pop before the model sees the message.
//   - other messages: only the top overlay (timers, custom msgs from cmds owned by it).
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
