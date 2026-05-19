package overlay

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// minimizeTrackingModel records OverlayMinimizedMsg / OnOverlayMinimize
// callbacks so tests can assert which signal fired and in what order.
type minimizeTrackingModel struct {
	view              string
	onCallbackVal     bool
	onCallbackCalled  bool
	msgVal            bool
	msgReceived       bool
	overlayTitleValue string
}

func (m *minimizeTrackingModel) Init() tea.Cmd { return nil }
func (m *minimizeTrackingModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if mm, ok := msg.(OverlayMinimizedMsg); ok {
		m.msgVal = mm.Minimized
		m.msgReceived = true
	}
	return m, nil
}
func (m *minimizeTrackingModel) View() string { return m.view }
func (m *minimizeTrackingModel) OnOverlayMinimize(minimized bool) tea.Cmd {
	m.onCallbackVal = minimized
	m.onCallbackCalled = true
	return nil
}
func (m *minimizeTrackingModel) OverlayTitle() string {
	if m.overlayTitleValue == "" {
		return "tracker"
	}
	return m.overlayTitleValue
}

// rendersMinimizeChromeFixture builds a stack with a single chromed
// window that opts into both minimize and close buttons, plus enough
// body width that the tab fits its glyphs without truncation. Tests
// share this builder so geometry assertions stay focused on the
// minimize behaviour rather than fiddly setup.
func rendersMinimizeChromeFixture(t *testing.T, m tea.Model) (*OverlayStack, OverlayConfig) {
	t.Helper()
	cfg := DefaultOverlayConfig()
	cfg.WindowChrome = EnableWindowChrome("review · idle")
	cfg.WindowChrome.ShowMinimizeButton = true
	cfg.WindowChrome.Resizable = true
	cfg.WindowChrome.MinWidth = 30
	cfg.WindowChrome.MinHeight = 6
	s := &OverlayStack{}
	s.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	s.Push(m, cfg)
	s.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	return s, cfg
}

// TestWindowChrome_minimizeButtonGlyphsHaveSameWidth is the load-bearing
// invariant behind the close button keeping its column when minimize
// toggles. If the two glyphs ever drift in width, ComputeChromeRegions
// would compute different CloseX values for the two states and clicks
// on [x] would mistarget after toggling. The minimize button position
// itself stays put because it's derived from CloseX − constant width.
func TestWindowChrome_minimizeButtonGlyphsHaveSameWidth(t *testing.T) {
	if got, want := lipgloss.Width(MinimizeButtonGlyph), MinimizeButtonWidth; got != want {
		t.Fatalf("MinimizeButtonGlyph width %d != MinimizeButtonWidth %d", got, want)
	}
	if got, want := lipgloss.Width(RestoreButtonGlyph), MinimizeButtonWidth; got != want {
		t.Fatalf("RestoreButtonGlyph width %d != MinimizeButtonWidth %d", got, want)
	}
}

// TestWindowChrome_minimizeRendersOnlyTabStrip is the headline rendering
// behaviour: when LayerState.Minimized is true, the modal collapses to
// just the tab cap + tab row + flat bottom (three rows above the
// configurable TabOffsetTop). No body, no resizable border below.
//
// We measure the modal string directly (via RenderEntryModal) rather
// than the full composited view, since OverlayStack.View pads to the
// viewport height regardless of modal size.
func TestWindowChrome_minimizeRendersOnlyTabStrip(t *testing.T) {
	s, _ := rendersMinimizeChromeFixture(t, staticModel{view: strings.Repeat("M", 40) + "\n" + strings.Repeat("M", 40) + "\n" + strings.Repeat("M", 40)})

	ent := &s.entries[len(s.entries)-1]
	effCfg := ent.effectiveConfig()

	expanded := RenderEntryModal(ent.model.View(), effCfg, &ent.layer)
	expandedLines := strings.Count(expanded, "\n") + 1

	ent.layer.Minimized = true
	minimized := RenderEntryModal(ent.model.View(), effCfg, &ent.layer)
	minimizedLines := strings.Count(minimized, "\n") + 1

	if minimizedLines >= expandedLines {
		t.Fatalf("minimized modal line count (%d) should drop below expanded (%d); minimized:\n%s\nexpanded:\n%s", minimizedLines, expandedLines, minimized, expanded)
	}

	// Minimized output must still contain the chrome title — the whole
	// UX point of minimize is "you can still see what's collapsed".
	if !strings.Contains(minimized, effCfg.WindowChrome.Title) {
		t.Fatalf("minimized view must still render the chrome title %q; got:\n%s", effCfg.WindowChrome.Title, minimized)
	}
	if !strings.Contains(minimized, RestoreButtonGlyph) {
		t.Fatalf("minimized view must show the restore glyph %q so the user can click to expand; got:\n%s", RestoreButtonGlyph, minimized)
	}
}

