package overlay

import tea "github.com/charmbracelet/bubbletea"

// FocusTrap is a small helper for root models: when the stack is non-empty,
// tea.KeyMsg and tea.MouseMsg should not be passed to the base layer so input
// cannot leak behind the overlay. Other messages (e.g. ticks, custom cmds) are
// usually still delivered to both base and stack per your app’s needs.
type FocusTrap struct {
	Stack *OverlayStack
}

// InteractiveToBase reports whether key/mouse messages should be delegated to
// the main model. When false, only OverlayStack.Update should handle them.
func (f FocusTrap) InteractiveToBase(msg tea.Msg) bool {
	switch msg.(type) {
	case tea.KeyMsg, tea.MouseMsg:
		if f.Stack == nil {
			return true
		}
		return f.Stack.MainReceivesKeyMsg()
	default:
		return true
	}
}
