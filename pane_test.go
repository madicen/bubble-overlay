package overlay

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func TestWindow_emptyKeyClearsState(t *testing.T) {
	var w Window
	w.State.OriginInitialized = true
	w.State.OriginTop = 5
	w.lastKey = "a"
	out := w.View("base", "", "", "", 80, 25)
	if out != "base" {
		t.Fatal("empty key should pass through base")
	}
	if w.lastKey != "" || w.State.OriginInitialized {
		t.Fatalf("expected cleared state, lastKey=%q state=%+v", w.lastKey, w.State)
	}
}

func TestWindow_keyChangeResetsOrigin(t *testing.T) {
	var w Window
	w.View("base", "aaa", "A", "a", 80, 25)
	w.State.OriginInitialized = true
	w.State.OriginTop = 9
	w.State.OriginLeft = 4
	w.View("base", "bbb", "B", "b", 80, 25)
	if w.State.OriginTop == 9 && w.State.OriginLeft == 4 {
		t.Fatal("key change should reset drag origin")
	}
}

func TestWindow_closeButtonReturnsCloseCmd(t *testing.T) {
	var w Window
	closeCmd := tea.Cmd(func() tea.Msg { return nil })
	w.View("........", "hi", "T", "k", 80, 25)
	layout, ok := w.LastChromeLayout()
	if !ok {
		t.Fatal("expected layout")
	}
	x := layout.Left + layout.Regions.CloseX
	y := layout.Top + layout.Regions.CloseY
	consumed, got := w.Update(tea.MouseMsg{
		X: x, Y: y,
		Button: tea.MouseButtonLeft, Action: tea.MouseActionPress,
	}, "hi", "T", "k", 80, 25, closeCmd)
	if !consumed {
		t.Fatal("close click should be consumed")
	}
	if got == nil {
		t.Fatal("expected closeCmd to be returned on [x]")
	}
	if w.lastKey != "" {
		t.Fatal("close should clear active key")
	}
}

// TestWindow_minimizeButtonClickToggles pins the [-] press path through
// Window.Update (the embed used by consumers that build their own modal
// loop on top of the layout-cached chrome handler). The layout-handler
// branch in chrome_layout.go regressed silently before this test —
// HandleChromePointer in window.go had minimize support but the layout
// variant skipped straight from the close check to the resize/drag
// checks, so chrome consumers wired through Window saw [-] presses
// turn into drags instead of minimize toggles.
func TestWindow_minimizeButtonClickToggles(t *testing.T) {
	var w Window
	w.Configure = func(c *OverlayConfig) {
		c.WindowChrome.ShowMinimizeButton = true
	}
	w.View("........", "first line\nsecond line", "T", "k", 80, 25)
	layout, ok := w.LastChromeLayout()
	if !ok {
		t.Fatal("expected layout after View")
	}
	if layout.Regions.MinimizeW == 0 {
		t.Fatalf("ShowMinimizeButton should paint the minimize region; got regions=%+v", layout.Regions)
	}

	x := layout.Left + layout.Regions.MinimizeX
	y := layout.Top + layout.Regions.MinimizeY
	consumed, _ := w.Update(tea.MouseMsg{
		X: x, Y: y,
		Button: tea.MouseButtonLeft, Action: tea.MouseActionPress,
	}, "first line\nsecond line", "T", "k", 80, 25, nil)
	if !consumed {
		t.Fatal("press on [-] must be consumed by chrome — without consumption the event leaks to the underlay")
	}
	if !w.State.Minimized {
		t.Fatalf("press on [-] must flip LayerState.Minimized; state=%+v", w.State)
	}
	if w.State.Dragging {
		t.Fatal("[-] press must not leave Dragging=true — that would re-position the modal on the next motion event")
	}
}

// TestWindow_doubleClickTabTogglesMinimize pins the OS-style double-click
// gesture through the same Window.Update layout-cached path. We pre-arm
// LastTabPressAt to skip the first press timing so the test stays
// deterministic.
func TestWindow_doubleClickTabTogglesMinimize(t *testing.T) {
	var w Window
	w.Configure = func(c *OverlayConfig) {
		c.WindowChrome.ShowMinimizeButton = true
	}
	w.View("........", "first line\nsecond line", "T", "k", 80, 25)
	layout, ok := w.LastChromeLayout()
	if !ok {
		t.Fatal("expected layout after View")
	}

	// Aim a single column to the LEFT of the buttons cluster so the
	// press routes to cellInTabDrag, not the [-] / [x] hit-rects.
	x := layout.Left + layout.Regions.TabLeft + 1
	y := layout.Top + layout.Regions.TabTop + 1

	// Simulate "first press was 50ms ago" — well inside the 500ms
	// DoubleClickThreshold — without sleeping.
	w.State.LastTabPressAt = time.Now().Add(-50 * time.Millisecond)

	consumed, _ := w.Update(tea.MouseMsg{
		X: x, Y: y,
		Button: tea.MouseButtonLeft, Action: tea.MouseActionPress,
	}, "first line\nsecond line", "T", "k", 80, 25, nil)
	if !consumed {
		t.Fatal("second press in tab drag area must be consumed")
	}
	if !w.State.Minimized {
		t.Fatalf("double-click on tab must toggle Minimized; state=%+v", w.State)
	}
	if w.State.Dragging {
		t.Fatal("double-click handler must cancel the second press's drag")
	}
	if !w.State.LastTabPressAt.IsZero() {
		t.Fatal("LastTabPressAt should reset after toggle so a third quick click doesn't immediately re-toggle")
	}
}

func TestWindow_singleRenderPerFrame(t *testing.T) {
	var w Window
	w.View("base", "x", "T", "k", 80, 25)
	modal1 := w.frameModal
	_, _ = w.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}, "x", "T", "k", 80, 25, nil)
	if w.frameModal != modal1 {
		t.Fatal("Update should reuse cached modal within the same frame")
	}
}

func TestLayerState_reset(t *testing.T) {
	st := &LayerState{OriginInitialized: true, OriginTop: 9, Dragging: true}
	st.Reset()
	if st.OriginInitialized || st.Dragging {
		t.Fatalf("Reset: %+v", st)
	}
}

func TestWindow_keyboardResizeMsg(t *testing.T) {
	var w Window
	w.View(strings.Repeat(".", 80), "one\ntwo", "R", "k", 80, 25)
	_, cmd := w.Update(tea.KeyMsg{Type: tea.KeyShiftRight, Alt: true}, "one\ntwo", "R", "k", 80, 25, nil)
	if cmd == nil {
		t.Fatal("resize key should emit cmd")
	}
	if _, ok := cmd().(WindowResizedMsg); !ok {
		t.Fatalf("expected WindowResizedMsg, got %T", cmd())
	}
}
