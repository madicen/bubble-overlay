package overlayv2

import tea "charm.land/bubbletea/v2"

// FocusTrap helps the root model skip sending interactive input to the base layer.
type FocusTrap struct {
	Stack *Stack
}

// InteractiveToBase is false for key and mouse messages when an overlay has focus.
func (f FocusTrap) InteractiveToBase(msg tea.Msg) bool {
	if !isInteractiveMsg(msg) {
		return true
	}
	if f.Stack == nil {
		return true
	}
	return f.Stack.MainReceivesKeys()
}

func isInteractiveMsg(msg tea.Msg) bool {
	switch msg.(type) {
	case tea.KeyPressMsg, tea.KeyReleaseMsg:
		return true
	case tea.MouseMsg:
		return true
	default:
		return false
	}
}
