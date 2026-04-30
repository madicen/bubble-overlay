package overlayv2

import tea "charm.land/bubbletea/v2"

type FocusTrap struct {
	Stack *Stack
}

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
