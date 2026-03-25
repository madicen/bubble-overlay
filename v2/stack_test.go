package overlayv2

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	bov "github.com/madicen/bubble-overlay"
)

type staticV2 struct{ view string }

func (staticV2) Init() tea.Cmd { return nil }

func (m staticV2) Update(tea.Msg) (tea.Model, tea.Cmd) { return m, nil }

func (m staticV2) View() tea.View { return tea.NewView(m.view) }

type keySpyV2 struct {
	id   int
	keys *[]int
}

func (m *keySpyV2) Init() tea.Cmd { return nil }

func (m *keySpyV2) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if _, ok := msg.(tea.KeyPressMsg); ok {
		*m.keys = append(*m.keys, m.id)
	}
	return m, nil
}

func (m *keySpyV2) View() tea.View { return tea.NewView("modal") }

type wsSpyV2 struct {
	sizes *[]tea.WindowSizeMsg
}

func (m *wsSpyV2) Init() tea.Cmd { return nil }

func (m *wsSpyV2) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if ws, ok := msg.(tea.WindowSizeMsg); ok {
		*m.sizes = append(*m.sizes, ws)
	}
	return m, nil
}

func (m *wsSpyV2) View() tea.View { return tea.NewView("w") }

func TestModalStackIntegrity_depth_push_pop(t *testing.T) {
	var s Stack
	if s.Depth() != 0 {
		t.Fatalf("depth want 0 got %d", s.Depth())
	}
	s.Push(staticV2{view: "a"}, bov.DefaultOverlayConfig())
	s.Push(staticV2{view: "b"}, bov.DefaultOverlayConfig())
	if s.Depth() != 2 {
		t.Fatalf("depth want 2 got %d", s.Depth())
	}
	popped, _ := s.Pop()
	if popped == nil {
		t.Fatal("pop")
	}
	if s.Depth() != 1 {
		t.Fatalf("depth want 1 got %d", s.Depth())
	}
	_, _ = s.Pop()
	if s.Depth() != 0 {
		t.Fatalf("depth want 0 got %d", s.Depth())
	}
}

func TestModalStackIntegrity_key_only_top(t *testing.T) {
	var hits []int
	var s Stack
	s.Push(&keySpyV2{id: 1, keys: &hits}, bov.DefaultOverlayConfig())
	s.Push(&keySpyV2{id: 2, keys: &hits}, bov.DefaultOverlayConfig())
	s.Update(tea.WindowSizeMsg{Width: 80, Height: 25})
	// printable key press
	var k tea.KeyPressMsg
	k.Code = 'a'
	k.Text = "a"
	s.Update(k)
	if len(hits) != 1 || hits[0] != 2 {
		t.Fatalf("want top-only [2], got %v", hits)
	}
}

func TestModalStackIntegrity_escape_pops(t *testing.T) {
	var s Stack
	s.Push(staticV2{view: "x"}, bov.DefaultOverlayConfig())
	s.Update(tea.WindowSizeMsg{Width: 80, Height: 25})
	var esc tea.KeyPressMsg
	esc.Code = tea.KeyEscape
	s.Update(esc)
	if s.Depth() != 0 {
		t.Fatalf("escape should pop, depth=%d", s.Depth())
	}
}

func TestModalStackIntegrity_window_size_broadcast(t *testing.T) {
	var got []tea.WindowSizeMsg
	var s Stack
	s.Push(&wsSpyV2{sizes: &got}, bov.DefaultOverlayConfig())
	s.Push(&wsSpyV2{sizes: &got}, bov.DefaultOverlayConfig())
	s.Update(tea.WindowSizeMsg{Width: 40, Height: 12})
	if len(got) != 2 {
		t.Fatalf("want 2 size msgs, got %d", len(got))
	}
}

func TestOverlayStability_line_count_after_resize(t *testing.T) {
	main := strings.Repeat("M", 40)
	base := strings.Join([]string{main, main, main, main}, "\n")
	var s Stack
	s.Push(staticV2{view: "OO\nOO"}, bov.OverlayConfig{Placement: bov.Fixed(0, 0), DimOpacity: 0})

	for _, sz := range []struct{ w, h int }{{30, 10}, {50, 6}, {80, 25}} {
		s.Update(tea.WindowSizeMsg{Width: sz.w, Height: sz.h})
		v := s.CompositeView(base, sz.w, sz.h)
		lines := strings.Count(v.Content, "\n") + 1
		if lines != sz.h {
			t.Fatalf("viewport %dx%d: want %d lines got %d", sz.w, sz.h, sz.h, lines)
		}
	}
}

func TestModalStackIntegrity_click_outside_pops(t *testing.T) {
	var s Stack
	cfg := bov.DefaultOverlayConfig()
	cfg.CloseOnClickOutside = true
	cfg.Placement = bov.Fixed(5, 5)
	cfg.DimOpacity = 0
	s.Push(staticV2{view: "MMMMM\nMMMMM\nMMMMM"}, cfg)
	s.Update(tea.WindowSizeMsg{Width: 30, Height: 20})
	s.Update(tea.MouseClickMsg{X: 0, Y: 0, Button: tea.MouseLeft})
	if s.Depth() != 0 {
		t.Fatalf("click outside should pop, depth=%d", s.Depth())
	}
}

func TestCompositeView_matches_string_pipeline(t *testing.T) {
	main := strings.Repeat(".", 10) + "\n" + strings.Repeat(".", 10)
	var s Stack
	s.Push(staticV2{view: "XX"}, bov.OverlayConfig{Placement: bov.Fixed(0, 8), DimOpacity: 0})
	v := s.CompositeView(main, 10, 2)
	want := bov.OverlayView(main, "XX", 10, 2, 0, 8)
	if v.Content != want {
		t.Fatalf("pipeline mismatch:\nwant: %q\ngot:  %q", want, v.Content)
	}
}
