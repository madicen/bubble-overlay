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
