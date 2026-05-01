package overlay

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

type staticModel struct{ view string }

func (staticModel) Init() tea.Cmd                         { return nil }
func (m staticModel) Update(tea.Msg) (tea.Model, tea.Cmd) { return m, nil }
func (m staticModel) View() string                        { return m.view }

type keySpy struct {
	id   int
	keys *[]int
}

func (m *keySpy) Init() tea.Cmd { return nil }
func (m *keySpy) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if _, ok := msg.(tea.KeyMsg); ok {
		*m.keys = append(*m.keys, m.id)
	}
	return m, nil
}
func (m *keySpy) View() string { return "modal" }

type wsSpy struct {
	sizes *[]tea.WindowSizeMsg
}

func (m *wsSpy) Init() tea.Cmd { return nil }
func (m *wsSpy) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if ws, ok := msg.(tea.WindowSizeMsg); ok {
		*m.sizes = append(*m.sizes, ws)
	}
	return m, nil
}
func (m *wsSpy) View() string { return "w" }

type closerSpy struct {
	staticModel
	closed *bool
}

func (c *closerSpy) OnOverlayClose() tea.Cmd {
	*c.closed = true
	return nil
}

func TestOverlayStack_depth_push_pop(t *testing.T) {
	var s OverlayStack
	if s.Depth() != 0 {
		t.Fatalf("Depth want 0 got %d", s.Depth())
	}
	s.Push(staticModel{"a"}, DefaultOverlayConfig())
	s.Push(staticModel{"b"}, DefaultOverlayConfig())
	if s.Depth() != 2 {
		t.Fatalf("Depth want 2 got %d", s.Depth())
	}
	popped, _ := s.Pop()
	if popped == nil {
		t.Fatal("Pop returned nil")
	}
	if s.Depth() != 1 {
		t.Fatalf("Depth want 1 got %d", s.Depth())
	}
	_, _ = s.Pop()
	if s.Depth() != 0 {
		t.Fatalf("Depth want 0 got %d", s.Depth())
	}
	popped, _ = s.Pop()
	if popped != nil {
		t.Fatal("Pop on empty should return nil model")
	}
}

func TestOverlayStack_window_size_broadcast(t *testing.T) {
	var got []tea.WindowSizeMsg
	var s OverlayStack
	s.Push(&wsSpy{sizes: &got}, DefaultOverlayConfig())
	s.Push(&wsSpy{sizes: &got}, DefaultOverlayConfig())
	s.Update(tea.WindowSizeMsg{Width: 40, Height: 12})
	if len(got) != 2 {
		t.Fatalf("want both overlays to see size, got %d events", len(got))
	}
}

func TestOverlayStack_click_outside_pops(t *testing.T) {
	var s OverlayStack
	cfg := DefaultOverlayConfig()
	cfg.CloseOnClickOutside = true
	cfg.Placement = Fixed(5, 5)
	cfg.DimOpacity = 0
	s.Push(staticModel{view: "MMMMM\nMMMMM\nMMMMM"}, cfg)
	s.Update(tea.WindowSizeMsg{Width: 30, Height: 20})
	s.Update(tea.MouseMsg{
		X: 0, Y: 0, Action: tea.MouseActionPress, Button: tea.MouseButtonLeft,
	})
	if s.Depth() != 0 {
		t.Fatalf("click outside should pop, depth=%d", s.Depth())
	}
}

func TestOverlayStack_click_inside_uses_clamped_origin_wide_modal(t *testing.T) {
	// Modal wider than viewport: Origin leaves left=5 but OverlayView clamps to 0.
	// Hit-testing must use ClampedOrigin so a click on the painted left edge counts as inside.
	var s OverlayStack
	cfg := DefaultOverlayConfig()
	cfg.CloseOnClickOutside = true
	cfg.Placement = Fixed(5, 5)
	cfg.DimOpacity = 0
	s.Push(staticModel{view: strings.Repeat("M", 40)}, cfg)
	s.Update(tea.WindowSizeMsg{Width: 30, Height: 20})
	s.Update(tea.MouseMsg{
		X: 2, Y: 5, Action: tea.MouseActionPress, Button: tea.MouseButtonLeft,
	})
	if s.Depth() != 1 {
		t.Fatalf("click on painted modal should not pop, depth=%d", s.Depth())
	}
}

func TestOverlayStack_pop_invokes_OnOverlayClose(t *testing.T) {
	var closed bool
	var s OverlayStack
	m := &closerSpy{staticModel: staticModel{view: "x"}, closed: &closed}
	s.Push(m, DefaultOverlayConfig())
	_, cmd := s.Pop()
	if cmd != nil {
		t.Fatalf("unexpected cmd %v", cmd)
	}
	if !closed {
		t.Fatal("OnOverlayClose not called")
	}
}

func TestOverlayStack_escape_pops_top(t *testing.T) {
	var s OverlayStack
	s.Push(staticModel{"x"}, DefaultOverlayConfig())
	s.Update(tea.WindowSizeMsg{Width: 80, Height: 25})
	s.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if s.Depth() != 0 {
		t.Fatalf("Escape should pop, depth=%d", s.Depth())
	}
}

func TestOverlayStack_key_only_top_model(t *testing.T) {
	var hits []int
	var s OverlayStack
	s.Push(&keySpy{id: 1, keys: &hits}, DefaultOverlayConfig())
	s.Push(&keySpy{id: 2, keys: &hits}, DefaultOverlayConfig())
	s.Update(tea.WindowSizeMsg{Width: 80, Height: 25})
	s.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	if len(hits) != 1 || hits[0] != 2 {
		t.Fatalf("want top-only key [2], got %v", hits)
	}
}

func TestPlacement_rightDrawer_and_center(t *testing.T) {
	top, left := RightDrawer().Origin(10, 4, 80, 25)
	if top != 0 || left != 70 {
		t.Fatalf("RightDrawer: got top=%d left=%d", top, left)
	}
	top, left = Center().Origin(10, 4, 80, 25)
	if top != 10 || left != 35 {
		t.Fatalf("Center: got top=%d left=%d", top, left)
	}
}

func TestFocusTrap_InteractiveToBase(t *testing.T) {
	var s OverlayStack
	ft := FocusTrap{Stack: &s}
	if !ft.InteractiveToBase(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}) {
		t.Fatal("empty stack: base should receive keys")
	}
	s.Push(staticModel{"x"}, DefaultOverlayConfig())
	if ft.InteractiveToBase(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}) {
		t.Fatal("stack open: base should not receive keys")
	}
	type customMsg struct{}
	if !ft.InteractiveToBase(customMsg{}) {
		t.Fatal("non-interactive msgs not trapped from base")
	}
}

func TestOverlayStack_View_nested(t *testing.T) {
	main := strings.Repeat(".", 10) + "\n" + strings.Repeat(".", 10)
	var s OverlayStack
	s.Push(staticModel{view: "OO"}, OverlayConfig{Placement: Fixed(0, 0), DimOpacity: 0})
	s.Push(staticModel{view: "XX"}, OverlayConfig{Placement: Fixed(0, 8), DimOpacity: 0})
	out := s.View(main, 10, 2)
	if !strings.Contains(out, "XX") || !strings.Contains(out, "OO") {
		t.Fatalf("expected nested overlays in view:\n%s", out)
	}
}

func TestDevStackDepthFooter(t *testing.T) {
	if DevStackDepthFooter(3, false) != "" {
		t.Fatal("dev false should hide footer")
	}
	if !strings.Contains(DevStackDepthFooter(2, true), "2") {
		t.Fatalf("footer: %q", DevStackDepthFooter(2, true))
	}
}
