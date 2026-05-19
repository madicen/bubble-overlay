package overlay

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// titlerModel implements OverlayTitler so we can lock in dynamic-title
// behaviour through the OverlayStack render path.
type titlerModel struct {
	body  string
	title string
}

func (titlerModel) Init() tea.Cmd                         { return nil }
func (m titlerModel) Update(tea.Msg) (tea.Model, tea.Cmd) { return m, nil }
func (m titlerModel) View() string                        { return m.body }
func (m titlerModel) OverlayTitle() string                { return m.title }

// TestDefaultChromeMaskRune_isPUA pins the library default to a Private-Use-Area
// rune. The previous default U+FFFC (OBJECT REPLACEMENT CHARACTER) collides
// with glamour-rendered markdown and other rich content, which then bleeds
// through chromed modals as transparent holes.
func TestDefaultChromeMaskRune_isPUA(t *testing.T) {
	if DefaultChromeMaskRune == '\ufffc' {
		t.Fatalf("DefaultChromeMaskRune must not be U+FFFC (OBJECT REPLACEMENT CHARACTER) — it collides with real content")
	}
	if DefaultChromeMaskRune < '\ue000' || DefaultChromeMaskRune > '\uf8ff' {
		t.Fatalf("DefaultChromeMaskRune must live in the BMP Private Use Area U+E000..U+F8FF (got %U)", DefaultChromeMaskRune)
	}
}

// titledBody is wide enough that tabInnerText doesn't truncate the title
// down to nothing. The renderer derives tab width from the content width,
// so the body needs more cells than the title's display width.
const titledBody = "body content wide enough for a title to fit in the tab"

// TestOverlayStack_OverlayTitler_overridesStaticTitle proves that a model
// implementing OverlayTitler gets its dynamic title baked into the rendered
// chrome each frame, with the cfg.WindowChrome.Title acting as a fallback.
func TestOverlayStack_OverlayTitler_overridesStaticTitle(t *testing.T) {
	cfg := DefaultOverlayConfig()
	cfg.WindowChrome = EnableWindowChrome("static-fallback")
	var s OverlayStack
	s.Push(titlerModel{body: titledBody, title: "dynamic-now"}, cfg)
	got := s.View("background", 80, 25)
	if !strings.Contains(got, "dynamic-now") {
		t.Fatalf("expected dynamic title in composite output, got: %q", got)
	}
	if strings.Contains(got, "static-fallback") {
		t.Fatalf("static-fallback must not leak when OverlayTitler returns non-empty (got: %q)", got)
	}
}

// TestOverlayStack_OverlayTitler_emptyFallsBackToStatic guards the
// "interface is safe to opt into" promise — returning "" preserves the
// cfg.WindowChrome.Title without forcing the consumer to mirror it.
func TestOverlayStack_OverlayTitler_emptyFallsBackToStatic(t *testing.T) {
	cfg := DefaultOverlayConfig()
	cfg.WindowChrome = EnableWindowChrome("static-fallback")
	var s OverlayStack
	s.Push(titlerModel{body: titledBody, title: ""}, cfg)
	got := s.View("background", 80, 25)
	if !strings.Contains(got, "static-fallback") {
		t.Fatalf("expected static fallback when OverlayTitler returns empty, got: %q", got)
	}
}

// resizerSpy records OnOverlayResize calls and forwards OverlayResizedMsg
// captures through its Update so we can verify both signals fire.
type resizerSpy struct {
	view             string
	hookCalls        []resizeEvent
	updateCalls      []resizeEvent
}

type resizeEvent struct{ w, h int }

func (m *resizerSpy) Init() tea.Cmd { return nil }
func (m *resizerSpy) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if r, ok := msg.(OverlayResizedMsg); ok {
		m.updateCalls = append(m.updateCalls, resizeEvent{r.NewContentWidth, r.NewContentHeight})
	}
	return m, nil
}
func (m *resizerSpy) View() string { return m.view }
func (m *resizerSpy) OnOverlayResize(w, h int) tea.Cmd {
	m.hookCalls = append(m.hookCalls, resizeEvent{w, h})
	return nil
}