// TestWindowChrome_minimizeTabRoundTripsTitle pins the title to BOTH
// states: expanded shows MinimizeButtonGlyph, minimized shows
// RestoreButtonGlyph, and both keep the dynamic OverlayTitler value
// (so a spinner-laden phase title remains visible while collapsed).
func TestWindowChrome_minimizeTabRoundTripsTitle(t *testing.T) {
	tracker := &minimizeTrackingModel{
		view:              strings.Repeat("M", 40) + "\n" + strings.Repeat("M", 40),
		overlayTitleValue: "review · ⠋ running",
	}
	s, _ := rendersMinimizeChromeFixture(t, tracker)

	expandedView := s.View("bg", 120, 40)
	if !strings.Contains(expandedView, MinimizeButtonGlyph) {
		t.Fatalf("expanded view must show minimize glyph %q", MinimizeButtonGlyph)
	}
	if !strings.Contains(expandedView, "running") {
		t.Fatalf("expanded view should carry OverlayTitler value, got:\n%s", expandedView)
	}

	s.entries[0].layer.Minimized = true
	minimizedView := s.View("bg", 120, 40)
	if !strings.Contains(minimizedView, RestoreButtonGlyph) {
		t.Fatalf("minimized view must show restore glyph %q", RestoreButtonGlyph)
	}
	if !strings.Contains(minimizedView, "running") {
		t.Fatalf("minimized view should still carry OverlayTitler value, got:\n%s", minimizedView)
	}
}

// TestOverlayStack_minimizeClickTogglesLayerAndBroadcasts is the
// end-to-end mouse path: a press on the minimize button must (a) flip
// LayerState.Minimized, (b) fire OnOverlayMinimize, (c) deliver
// OverlayMinimizedMsg to the model's Update, (d) NOT pop the overlay,
// and (e) leave the modal interactive (clicking restore brings it
// back).
func TestOverlayStack_minimizeClickTogglesLayerAndBroadcasts(t *testing.T) {
	tracker := &minimizeTrackingModel{
		view:              strings.Repeat("M", 40) + "\n" + strings.Repeat("M", 40) + "\n" + strings.Repeat("M", 40),
		overlayTitleValue: "review · idle", // match the static fallback so geometry is stable
	}
	s, _ := rendersMinimizeChromeFixture(t, tracker)

	// First View() seeds the layer's content size and origin.
	_ = s.View("bg", 120, 40)

	pressMsg := minimizeButtonPress(s, 120, 40)
	if c := s.Update(pressMsg); c != nil {
		// notifyMinimize batches commands; we drain them so any
		// follow-up tea.Cmd resolves before our assertions.
		_ = c()
	}

	if !s.entries[0].layer.Minimized {
		t.Fatal("minimize click did not flip LayerState.Minimized")
	}
	if s.Depth() != 1 {
		t.Fatalf("minimize click should not pop the overlay (depth=%d want 1)", s.Depth())
	}
	if !tracker.onCallbackCalled || !tracker.onCallbackVal {
		t.Fatalf("expected OnOverlayMinimize(true) callback, got called=%v val=%v", tracker.onCallbackCalled, tracker.onCallbackVal)
	}
	if !tracker.msgReceived || !tracker.msgVal {
		t.Fatalf("expected OverlayMinimizedMsg{Minimized: true} delivery, got received=%v val=%v", tracker.msgReceived, tracker.msgVal)
	}

	// Reset the tracker so we can see the restore callback in isolation,
	// then click again — the glyph at that position is now
	// RestoreButtonGlyph and the layer is collapsed.
	tracker.onCallbackCalled = false
	tracker.onCallbackVal = false
	tracker.msgReceived = false
	tracker.msgVal = false

	_ = s.View("bg", 120, 40)
	restoreMsg := minimizeButtonPress(s, 120, 40)
	if c := s.Update(restoreMsg); c != nil {
		_ = c()
	}
	if s.entries[0].layer.Minimized {
		t.Fatal("restore click did not flip LayerState.Minimized back to false")
	}
	if !tracker.onCallbackCalled || tracker.onCallbackVal {
		t.Fatalf("expected OnOverlayMinimize(false) on restore, got called=%v val=%v", tracker.onCallbackCalled, tracker.onCallbackVal)
	}
	if !tracker.msgReceived || tracker.msgVal {
		t.Fatalf("expected OverlayMinimizedMsg{Minimized: false} on restore, got received=%v val=%v", tracker.msgReceived, tracker.msgVal)
	}
}

