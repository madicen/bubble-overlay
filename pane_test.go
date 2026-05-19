package overlay

import (
	"strings"
	"testing"

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