// TestOverlayStack_resizeFiresOnOverlayResizeAndMsg verifies that a single
// resize gesture (Alt+Shift+Right) reaches the model through BOTH paths —
// the OnOverlayResize hook and an OverlayResizedMsg delivered through Update.
// Both fire so consumers can pick whichever fits their architecture.
func TestOverlayStack_resizeFiresOnOverlayResizeAndMsg(t *testing.T) {
	cfg := DefaultOverlayConfig()
	cfg.WindowChrome = EnableWindowChrome("resizer")
	cfg.WindowChrome.Resizable = true
	cfg.WindowChrome.Keyboard = true
	cfg.WindowChrome.KeyStep = 3
	cfg.WindowChrome.MinWidth = 4
	cfg.WindowChrome.MinHeight = 2

	spy := &resizerSpy{view: strings.Repeat("x", 30)}
	var s OverlayStack
	// Tell the stack our viewport so it doesn't fall back to 80x25.
	s.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	s.Push(spy, cfg)

	// Re-broadcast size so the freshly-pushed entry has a layout cache.
	s.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

	// Alt+Shift+Right grows the content width by KeyStep cells. (v1 maps
	// shift+right with alt to KeyShiftRight, the same way window_test does.)
	s.Update(tea.KeyMsg{Type: tea.KeyShiftRight, Alt: true})

	if len(spy.hookCalls) == 0 {
		t.Fatalf("expected at least one OnOverlayResize callback, got 0")
	}
	if len(spy.updateCalls) == 0 {
		t.Fatalf("expected at least one OverlayResizedMsg delivery, got 0")
	}
	last := spy.hookCalls[len(spy.hookCalls)-1]
	if last.w <= 30 {
		t.Fatalf("expected width > 30 after grow, got %d", last.w)
	}
	if got := spy.updateCalls[len(spy.updateCalls)-1]; got != last {
		t.Fatalf("hook and msg disagreed on final dims: hook=%v msg=%v", last, got)
	}
}

// TestWindow_ConfigureOverridesDefaults proves consumers can swap out the
// Window's hard-coded defaults (CenterContent, MinWidth=32, MinHeight=6,
// default mask) via the Configure hook without forking the library.
func TestWindow_ConfigureOverridesDefaults(t *testing.T) {
	var captured OverlayConfig
	w := &Window{
		Configure: func(cfg *OverlayConfig) {
			cfg.WindowChrome.CenterContent = false
			cfg.WindowChrome.MinWidth = 8
			cfg.WindowChrome.MinHeight = 3
			cfg.WindowChrome.ChromeMaskRune = '\uE001'
			captured = *cfg
		},
	}
	// Drive a single frame so configFor runs through Configure.
	w.View("base", "hello", "title", "k", 60, 20)

	if captured.WindowChrome.CenterContent {
		t.Fatal("Configure should have cleared CenterContent")
	}
	if captured.WindowChrome.MinWidth != 8 {
		t.Fatalf("MinWidth want 8 got %d", captured.WindowChrome.MinWidth)
	}
	if captured.WindowChrome.MinHeight != 3 {
		t.Fatalf("MinHeight want 3 got %d", captured.WindowChrome.MinHeight)
	}
	if captured.WindowChrome.ChromeMaskRune != '\uE001' {
		t.Fatalf("ChromeMaskRune not propagated: got %U", captured.WindowChrome.ChromeMaskRune)
	}
	// And Enabled must remain on regardless of what Configure does, so the
	// Window's render assumptions hold.
	if !captured.WindowChrome.Enabled {
		t.Fatal("Configure must not be able to disable WindowChrome.Enabled (Window's render path depends on it)")
	}
}