// minimizeButtonPress derives the exact painted position of the
// minimize / restore button for the stack's top entry. We have to use
// the EFFECTIVE config (via effectiveConfig + ContentSizeForLayer),
// because OverlayTitler can change the rendered title — and therefore
// the tab width and inner button positions — between frames. Computing
// regions from the static cfg.WindowChrome.Title would mis-locate the
// button whenever a model returns a non-empty dynamic title.
func minimizeButtonPress(s *OverlayStack, viewW, viewH int) tea.MouseMsg {
	ent := &s.entries[len(s.entries)-1]
	cfg := ent.effectiveConfig()
	wc := cfg.WindowChrome.effective()
	modal := RenderEntryModal(ent.model.View(), cfg, &ent.layer)
	cw, ch := ContentSizeForLayer(modal, wc, &ent.layer)
	reg := ComputeChromeRegions(wc, cw, ch)
	top, left, _, _ := s.topLayout(viewW, viewH)
	return tea.MouseMsg{
		X:      left + reg.MinimizeX + reg.MinimizeW/2,
		Y:      top + reg.MinimizeY,
		Action: tea.MouseActionPress,
		Button: tea.MouseButtonLeft,
	}
}

// TestOverlayStack_minimizeSuppressesResizeHits guarantees clicks on
// what *would have been* a resize handle land are inert while
// minimized — the body isn't painted, so its right/bottom edges aren't
// a real target. Without this, a hidden resize gesture could start in
// invisible space and produce confusing geometry on restore.
func TestOverlayStack_minimizeSuppressesResizeHits(t *testing.T) {
	tracker := &minimizeTrackingModel{view: strings.Repeat("M", 40) + "\n" + strings.Repeat("M", 40) + "\n" + strings.Repeat("M", 40)}
	s, cfg := rendersMinimizeChromeFixture(t, tracker)
	_ = s.View("bg", 120, 40)

	// Locate the corner resize handle while the modal is still
	// expanded so we know exactly where it would be.
	cw, ch := s.entries[0].layer.ContentWidth, s.entries[0].layer.ContentHeight
	reg := ComputeChromeRegions(cfg.WindowChrome, cw, ch)
	top, left, _, _ := s.topLayout(120, 40)
	cornerMsg := tea.MouseMsg{
		X:      left + reg.ResizeCornerX,
		Y:      top + reg.ResizeCornerY,
		Action: tea.MouseActionPress,
		Button: tea.MouseButtonLeft,
	}

	// Flip to minimized and click that coordinate. Layer must NOT
	// enter the Resizing state.
	s.entries[0].layer.Minimized = true
	_ = s.View("bg", 120, 40)
	s.Update(cornerMsg)
	if s.entries[0].layer.Resizing {
		t.Fatal("clicking the would-be resize corner while minimized started a resize gesture — handles must be inert")
	}
}

