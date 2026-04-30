package overlay

import tea "github.com/charmbracelet/bubbletea"

type FocusTrap struct {
	Stack *OverlayStack
}

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
