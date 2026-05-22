package overlay

import (
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// DoubleClickThreshold is the maximum wall-clock gap between two
// chrome tab presses for the chrome handler to treat the second press
// as a double-click. 500ms matches the OS-level default on every major
// desktop platform, which is the gesture this feature mimics
// (double-click title bar to minimize / maximize).
//
// Consumers that need a different threshold can override it via
// LayerState.LastTabPressAt manipulation, but the constant itself is
// intentionally hard-coded so the gesture feels uniform across apps
// built on this library.
const DoubleClickThreshold = 500 * time.Millisecond

// CloseButtonGlyph is the close control label inside the tab.
const CloseButtonGlyph = "[x]"

// CloseButtonWidth is the display width of CloseButtonGlyph for hit-testing.
const CloseButtonWidth = 3

// MinimizeButtonGlyph is the label shown for the minimize control when the
// window is expanded. Clicking it collapses the window down to its tab.
const MinimizeButtonGlyph = "[-]"

// RestoreButtonGlyph is the label shown in place of the minimize control
// when the window is collapsed. Clicking it expands the body back to the
// last content size. The two glyphs share a width so the close button to
// their right doesn't shift between states.
const RestoreButtonGlyph = "[+]"

// MinimizeButtonWidth is the display width of MinimizeButtonGlyph (and
// RestoreButtonGlyph — they're sized identically for layout stability).
const MinimizeButtonWidth = 3

// DefaultChromeMaskRune is the pass-through padding rune for WindowChrome
// auto-wrap. It lives in the Unicode Private Use Area (U+E000), which is
// guaranteed not to appear in normal text — glamour-rendered markdown,
// terminal images, and other rich content occasionally include U+FFFC
// (OBJECT REPLACEMENT CHARACTER), which previously caused mergeMaskRune
// to punch transparent holes through chromed modals.
//
// Set WindowChrome.ChromeMaskRune to override if your content happens to
// produce U+E000 (extremely unlikely outside custom font pickers).
const DefaultChromeMaskRune = '\ue000'

const (
	defaultTabBackground = "238"
	defaultTabForeground = "252"
	defaultTabBorder     = "63"
	defaultTabOffsetTop  = 1
	defaultTabOffsetLeft = 0
	defaultMinContentW   = 20
	defaultMinContentH   = 4
)

// WindowChrome configures optional tab title bar, drag, and close for a stack entry.
// When Enabled is true, use EnableWindowChrome for recommended defaults, or set
// AutoWrap, Draggable, and ShowCloseButton explicitly.
type WindowChrome struct {
	Enabled            bool
	Title              string
	ShowCloseButton    bool
	ShowMinimizeButton bool // render [-] / [+] toggle to the left of the close button
	AutoWrap           bool
	TitleBarHeight     int // legacy; tab layout uses TabOffsetTop + tab rows when zero
	Draggable          bool
	TabBackground      string // lipgloss color for tab fill (default muted "238")
	TabForeground      string // lipgloss color for tab text
	TabBorder          string // lipgloss color for tab border runes
	TabOffsetTop       int    // rows above tab (default 1)
	TabOffsetLeft      int    // columns left of tab (default 0)
	ChromeMaskRune     rune   // pass-through padding in auto-wrap chrome; 0 uses DefaultChromeMaskRune
	Resizable          bool   // drag right/bottom edges (and corner) to resize content
	Keyboard           bool   // Alt+arrow move, Alt+Shift+arrow resize
	KeyStep            int    // cells per keypress when Keyboard is enabled (default 1)
	CenterContent      bool   // keep content centered in the body when Resizable
	ContentPadTop      int    // blank lines pinned to the top of the resizable body
	MinWidth           int    // minimum content width when Resizable
	MinHeight          int    // minimum content height (lines) when Resizable
}

// EnableWindowChrome returns a WindowChrome with tab title bar, drag, close, and auto-wrap enabled.
func EnableWindowChrome(title string) WindowChrome {
	return WindowChrome{
		Enabled:         true,
		Title:           title,
		ShowCloseButton: true,
		AutoWrap:        true,
		Draggable:       true,
		TabBackground:   MutedTabBackground(defaultTabBorder),
		TabForeground:   defaultTabForeground,
		TabBorder:       defaultTabBorder,
		TabOffsetTop:    defaultTabOffsetTop,
		TabOffsetLeft:   defaultTabOffsetLeft,
	}
}

// Effective returns chrome settings with defaults applied when Enabled.
func (w WindowChrome) Effective() WindowChrome {
	return w.effective()
}

func (w WindowChrome) effective() WindowChrome {
	if !w.Enabled {
		return w
	}
	out := w
	if out.TitleBarHeight < 1 {
		out.TitleBarHeight = out.tabOffsetTop() + 2
	}
	if out.TabBorder == "" {
		out.TabBorder = defaultTabBorder
	}
	if out.TabBackground == "" {
		out.TabBackground = MutedTabBackground(out.TabBorder)
	}
	if out.TabForeground == "" {
		out.TabForeground = defaultTabForeground
	}
	if out.TabOffsetTop <= 0 {
		out.TabOffsetTop = defaultTabOffsetTop
	}
	if out.TabOffsetLeft < 0 {
		out.TabOffsetLeft = defaultTabOffsetLeft
	}
	return out
}

// MutedTabBackground returns a fill color slightly darker than a 6×6×6 border index,
// lowering the strongest channel(s) so the hue stays aligned with the border.
// Non-cube values (hex names, grays) fall back to defaultTabBackground.
func MutedTabBackground(border string) string {
	n, err := strconv.Atoi(strings.TrimSpace(border))
	if err != nil || n < 16 || n > 231 {
		return defaultTabBackground
	}
	c := n - 16
	r, g, b := c/36, (c/6)%6, c%6
	steps := 1
	if max(r, g, b) >= 4 {
		steps = 2
	}
	for steps > 0 {
		switch {
		case b >= g && b >= r && b > 0:
			b--
		case g >= b && g >= r && g > 0:
			g--
		case r > 0:
			r--
		default:
			steps = 0
		}
		steps--
	}
	return strconv.Itoa(16 + 36*r + 6*g + b)
}

func (w WindowChrome) tabOffsetTop() int {
	if w.TabOffsetTop > 0 {
		return w.TabOffsetTop
	}
	return defaultTabOffsetTop
}

func (w WindowChrome) tabOffsetLeft() int {
	if w.TabOffsetLeft >= 0 {
		return w.TabOffsetLeft
	}
	return defaultTabOffsetLeft
}

func (w WindowChrome) autoWrap() bool {
	return w.Enabled && w.AutoWrap
}

func (w WindowChrome) draggable() bool {
	return w.Enabled && w.Draggable
}

func (w WindowChrome) showClose() bool {
	return w.Enabled && w.ShowCloseButton
}

func (w WindowChrome) showMinimize() bool {
	return w.Enabled && w.ShowMinimizeButton
}

func (w WindowChrome) resizable() bool {
	return w.Enabled && w.Resizable
}

func (w WindowChrome) keyboard() bool {
	return w.Enabled && w.Keyboard
}

func (w WindowChrome) keyStep() int {
	if w.KeyStep > 0 {
		return w.KeyStep
	}
	return 1
}

// tabOnBorder is true when the title row merges with the window top edge (resizable windows).
func (w WindowChrome) tabOnBorder() bool {
	return w.resizable()
}

func (w WindowChrome) centerContent() bool {
	return w.resizable() && w.CenterContent
}

func (w WindowChrome) contentPadTop() int {
	if w.ContentPadTop > 0 {
		return w.ContentPadTop
	}
	return 0
}

func (w WindowChrome) minWidth() int {
	if w.MinWidth > 0 {
		return w.MinWidth
	}
	return defaultMinContentW
}

func (w WindowChrome) minHeight() int {
	if w.MinHeight > 0 {
		return w.MinHeight
	}
	return defaultMinContentH
}

func (w WindowChrome) chromeMaskRune() rune {
	if !w.Enabled {
		return 0
	}
	if w.ChromeMaskRune == 0 {
		return DefaultChromeMaskRune
	}
	return w.ChromeMaskRune
}

// ModalBodyWidth returns the display width of content lines below the tab chrome.
func ModalBodyWidth(modal string, wc WindowChrome) int {
	wc = wc.effective()
	lines := strings.Split(modal, "\n")
	skip := wc.tabOffsetTop() + 2
	if skip >= len(lines) {
		return contentWidth(modal)
	}
	body := strings.Join(lines[skip:], "\n")
	if wc.resizable() && len(lines) > skip+2 {
		body = strings.Join(lines[skip+1:len(lines)-1], "\n")
	}
	return contentWidth(body)
}

// ModalBodyHeight returns the content line count below the tab chrome.
func ModalBodyHeight(modal string, wc WindowChrome) int {
	wc = wc.effective()
	lines := strings.Split(modal, "\n")
	skip := wc.tabOffsetTop() + 2
	if skip >= len(lines) {
		return 0
	}
	n := len(lines) - skip
	if wc.resizable() && n >= 2 {
		return n - 2
	}
	return n
}

func contentHeight(content string) int {
	if content == "" {
		return 0
	}
	return len(strings.Split(content, "\n"))
}

// ContentSizeForLayer returns the content dimensions used for layout and hit-testing.
func ContentSizeForLayer(modal string, wc WindowChrome, layer *LayerState) (w, h int) {
	if layer != nil && layer.ContentWidth > 0 && layer.ContentHeight > 0 && wc.resizable() {
		return layer.ContentWidth, layer.ContentHeight
	}
	return ModalBodyWidth(modal, wc), ModalBodyHeight(modal, wc)
}

// InitLayerContentSize seeds layer content dimensions from a model view on Push.
func InitLayerContentSize(layer *LayerState, modelView string, wc WindowChrome) {
	if layer == nil || !wc.resizable() {
		return
	}
	layer.ContentWidth = max(wc.minWidth(), contentWidth(modelView))
	layer.ContentHeight = max(wc.minHeight(), contentHeight(modelView))
	layer.ContentSizeInitialized = true
}

func (w WindowChrome) titleBarHeight() int {
	w = w.effective()
	if !w.Enabled {
		return 0
	}
	return w.TitleBarHeight
}

// ResizeEdge identifies which resize handle is active.
type ResizeEdge uint8

const (
	ResizeNone ResizeEdge = iota
	ResizeRight
	ResizeBottom
	ResizeCorner
)

// ChromeRegions describes tab, close, minimize, and resize hit areas relative to the modal's top-left cell.
type ChromeRegions struct {
	TabTop, TabLeft, TabW, TabH                                int
	CloseX, CloseY, CloseW, CloseH                             int
	MinimizeX, MinimizeY, MinimizeW, MinimizeH                 int
	ResizeRightX, ResizeRightY, ResizeRightW, ResizeRightH     int
	ResizeBottomX, ResizeBottomY, ResizeBottomW, ResizeBottomH int
	ResizeCornerX, ResizeCornerY, ResizeCornerW, ResizeCornerH int
}

// ComputeChromeRegions returns hit rectangles for a framed modal's content size.
//
// Region positions are based on the expanded tab layout. The minimize / restore
// buttons share a width (both glyphs are MinimizeButtonWidth), so the close
// button's column doesn't shift between expanded and minimized states — that
// keeps these regions valid as hit-test rects regardless of layer.Minimized.
// Resize regions are only populated when the modal is resizable AND has a body
// to drag against (contentH > 0); callers should additionally suppress resize
// dispatch when the layer is currently minimized.
func ComputeChromeRegions(wc WindowChrome, contentW, contentH int) ChromeRegions {
	wc = wc.effective()
	ot, ol := wc.tabOffsetTop(), wc.tabOffsetLeft()
	inner := tabInnerText(wc.Title, wc.showMinimize(), false, wc.showClose(), contentW)
	tabW := innerWidth(inner) + 2
	tabTop := ot
	tabLeft := ol
	reg := ChromeRegions{
		TabTop:  tabTop,
		TabLeft: tabLeft,
		TabW:    tabW,
		TabH:    2,
		CloseW:  CloseButtonWidth,
		CloseH:  1,
	}
	if wc.showClose() {
		reg.CloseX = tabLeft + 1 + lipgloss.Width(inner) - CloseButtonWidth
		reg.CloseY = tabTop + 1
	}
	if wc.showMinimize() {
		reg.MinimizeW = MinimizeButtonWidth
		reg.MinimizeH = 1
		reg.MinimizeY = tabTop + 1
		// When both buttons are shown the minimize sits exactly one
		// space to the left of close. When only minimize is shown, it
		// takes the close button's slot at the rightmost end of the
		// inner text. Computing from inner width (rather than mirroring
		// CloseX) keeps the formula correct in both cases.
		if wc.showClose() {
			reg.MinimizeX = reg.CloseX - MinimizeButtonWidth - 1
		} else {
			reg.MinimizeX = tabLeft + 1 + lipgloss.Width(inner) - MinimizeButtonWidth
		}
	}
	if wc.resizable() && contentW > 0 && contentH > 0 {
		bodyTop := ot + 2 // stacked: tab cap + tab row, then box top
		if wc.tabOnBorder() {
			bodyTop = ot + 1 // tab cap row, then title-on-border row is box top
		}
		boxW := contentW + 2
		bodyH := contentH + 2
		reg.ResizeRightX = boxW - 1
		reg.ResizeRightY = bodyTop
		reg.ResizeRightW = 1
		reg.ResizeRightH = bodyH
		reg.ResizeBottomX = 0
		reg.ResizeBottomY = bodyTop + bodyH - 1
		reg.ResizeBottomW = boxW
		reg.ResizeBottomH = 1
		reg.ResizeCornerX = boxW - 1
		reg.ResizeCornerY = bodyTop + bodyH - 1
		reg.ResizeCornerW = 1
		reg.ResizeCornerH = 1
	}
	return reg
}

// WindowFrameOpts styles the tab chrome rendered by WindowFrame.
type WindowFrameOpts struct {
	Width           int
	TitleStyle      lipgloss.Style
	TabBackground   string
	TabForeground   string
	TabBorder       string
	TabOffsetTop    int
	TabOffsetLeft   int
	ShowCloseButton bool
}

func (wc WindowChrome) frameOpts(contentWidth int) WindowFrameOpts {
	wc = wc.effective()
	return WindowFrameOpts{
		Width:           contentWidth,
		TabBackground:   wc.TabBackground,
		TabForeground:   wc.TabForeground,
		TabBorder:       wc.TabBorder,
		TabOffsetTop:    wc.tabOffsetTop(),
		TabOffsetLeft:   wc.tabOffsetLeft(),
		ShowCloseButton: wc.showClose(),
	}
}

// WindowFrame renders a tab-style title and bordered content. Width defaults to content width.
func WindowFrame(content, title string, opts WindowFrameOpts) string {
	cw := opts.Width
	if cw <= 0 {
		cw = contentWidth(content)
	}
	wc := WindowChrome{
		Enabled:         true,
		Title:           title,
		ShowCloseButton: opts.ShowCloseButton,
		TabBackground:   opts.TabBackground,
		TabForeground:   opts.TabForeground,
		TabBorder:       opts.TabBorder,
		TabOffsetTop:    opts.TabOffsetTop,
		TabOffsetLeft:   opts.TabOffsetLeft,
	}
	if wc.TabBackground == "" {
		wc.TabBackground = defaultTabBackground
	}
	if wc.TabForeground == "" {
		wc.TabForeground = defaultTabForeground
	}
	if wc.TabBorder == "" {
		wc.TabBorder = defaultTabBorder
	}
	if wc.TabOffsetTop <= 0 {
		wc.TabOffsetTop = defaultTabOffsetTop
	}
	if wc.TabOffsetLeft < 0 {
		wc.TabOffsetLeft = defaultTabOffsetLeft
	}
	// WindowFrame is a static one-shot framing helper without per-layer
	// state — minimize toggling lives on LayerState and only makes sense
	// in the stack-driven path, so always render expanded here.
	return renderTabFrame(content, cw, contentHeight(content), wc, false)
}

func contentWidth(content string) int {
	w := 0
	for _, line := range strings.Split(content, "\n") {
		if lw := lipgloss.Width(line); lw > w {
			w = lw
		}
	}
	return w
}

func tabTitleText(title string, contentWidth int) string {
	maxInner := max(0, contentWidth)
	if maxInner == 0 {
		return title
	}
	return truncateWidth(title, maxInner)
}

func tabInnerText(title string, showMinimize, minimized, showClose bool, contentWidth int) string {
	// Build the trailing buttons cluster: ` <minimize?> <close?> ` with a
	// single space between every part and a trailing space inside the tab.
	// The minimize glyph flips between MinimizeButtonGlyph and
	// RestoreButtonGlyph based on current state; both share a width so the
	// close button's column doesn't move as the state toggles.
	var parts []string
	if showMinimize {
		glyph := MinimizeButtonGlyph
		if minimized {
			glyph = RestoreButtonGlyph
		}
		parts = append(parts, glyph)
	}
	if showClose {
		parts = append(parts, CloseButtonGlyph)
	}
	suffix := " "
	if len(parts) > 0 {
		suffix = " " + strings.Join(parts, " ") + " "
	}
	prefix := " "
	maxTitle := max(0, contentWidth-lipgloss.Width(prefix+suffix))
	t := truncateWidth(title, maxTitle)
	return prefix + t + suffix
}

func innerWidth(inner string) int {
	return lipgloss.Width(inner)
}

func renderTabFrame(content string, cw, ch int, wc WindowChrome, minimized bool) string {
	ot, ol := wc.tabOffsetTop(), wc.tabOffsetLeft()
	mask := wc.chromeMaskRune()
	inner := tabInnerText(wc.Title, wc.showMinimize(), minimized, wc.showClose(), cw)
	iw := innerWidth(inner)
	tabW := iw + 2

	borderSt := lipgloss.NewStyle().Foreground(lipgloss.Color(wc.TabBorder))
	tabSt := lipgloss.NewStyle().
		Background(lipgloss.Color(wc.TabBackground)).
		Foreground(lipgloss.Color(wc.TabForeground))

	boxW := cw
	if wc.resizable() {
		boxW = cw + 2
	}
	modalW := max(boxW, ol+tabW)
	// When minimized, the window collapses to just the tab cap + tab row,
	// so the painted width only needs to cover the tab itself. Without
	// this, a minimized modal would still pad out to the full body width
	// with mask runes — defeating the visual point of minimizing.
	if minimized {
		modalW = ol + tabW
	}

	var out []string

	for range ot {
		out = append(out, maskFill(modalW, mask))
	}

	// Use rounded corners for draggable windows (when resizable)
	var tabTop string
	if wc.resizable() {
		tabTop = borderSt.Render("╭" + strings.Repeat("─", iw) + "╮")
	} else {
		tabTop = borderSt.Render("┌" + strings.Repeat("─", iw) + "┐")
	}
	tabRow := borderSt.Render("│") + tabSt.Render(inner) + borderSt.Render("│")
	out = append(out, composeChromeLine(maskFill(ol, mask), tabTop, maskFill(max(0, modalW-ol-tabW), mask)))

	if minimized {
		// Close the tab with a flat bottom border so the strip reads as
		// a self-contained pill: ┌────┐ / │ … │ / └────┘. We deliberately
		// skip both the tab-on-border body cap and the resizable body /
		// bottom border — the whole point of minimize is "show no body".
		out = append(out, composeChromeLine(maskFill(ol, mask), tabRow, maskFill(max(0, modalW-ol-tabW), mask)))
		var tabBot string
		if wc.resizable() {
			tabBot = borderSt.Render("╰" + strings.Repeat("─", iw) + "╯")
		} else {
			tabBot = borderSt.Render("└" + strings.Repeat("─", iw) + "┘")
		}
		out = append(out, composeChromeLine(maskFill(ol, mask), tabBot, maskFill(max(0, modalW-ol-tabW), mask)))
		return strings.Join(out, "\n")
	}

	if wc.resizable() && wc.tabOnBorder() {
		out = append(out, renderTabOnBorderLine(ol, iw, cw, inner, tabSt, borderSt, mask, modalW))
	} else {
		out = append(out, composeChromeLine(maskFill(ol, mask), tabRow, maskFill(max(0, modalW-ol-tabW), mask)))
	}

	if wc.resizable() {
		lines := fitContentLines(content, cw, ch, wc.centerContent(), wc.contentPadTop())
		var topBox string
		if !wc.tabOnBorder() {
			if wc.resizable() {
				topBox = borderSt.Render("╭" + strings.Repeat("─", cw) + "╮")
			} else {
				topBox = borderSt.Render("┌" + strings.Repeat("─", cw) + "┐")
			}
			out = append(out, padLineWithMask(topBox, modalW, mask))
		}
		for _, line := range lines {
			row := borderSt.Render("│") + line + borderSt.Render("│")
			out = append(out, padLineWithMask(row, modalW, mask))
		}
		var bottomBox string
		if wc.resizable() {
			bottomBox = borderSt.Render("╰" + strings.Repeat("─", cw) + "╯")
		} else {
			bottomBox = borderSt.Render("└" + strings.Repeat("─", cw) + "┘")
		}
		out = append(out, padLineWithMask(bottomBox, modalW, mask))
	} else {
		for _, line := range strings.Split(content, "\n") {
			out = append(out, padLineWithMask(line, modalW, mask))
		}
	}

	return strings.Join(out, "\n")
}

// renderTabOnBorderLine draws ├ title ┴──────┐ on the window's top edge (tab cap is the row above).
func renderTabOnBorderLine(ol, iw, cw int, inner string, tabSt, borderSt lipgloss.Style, mask rune, modalW int) string {
	boxW := cw + 2
	var prefix string
	if ol > 0 {
		prefix = borderSt.Render("╭" + strings.Repeat("─", ol-1))
	}
	tabPart := borderSt.Render("├") + tabSt.Render(inner) + borderSt.Render("┴")
	mid := prefix + tabPart
	used := lipgloss.Width(mid)
	hlineLen := boxW - used - 1 // remaining ─ before ┐
	if hlineLen < 0 {
		hlineLen = 0
	}
	right := borderSt.Render(strings.Repeat("─", hlineLen) + "╮")
	line := mid + right
	return padLineWithMask(line, modalW, mask)
}

func fitContentLines(content string, cw, ch int, center bool, padTop int) []string {
	if padTop < 0 {
		padTop = 0
	}
	if padTop >= ch {
		padTop = max(0, ch-1)
	}
	bodyH := ch - padTop
	lines := strings.Split(content, "\n")
	if bodyH <= 0 {
		bodyH = len(lines)
	}
	if len(lines) > bodyH {
		lines = lines[:bodyH]
	}
	for i, line := range lines {
		lines[i] = padLineWidth(truncateWidth(line, cw), cw, center)
	}
	if center && len(lines) < bodyH {
		top := (bodyH - len(lines)) / 2
		padded := make([]string, bodyH)
		for i := range padded {
			padded[i] = strings.Repeat(" ", cw)
		}
		copy(padded[top:], lines)
		lines = padded
	} else {
		for len(lines) < bodyH {
			lines = append(lines, strings.Repeat(" ", cw))
		}
	}
	if padTop == 0 {
		return lines
	}
	top := make([]string, padTop)
	for i := range top {
		top[i] = strings.Repeat(" ", cw)
	}
	return append(top, lines...)
}

func padLineWidth(line string, width int, center bool) string {
	w := lipgloss.Width(line)
	if w >= width {
		return line
	}
	pad := width - w
	if center {
		left := pad / 2
		return strings.Repeat(" ", left) + line + strings.Repeat(" ", pad-left)
	}
	return line + strings.Repeat(" ", pad)
}

func maskFill(n int, mask rune) string {
	if n <= 0 {
		return ""
	}
	return strings.Repeat(string(mask), n)
}

func composeChromeLine(left, mid, right string) string {
	return left + mid + right
}

func padLineWithMask(line string, total int, mask rune) string {
	w := lipgloss.Width(line)
	if w >= total {
		return line
	}
	return line + maskFill(total-w, mask)
}

func truncateWidth(s string, maxW int) string {
	if maxW <= 0 {
		return ""
	}
	if lipgloss.Width(s) <= maxW {
		return s
	}
	r := []rune(s)
	for len(r) > 0 && lipgloss.Width(string(r)) > maxW {
		r = r[:len(r)-1]
	}
	return string(r)
}

// LayerState holds per-entry origin and drag state for window chrome.
type LayerState struct {
	OriginTop, OriginLeft       int
	OriginInitialized           bool
	Dragging                    bool
	DragOffsetX                 int
	DragOffsetY                 int
	ContentWidth, ContentHeight int
	ContentSizeInitialized      bool
	Resizing                    bool
	ResizeEdge                  ResizeEdge
	ResizeStartX, ResizeStartY  int
	ResizeStartW, ResizeStartH  int
	// Minimized collapses the window to its tab strip (no body, no
	// bottom border) when true. The window stays draggable in this
	// state, but resize handles disappear; the minimize button glyph
	// flips to RestoreButtonGlyph so clicking it expands the body
	// back to its previous ContentWidth / ContentHeight.
	Minimized bool
	// LastTabPressAt is the wall-clock time of the most recent left-
	// button press that landed in the tab's drag area. The chrome
	// handler uses it to detect a double-click (two presses within
	// DoubleClickThreshold) on the tab strip, which toggles the
	// minimized state. Exposed so tests can simulate "the first press
	// was N ms ago" without needing real sleeps or a clock injection.
	LastTabPressAt time.Time
}

// ResetOrigin clears draggable origin so the next layout pass re-seeds from Placement.
func (st *LayerState) ResetOrigin() {
	if st == nil {
		return
	}
	st.OriginTop = 0
	st.OriginLeft = 0
	st.OriginInitialized = false
	st.Dragging = false
	st.DragOffsetX = 0
	st.DragOffsetY = 0
}

// Reset clears all layer state (origin, drag, resize, content size).
func (st *LayerState) Reset() {
	if st == nil {
		return
	}
	*st = LayerState{}
}

// RenderEntryModal returns the modal string for an entry, including auto-wrap chrome when configured.
func RenderEntryModal(modelView string, cfg OverlayConfig, layer *LayerState) string {
	wc := cfg.WindowChrome.effective()
	if !wc.autoWrap() {
		return modelView
	}
	cw := contentWidth(modelView)
	ch := contentHeight(modelView)
	if layer != nil && wc.resizable() {
		if layer.ContentWidth > 0 {
			cw = layer.ContentWidth
		}
		if layer.ContentHeight > 0 {
			ch = layer.ContentHeight
		}
	}
	minimized := layer != nil && layer.Minimized
	if wc.resizable() {
		return renderTabFrame(modelView, cw, ch, wc, minimized)
	}
	if minimized {
		// Non-resizable chrome doesn't go through renderTabFrame (it
		// uses WindowFrame, which is stateless). Route minimized layers
		// here so the painted output still collapses to the tab strip.
		return renderTabFrame(modelView, cw, contentHeight(modelView), wc, true)
	}
	return WindowFrame(modelView, wc.Title, wc.frameOpts(cw))
}

// ComposeModalLayer composites a framed modal over cur using mask pass-through when chrome is enabled.
func ComposeModalLayer(cur, modal string, cfg OverlayConfig, top, left, viewW, viewH int) string {
	if cfg.WindowChrome.Enabled {
		if mask := cfg.WindowChrome.chromeMaskRune(); mask != 0 {
			return OverlayViewWithMask(cur, modal, viewW, viewH, top, left, mask)
		}
	}
	return OverlayView(cur, modal, viewW, viewH, top, left)
}

// EntryClampedOrigin returns the compositor origin for an entry after seeding from Placement when needed.
func EntryClampedOrigin(cfg OverlayConfig, st *LayerState, modal string, viewW, viewH int) (top, left int) {
	mw, mh := ModalCellSize(modal)
	wc := cfg.WindowChrome.effective()
	if wc.draggable() {
		if st != nil && !st.OriginInitialized {
			t, l := cfg.Placement.Origin(mw, mh, viewW, viewH)
			if st != nil {
				st.OriginTop, st.OriginLeft = t, l
				st.OriginInitialized = true
			}
		}
		ot, ol := 0, 0
		if st != nil {
			ot, ol = st.OriginTop, st.OriginLeft
		}
		return Fixed(ot, ol).ClampedOrigin(mw, mh, viewW, viewH)
	}
	return cfg.Placement.ClampedOrigin(mw, mh, viewW, viewH)
}

// ReclampLayerOrigin clamps a draggable origin after viewport resize.
func ReclampLayerOrigin(cfg OverlayConfig, st *LayerState, modal string, viewW, viewH int) {
	if st == nil || !cfg.WindowChrome.draggable() {
		return
	}
	mw, mh := ModalCellSize(modal)
	st.OriginTop, st.OriginLeft = ClampOverlayOrigin(mw, mh, viewW, viewH, st.OriginTop, st.OriginLeft)
}

// ChromeMouseResult describes stack handling of a mouse message on window chrome.
type ChromeMouseResult struct {
	Consumed bool
	Pop      bool
	// MinimizeToggled is set when the press flipped the layer's Minimized
	// state. Stack callers use this to broadcast OverlayMinimizedMsg and
	// invoke OverlayMinimizer hooks; pure-chrome callers can ignore it
	// (Consumed is still true so they know to swallow the event).
	MinimizeToggled bool
}

// ChromePointerAction describes a pointer event for window chrome handling.
type ChromePointerAction uint8

const (
	ChromePointerPress ChromePointerAction = iota
	ChromePointerRelease
	ChromePointerMotion
)

// HandleChromeMouse updates drag state or requests pop for close. top/left/mw/mh must match painting.
func HandleChromeMouse(msg tea.MouseMsg, cfg OverlayConfig, st *LayerState, modal string, top, left, mw, mh, viewW, viewH int) ChromeMouseResult {
	var action ChromePointerAction
	switch msg.Action {
	case tea.MouseActionPress:
		action = ChromePointerPress
	case tea.MouseActionRelease:
		action = ChromePointerRelease
	case tea.MouseActionMotion:
		action = ChromePointerMotion
	default:
		return ChromeMouseResult{}
	}
	leftBtn := msg.Button == tea.MouseButtonLeft
	return HandleChromePointer(action, leftBtn, msg.X, msg.Y, cfg, st, modal, top, left, mw, mh, viewW, viewH)
}

// HandleChromePointer is the Bubble Tea version-agnostic chrome handler (v2 maps into this).
func HandleChromePointer(action ChromePointerAction, leftButton bool, x, y int, cfg OverlayConfig, st *LayerState, modal string, top, left, mw, mh, viewW, viewH int) ChromeMouseResult {
	wc := cfg.WindowChrome.effective()
	if !wc.Enabled || st == nil {
		return ChromeMouseResult{}
	}
	rx, ry := x-left, y-top
	cw, ch := ContentSizeForLayer(modal, wc, st)
	reg := ComputeChromeRegions(wc, cw, ch)

	switch action {
	case ChromePointerPress:
		if !leftButton {
			return ChromeMouseResult{}
		}
		if wc.showClose() && cellInRect(rx, ry, reg.CloseX, reg.CloseY, reg.CloseW, reg.CloseH) {
			return ChromeMouseResult{Consumed: true, Pop: true}
		}
		if wc.showMinimize() && cellInRect(rx, ry, reg.MinimizeX, reg.MinimizeY, reg.MinimizeW, reg.MinimizeH) {
			st.Minimized = !st.Minimized
			// A pending drag / resize on the modal would survive the
			// state flip and behave nonsensically once the body is gone,
			// so cancel them here.
			st.Dragging = false
			st.Resizing = false
			st.ResizeEdge = ResizeNone
			return ChromeMouseResult{Consumed: true, MinimizeToggled: true}
		}
		// When minimized the body is gone, so resize handles aren't
		// painted and shouldn't accept hits. The minimize button itself
		// is checked above so toggling back to expanded still works.
		if wc.resizable() && !st.Minimized {
			if cellInRect(rx, ry, reg.ResizeCornerX, reg.ResizeCornerY, reg.ResizeCornerW, reg.ResizeCornerH) {
				st.Resizing = true
				st.ResizeEdge = ResizeCorner
				st.ResizeStartX, st.ResizeStartY = x, y
				st.ResizeStartW, st.ResizeStartH = cw, ch
				return ChromeMouseResult{Consumed: true}
			}
			if cellInRect(rx, ry, reg.ResizeRightX, reg.ResizeRightY, reg.ResizeRightW, reg.ResizeRightH) {
				st.Resizing = true
				st.ResizeEdge = ResizeRight
				st.ResizeStartX, st.ResizeStartY = x, y
				st.ResizeStartW, st.ResizeStartH = cw, ch
				return ChromeMouseResult{Consumed: true}
			}
			if cellInRect(rx, ry, reg.ResizeBottomX, reg.ResizeBottomY, reg.ResizeBottomW, reg.ResizeBottomH) {
				st.Resizing = true
				st.ResizeEdge = ResizeBottom
				st.ResizeStartX, st.ResizeStartY = x, y
				st.ResizeStartW, st.ResizeStartH = cw, ch
				return ChromeMouseResult{Consumed: true}
			}
		}
		if wc.draggable() && cellInTabDrag(rx, ry, reg) {
			// Double-click on the tab toggles minimize, mirroring the
			// desktop OS gesture. We only honour it when the minimize
			// affordance is enabled — otherwise users have no visible
			// hint that this gesture exists, and it'd be a surprise.
			//
			// "Same target" is defined as "two presses in the tab drag
			// area within DoubleClickThreshold"; we deliberately don't
			// require the same exact cell so the gesture survives a
			// few cells of cursor jitter between presses.
			now := time.Now()
			if wc.showMinimize() && !st.LastTabPressAt.IsZero() && now.Sub(st.LastTabPressAt) <= DoubleClickThreshold {
				st.Minimized = !st.Minimized
				st.Dragging = false
				st.Resizing = false
				st.ResizeEdge = ResizeNone
				// Clear the press tracker so a third quick click
				// doesn't immediately re-toggle (double-click is a
				// discrete event, not a repeating one).
				st.LastTabPressAt = time.Time{}
				return ChromeMouseResult{Consumed: true, MinimizeToggled: true}
			}
			st.LastTabPressAt = now
			st.Dragging = true
			st.DragOffsetX = x - left
			st.DragOffsetY = y - top
			return ChromeMouseResult{Consumed: true}
		}
	case ChromePointerMotion:
		if st.Resizing && wc.resizable() {
			applyResize(st, wc, x, y, left, top, viewW, viewH)
			return ChromeMouseResult{Consumed: true}
		}
		if st.Dragging && wc.draggable() {
			st.OriginTop = y - st.DragOffsetY
			st.OriginLeft = x - st.DragOffsetX
			st.OriginTop, st.OriginLeft = ClampOverlayOrigin(mw, mh, viewW, viewH, st.OriginTop, st.OriginLeft)
			return ChromeMouseResult{Consumed: true}
		}
	case ChromePointerRelease:
		if leftButton && st.Resizing {
			st.Resizing = false
			st.ResizeEdge = ResizeNone
			return ChromeMouseResult{Consumed: true}
		}
		if leftButton && st.Dragging {
			st.Dragging = false
			return ChromeMouseResult{Consumed: true}
		}
	}
	return ChromeMouseResult{}
}

func contentSizeLimits(wc WindowChrome, modalTop, modalLeft, viewW, viewH int) (maxW, maxH int) {
	maxW = max(wc.minWidth(), viewW-modalLeft-2)
	chromeRows := 4
	if wc.tabOnBorder() {
		chromeRows = 3
	}
	maxH = max(wc.minHeight(), viewH-modalTop-wc.tabOffsetTop()-chromeRows)
	return maxW, maxH
}

func nudgeOrigin(st *LayerState, dTop, dLeft, mw, mh, viewW, viewH int) {
	if !st.OriginInitialized {
		st.OriginTop, st.OriginLeft = 0, 0
		st.OriginInitialized = true
	}
	st.OriginTop += dTop
	st.OriginLeft += dLeft
	st.OriginTop, st.OriginLeft = ClampOverlayOrigin(mw, mh, viewW, viewH, st.OriginTop, st.OriginLeft)
}

func nudgeContentSize(st *LayerState, wc WindowChrome, dW, dH, modalTop, modalLeft, viewW, viewH int) {
	maxW, maxH := contentSizeLimits(wc, modalTop, modalLeft, viewW, viewH)
	if dW != 0 {
		st.ContentWidth = clampInt(st.ContentWidth+dW, wc.minWidth(), maxW)
	}
	if dH != 0 {
		st.ContentHeight = clampInt(st.ContentHeight+dH, wc.minHeight(), maxH)
	}
	st.ContentSizeInitialized = true
}

func applyResize(st *LayerState, wc WindowChrome, x, y, modalLeft, modalTop, viewW, viewH int) {
	dx := x - st.ResizeStartX
	dy := y - st.ResizeStartY
	maxW, maxH := contentSizeLimits(wc, modalTop, modalLeft, viewW, viewH)

	switch st.ResizeEdge {
	case ResizeRight:
		st.ContentWidth = clampInt(st.ResizeStartW+dx, wc.minWidth(), maxW)
	case ResizeBottom:
		st.ContentHeight = clampInt(st.ResizeStartH+dy, wc.minHeight(), maxH)
	case ResizeCorner:
		st.ContentWidth = clampInt(st.ResizeStartW+dx, wc.minWidth(), maxW)
		st.ContentHeight = clampInt(st.ResizeStartH+dy, wc.minHeight(), maxH)
	}
	st.ContentSizeInitialized = true
}

// ChromeKeyResult describes stack handling of a key message on window chrome.
type ChromeKeyResult struct {
	Consumed bool
	Pop      bool
}

// HandleChromeKey handles Alt+arrow move and Alt+Shift+arrow resize when Keyboard is enabled.
func HandleChromeKey(msg tea.KeyMsg, cfg OverlayConfig, st *LayerState, modal string, top, left, mw, mh, viewW, viewH int) ChromeKeyResult {
	return HandleChromeKeyString(msg.String(), cfg, st, modal, top, left, mw, mh, viewW, viewH)
}

// chromeKeyAction classifies a normalized key binding for window chrome.
type chromeKeyAction uint8

const (
	chromeKeyNone chromeKeyAction = iota
	chromeKeyMoveRight
	chromeKeyMoveLeft
	chromeKeyMoveDown
	chromeKeyMoveUp
	chromeKeyGrowRight
	chromeKeyGrowLeft
	chromeKeyGrowDown
	chromeKeyGrowUp
)

func classifyChromeKey(key string) chromeKeyAction {
	switch key {
	case "alt+right", "alt+ctrl+right", "ctrl+alt+right":
		return chromeKeyMoveRight
	case "alt+left", "alt+ctrl+left", "ctrl+alt+left":
		return chromeKeyMoveLeft
	case "alt+down", "alt+ctrl+down", "ctrl+alt+down":
		return chromeKeyMoveDown
	case "alt+up", "alt+ctrl+up", "ctrl+alt+up":
		return chromeKeyMoveUp
	case "alt+shift+right", "alt+ctrl+shift+right", "ctrl+alt+shift+right":
		return chromeKeyGrowRight
	case "alt+shift+left", "alt+ctrl+shift+left", "ctrl+alt+shift+left":
		return chromeKeyGrowLeft
	case "alt+shift+down", "alt+ctrl+shift+down", "ctrl+alt+shift+down":
		return chromeKeyGrowDown
	case "alt+shift+up", "alt+ctrl+shift+up", "ctrl+alt+shift+up":
		return chromeKeyGrowUp
	default:
		return chromeKeyNone
	}
}

// HandleChromeKeyString is the key-string entry point (used by Bubble Tea v2 KeyPressMsg).
func HandleChromeKeyString(key string, cfg OverlayConfig, st *LayerState, modal string, top, left, mw, mh, viewW, viewH int) ChromeKeyResult {
	wc := cfg.WindowChrome.effective()
	if !wc.keyboard() || st == nil {
		return ChromeKeyResult{}
	}
	step := wc.keyStep()
	switch classifyChromeKey(key) {
	case chromeKeyMoveRight:
		if wc.draggable() {
			nudgeOrigin(st, 0, step, mw, mh, viewW, viewH)
			return ChromeKeyResult{Consumed: true}
		}
	case chromeKeyMoveLeft:
		if wc.draggable() {
			nudgeOrigin(st, 0, -step, mw, mh, viewW, viewH)
			return ChromeKeyResult{Consumed: true}
		}
	case chromeKeyMoveDown:
		if wc.draggable() {
			nudgeOrigin(st, step, 0, mw, mh, viewW, viewH)
			return ChromeKeyResult{Consumed: true}
		}
	case chromeKeyMoveUp:
		if wc.draggable() {
			nudgeOrigin(st, -step, 0, mw, mh, viewW, viewH)
			return ChromeKeyResult{Consumed: true}
		}
	case chromeKeyGrowRight:
		if wc.resizable() {
			nudgeContentSize(st, wc, step, 0, top, left, viewW, viewH)
			return ChromeKeyResult{Consumed: true}
		}
	case chromeKeyGrowLeft:
		if wc.resizable() {
			nudgeContentSize(st, wc, -step, 0, top, left, viewW, viewH)
			return ChromeKeyResult{Consumed: true}
		}
	case chromeKeyGrowDown:
		if wc.resizable() {
			nudgeContentSize(st, wc, 0, step, top, left, viewW, viewH)
			return ChromeKeyResult{Consumed: true}
		}
	case chromeKeyGrowUp:
		if wc.resizable() {
			nudgeContentSize(st, wc, 0, -step, top, left, viewW, viewH)
			return ChromeKeyResult{Consumed: true}
		}
	}
	return ChromeKeyResult{}
}

func clampInt(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func cellInRect(x, y, left, top, w, h int) bool {
	if w <= 0 || h <= 0 {
		return false
	}
	return x >= left && x < left+w && y >= top && y < top+h
}

func cellInTabDrag(x, y int, reg ChromeRegions) bool {
	if !cellInRect(x, y, reg.TabLeft, reg.TabTop, reg.TabW, reg.TabH) {
		return false
	}
	if reg.CloseW > 0 && cellInRect(x, y, reg.CloseX, reg.CloseY, reg.CloseW, reg.CloseH) {
		return false
	}
	return true
}