// TestOverlayStack_minimizeCancelsActiveDragAndResize guards against a
// state-machine wedge: if the user has a drag mid-flight and somehow
// triggers a minimize toggle (e.g. via a keyboard binding a host adds
// later), the chrome must drop the gesture flags so the next release
// doesn't try to apply motion against the collapsed window.
func TestOverlayStack_minimizeCancelsActiveDragAndResize(t *testing.T) {
	tracker := &minimizeTrackingModel{view: strings.Repeat("M", 40) + "\n" + strings.Repeat("M", 40)}
	s, _ := rendersMinimizeChromeFixture(t, tracker)
	_ = s.View("bg", 120, 40)

	// Force-set both gesture flags (simulating in-flight drag/resize).
	s.entries[0].layer.Dragging = true
	s.entries[0].layer.Resizing = true
	s.entries[0].layer.ResizeEdge = ResizeCorner

	// Synthesize a minimize-button press by reusing the toggle path
	// directly on the layer's chrome handler. Re-deriving the exact
	// minimize coords here would duplicate the click-toggle test;
	// we just need to confirm the cancel semantics fire.
	cfg := s.entries[0].cfg
	cfg.WindowChrome = cfg.WindowChrome.effective()
	cw, ch := s.entries[0].layer.ContentWidth, s.entries[0].layer.ContentHeight
	reg := ComputeChromeRegions(cfg.WindowChrome, cw, ch)
	modal := RenderEntryModal(tracker.View(), cfg, &s.entries[0].layer)
	mw, mh := ModalCellSize(modal)
	top, left, _, _ := s.topLayout(120, 40)
	res := HandleChromePointer(
		ChromePointerPress, true,
		left+reg.MinimizeX+reg.MinimizeW/2,
		top+reg.MinimizeY,
		cfg, &s.entries[0].layer,
		modal, top, left, mw, mh, 120, 40,
	)
	if !res.MinimizeToggled {
		t.Fatal("expected HandleChromePointer to flag MinimizeToggled when minimize button is pressed")
	}
	if s.entries[0].layer.Dragging || s.entries[0].layer.Resizing {
		t.Fatalf("minimize toggle must cancel active gestures: dragging=%v resizing=%v", s.entries[0].layer.Dragging, s.entries[0].layer.Resizing)
	}
}

// TestOverlayStack_doubleClickTabTogglesMinimize verifies the
// OS-style "double-click the title bar to minimize / restore" gesture:
// two presses on the tab drag area within DoubleClickThreshold should
// flip LayerState.Minimized, cancel any drag the first press kicked
// off, and broadcast via the usual MinimizeToggled path.
//
// We exploit the exported LastTabPressAt field to simulate "first
// press was 50ms ago" without sleeping, keeping the test fast and
// deterministic.
func TestOverlayStack_doubleClickTabTogglesMinimize(t *testing.T) {
	tracker := &minimizeTrackingModel{
		view:              strings.Repeat("M", 40) + "\n" + strings.Repeat("M", 40),
		overlayTitleValue: "review · idle",
	}
	s, _ := rendersMinimizeChromeFixture(t, tracker)
	_ = s.View("bg", 120, 40)

	// Pre-arm the layer as if a press happened 50ms ago — well inside
	// the 500ms threshold.
	s.entries[0].layer.LastTabPressAt = time.Now().Add(-50 * time.Millisecond)

	// Second press at a known tab-drag coordinate. Pick a column
	// inside the tab body but to the LEFT of the buttons cluster so
	// the press routes to the drag-area branch, not the minimize
	// button hit-test.
	pressMsg := tabDragPress(s, 120, 40)
	if c := s.Update(pressMsg); c != nil {
		_ = c() // drain notifyMinimize batched cmd
	}

	if !s.entries[0].layer.Minimized {
		t.Fatalf("double-click on tab should have minimized the window; LastTabPressAt=%v", s.entries[0].layer.LastTabPressAt)
	}
	if s.entries[0].layer.Dragging {
		t.Fatal("double-click handler must cancel the second press's drag — leaving Dragging=true would re-position the modal on the next motion")
	}
	if !tracker.onCallbackCalled || !tracker.onCallbackVal {
		t.Fatalf("double-click should fire OverlayMinimizer.OnOverlayMinimize(true), got called=%v val=%v", tracker.onCallbackCalled, tracker.onCallbackVal)
	}
	if !s.entries[0].layer.LastTabPressAt.IsZero() {
		t.Fatal("LastTabPressAt should reset after toggle so a third quick click doesn't immediately re-toggle")
	}
}