// TestMouseTargetsTop_emptyStackTrue keeps callers safe when they call
// MouseTargetsTop unconditionally — with no overlays present there's no
// "top" to route to but the host owns dispatch either way, so returning
// true ("yes, top consumes it") gives the host a single safe pattern:
//
//	if stack.MouseTargetsTop(msg, w, h) { return stack.Update(msg) }
//	return host.Update(msg)
//
// works correctly whether or not an overlay is currently open.
func TestMouseTargetsTop_emptyStackTrue(t *testing.T) {
	var s OverlayStack
	if !s.MouseTargetsTop(tea.MouseMsg{X: 5, Y: 5}, 80, 25) {
		t.Fatal("empty stack should report true so the unconditional-call pattern stays safe")
	}
}

// TestMouseTargetsTop_insideRect lets a click inside the modal painted
// rect route to the top overlay entry.
func TestMouseTargetsTop_insideRect(t *testing.T) {
	cfg := DefaultOverlayConfig()
	cfg.WindowChrome = EnableWindowChrome("inside")
	var s OverlayStack
	s.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	s.Push(staticModel{view: strings.Repeat("M", 30) + "\n" + strings.Repeat("M", 30)}, cfg)
	s.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

	// The compositor centers a 30-wide / 2-tall body in a 120×40 viewport,
	// so cell (60, 20) is guaranteed to land on the modal regardless of
	// the chrome's exact tab offset.
	if !s.MouseTargetsTop(tea.MouseMsg{X: 60, Y: 20}, 120, 40) {
		t.Fatal("expected click at modal center to target top overlay")
	}
}

// TestMouseTargetsTop_outsideRect is the pass-through case: a click far
// from a small modal must NOT route to the overlay, so the host can
// forward it to its main model and the background stays interactive.
func TestMouseTargetsTop_outsideRect(t *testing.T) {
	cfg := DefaultOverlayConfig()
	cfg.WindowChrome = EnableWindowChrome("outside")
	var s OverlayStack
	s.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	s.Push(staticModel{view: "x\nx"}, cfg)
	s.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

	// (0, 0) is the top-left corner of the viewport; the modal is
	// centered so this is unambiguously outside its painted rect.
	if s.MouseTargetsTop(tea.MouseMsg{X: 0, Y: 0}, 120, 40) {
		t.Fatal("expected click at viewport origin to fall outside the modal — host should receive the event")
	}
}

// TestMouseTargetsTop_gestureInProgress is the load-bearing invariant
// behind pass-through: once a drag or resize has started inside the
// modal, every subsequent motion / release event belongs to the
// overlay regardless of where the cursor goes, otherwise the chrome
// state machine would get stuck mid-gesture.
func TestMouseTargetsTop_gestureInProgress(t *testing.T) {
	cfg := DefaultOverlayConfig()
	cfg.WindowChrome = EnableWindowChrome("drag")
	cfg.WindowChrome.Resizable = true

	var s OverlayStack
	s.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	s.Push(staticModel{view: "tiny"}, cfg)
	s.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

	// Directly flip the top entry's drag flag. Doing it via a synthetic
	// mouse press inside the tab would also work but couples the test
	// to chrome region geometry that's exercised elsewhere.
	s.entries[len(s.entries)-1].layer.Dragging = true

	// Now an event far outside the modal must still report "top consumes
	// it" so the chrome motion handler runs.
	if !s.MouseTargetsTop(tea.MouseMsg{X: 0, Y: 0, Action: tea.MouseActionMotion}, 120, 40) {
		t.Fatal("expected motion event during drag to stay with the top overlay")
	}
}

// TestWindow_ConfigureCannotDisableChrome guards the load-bearing
// "WindowChrome.Enabled stays true" invariant — if a consumer's Configure
// accidentally clears Enabled, the Window forces it back on so the chrome
// render path stays consistent.
func TestWindow_ConfigureCannotDisableChrome(t *testing.T) {
	w := &Window{
		Configure: func(cfg *OverlayConfig) {
			cfg.WindowChrome.Enabled = false
		},
	}
	got := w.View("base.................", "hello", "title", "k", 60, 20)
	// Tab title still rendered → chrome stayed enabled.
	if !strings.Contains(got, "title") {
		t.Fatalf("expected tab title in output (chrome should remain enabled), got: %q", got)
	}
}
