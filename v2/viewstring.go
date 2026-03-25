package overlayv2

import tea "charm.land/bubbletea/v2"

// ViewString returns the ANSI string content of a [tea.View] for compositing.
// It reads the [tea.View.Content] field populated by the model's View().
func ViewString(v tea.View) string {
	return v.Content
}
