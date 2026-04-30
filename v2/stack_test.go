package overlayv2

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	bov "github.com/madicen/bubble-overlay"
)

type staticV2 struct{ view string }

func (staticV2) Init() tea.Cmd                         { return nil }
func (m staticV2) Update(tea.Msg) (tea.Model, tea.Cmd) { return m, nil }
func (m staticV2) View() tea.View                      { return tea.NewView(m.view) }

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
	s.Push(staticV2{view: "a"}, bov.DefaultOverlayConfig())
	s.Push(staticV2{view: "b"}, bov.DefaultOverlayConfig())
	if s.Depth() != 2 {
		t.Fatal("depth")
	}
	s.Pop()
	if s.Depth() != 1 {
		t.Fatal("depth")
	}
	s.Pop()
	if s.Depth() != 0 {
		t.Fatal("depth")
	}
}

func TestModalStackIntegrity_key_only_top(t *testing.T) {
	var hits []int
	var s Stack
	s.Push(&keySpyV2{id: 1, keys: &hits}, bov.DefaultOverlayConfig())
	s.Push(&keySpyV2{id: 2, keys: &hits}, bov.DefaultOverlayConfig())
	s.Update(tea.WindowSizeMsg{Width: 80, Height: 25})
	var k tea.KeyPressMsg
	k.Code = 'a'
	k.Text = "a"
	s.Update(k)
	if len(hits) != 1 || hits[0] != 2 {
		t.Fatalf("want [2] got %v", hits)
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
		t.Fatalf("depth %d", s.Depth())
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
		t.Fatalf("depth %d", s.Depth())
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
			t.Fatalf("want %d lines got %d for %dx%d", sz.h, lines, sz.w, sz.h)
		}
	}
}

func TestCompositeView_matches_string_pipeline(t *testing.T) {
	main := strings.Repeat(".", 10) + "\n" + strings.Repeat(".", 10)
	var s Stack
	s.Push(staticV2{view: "XX"}, bov.OverlayConfig{Placement: bov.Fixed(0, 8), DimOpacity: 0})
	v := s.CompositeView(main, 10, 2)
	want := bov.OverlayView(main, "XX", 10, 2, 0, 8)
	if v.Content != want {
		t.Fatalf("mismatch:\nwant %q\ngot  %q", want, v.Content)
	}
}