// TestOverlayStack_doubleClickTabBeyondThresholdDoesNotToggle is the
// other side of the gesture: two presses spaced WIDER than
// DoubleClickThreshold are two independent clicks, not a double-click.
// The first arms LastTabPressAt; the second sees a stale timestamp,
// re-arms it, and starts a normal drag.
func TestOverlayStack_doubleClickTabBeyondThresholdDoesNotToggle(t *testing.T) {
	tracker := &minimizeTrackingModel{
		view:              strings.Repeat("M", 40) + "\n" + strings.Repeat("M", 40),
		overlayTitleValue: "review · idle",
	}
	s, _ := rendersMinimizeChromeFixture(t, tracker)
	_ = s.View("bg", 120, 40)

	// Simulate "first press was 2 seconds ago" — well beyond the
	// 500ms threshold.
	s.entries[0].layer.LastTabPressAt = time.Now().Add(-2 * time.Second)

	pressMsg := tabDragPress(s, 120, 40)
	if c := s.Update(pressMsg); c != nil {
		_ = c()
	}

	if s.entries[0].layer.Minimized {
		t.Fatal("press more than DoubleClickThreshold after the previous one must NOT toggle minimize")
	}
	if !s.entries[0].layer.Dragging {
		t.Fatal("stale-timestamp press should start a fresh drag, not be swallowed by the double-click branch")
	}
	if tracker.onCallbackCalled {
		t.Fatalf("OverlayMinimizer hook should not fire when the gesture is not a double-click")
	}
}

// TestOverlayStack_doubleClickTabWithoutMinimizeButtonNoOp guards the
// discoverability gate: without ShowMinimizeButton, the user has no
// visible affordance to discover the double-click gesture, so we
// intentionally don't honour it — the second press starts a normal
// drag instead. (Consumers that want bare double-click minimize can
// always set ShowMinimizeButton=true; making the gate opt-out instead
// of opt-in would surprise existing consumers.)
func TestOverlayStack_doubleClickTabWithoutMinimizeButtonNoOp(t *testing.T) {
	cfg := DefaultOverlayConfig()
	cfg.WindowChrome = EnableWindowChrome("no-min")
	cfg.WindowChrome.ShowMinimizeButton = false // explicit
	cfg.WindowChrome.Resizable = true
	cfg.WindowChrome.MinWidth = 30
	cfg.WindowChrome.MinHeight = 6
	s := &OverlayStack{}
	s.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	s.Push(staticModel{view: strings.Repeat("M", 40) + "\n" + strings.Repeat("M", 40)}, cfg)
	s.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	_ = s.View("bg", 120, 40)

	s.entries[0].layer.LastTabPressAt = time.Now().Add(-50 * time.Millisecond)
	s.Update(tabDragPress(s, 120, 40))

	if s.entries[0].layer.Minimized {
		t.Fatal("without ShowMinimizeButton, double-click must not toggle minimize")
	}
}

// tabDragPress builds a left-button press at a coordinate guaranteed
// to live inside the tab drag area but to the LEFT of the buttons
// cluster — column tabLeft+1 hits the lead space of the inner text,
// which is plain background (no glyph hit-rect overlaps it).
func tabDragPress(s *OverlayStack, viewW, viewH int) tea.MouseMsg {
	ent := &s.entries[len(s.entries)-1]
	cfg := ent.effectiveConfig()
	wc := cfg.WindowChrome.effective()
	modal := RenderEntryModal(ent.model.View(), cfg, &ent.layer)
	cw, ch := ContentSizeForLayer(modal, wc, &ent.layer)
	reg := ComputeChromeRegions(wc, cw, ch)
	top, left, _, _ := s.topLayout(viewW, viewH)
	return tea.MouseMsg{
		X:      left + reg.TabLeft + 1, // lead space inside the tab
		Y:      top + reg.TabTop + 1,   // tab body row
		Action: tea.MouseActionPress,
		Button: tea.MouseButtonLeft,
	}
}

// TestOverlayStack_singleClickTabArmsTimestamp is the load-bearing
// invariant behind detect-second-press: a single press on the tab
// drag area must record LastTabPressAt so the next press has
// something to compare against. Without this, the double-click branch
// would never fire.
func TestOverlayStack_singleClickTabArmsTimestamp(t *testing.T) {
	tracker := &minimizeTrackingModel{
		view:              strings.Repeat("M", 40) + "\n" + strings.Repeat("M", 40),
		overlayTitleValue: "review · idle",
	}
	s, _ := rendersMinimizeChromeFixture(t, tracker)
	_ = s.View("bg", 120, 40)

	before := time.Now()
	s.Update(tabDragPress(s, 120, 40))
	after := time.Now()

	got := s.entries[0].layer.LastTabPressAt
	if got.IsZero() {
		t.Fatal("expected LastTabPressAt to be set after a tab drag press, got zero")
	}
	if got.Before(before) || got.After(after) {
		t.Fatalf("LastTabPressAt %v should fall inside [%v, %v]", got, before, after)
	}
}
