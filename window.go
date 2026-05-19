package overlay

import (
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// CloseButtonGlyph is the close control label inside the tab.
const CloseButtonGlyph = "[x]"

// CloseButtonWidth is the display width of CloseButtonGlyph for hit-testing.
const CloseButtonWidth = 3

// DefaultChromeMaskRune is the pass-through padding rune for WindowChrome auto-wrap
// (U+FFFC OBJECT REPLACEMENT). It is unlikely to appear in modal content. Set
// WindowChrome.ChromeMaskRune to override if your content might use this character.
const DefaultChromeMaskRune = '\ufffc'

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
	Enabled         bool
	Title           string
	ShowCloseButton bool
	AutoWrap        bool
	TitleBarHeight  int // legacy; tab layout uses TabOffsetTop + tab rows when zero
	Draggable       bool
	TabBackground   string // lipgloss color for tab fill (default muted "238")
	TabForeground   string // lipgloss color for tab text
	TabBorder       string // lipgloss color for tab border runes
	TabOffsetTop    int    // rows above tab (default 1)
	TabOffsetLeft   int    // columns left of tab (default 0)
	ChromeMaskRune  rune   // pass-through padding in auto-wrap chrome; 0 uses DefaultChromeMaskRune
	Resizable       bool   // drag right/bottom edges (and corner) to resize content
	Keyboard        bool   // Alt+arrow move, Alt+Shift+arrow resize
	KeyStep         int    // cells per keypress when Keyboard is enabled (default 1)
	CenterContent   bool   // keep content centered in the body when Resizable
	ContentPadTop   int    // blank lines pinned to the top of the resizable body
	MinWidth        int    // minimum content width when Resizable
	MinHeight       int    // minimum content height (lines) when Resizable
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

// ChromeRegions describes tab, close, and resize hit areas relative to the modal's top-left cell.
type ChromeRegions struct {
	TabTop, TabLeft, TabW, TabH int
	CloseX, CloseY, CloseW, CloseH int
	ResizeRightX, ResizeRightY, ResizeRightW, ResizeRightH int
	ResizeBottomX, ResizeBottomY, ResizeBottomW, ResizeBottomH int
	ResizeCornerX, ResizeCornerY, ResizeCornerW, ResizeCornerH int
}

// ComputeChromeRegions returns hit rectangles for a framed modal's content size.
func ComputeChromeRegions(wc WindowChrome, contentW, contentH int) ChromeRegions {
	wc = wc.effective()
	ot, ol := wc.tabOffsetTop(), wc.tabOffsetLeft()
	inner := tabInnerText(wc.Title, wc.showClose(), contentW)
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
		inner := tabInnerText(wc.Title, true, contentW)
		reg.CloseX = tabLeft + 1 + lipgloss.Width(inner)-CloseButtonWidth
		reg.CloseY = tabTop + 1
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
	return renderTabFrame(content, cw, contentHeight(content), wc)
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

func tabInnerText(title string, showClose bool, contentWidth int) string {
	if showClose {
		prefix := " "
		suffix := " " + CloseButtonGlyph + " "
		maxTitle := max(0, contentWidth-lipgloss.Width(prefix+suffix))
		t := truncateWidth(title, maxTitle)
		return prefix + t + suffix
	}
	t := tabTitleText(title, contentWidth)
	return " " + t + " "
}

func innerWidth(inner string) int {
	return lipgloss.Width(inner)
}

func renderTabFrame(content string, cw, ch int, wc WindowChrome) string {
	ot, ol := wc.tabOffsetTop(), wc.tabOffsetLeft()
	mask := wc.chromeMaskRune()
	inner := tabInnerText(wc.Title, wc.showClose(), cw)
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

	var out []string

	for range ot {
		out = append(out, maskFill(modalW, mask))
	}

	tabTop := borderSt.Render("┌" + strings.Repeat("─", iw) + "┐")
	tabRow := borderSt.Render("│") + tabSt.Render(inner) + borderSt.Render("│")
	out = append(out, composeChromeLine(maskFill(ol, mask), tabTop, maskFill(max(0, modalW-ol-tabW), mask)))

	if wc.resizable() && wc.tabOnBorder() {
		out = append(out, renderTabOnBorderLine(ol, iw, cw, inner, tabSt, borderSt, mask, modalW))
	} else {
		out = append(out, composeChromeLine(maskFill(ol, mask), tabRow, maskFill(max(0, modalW-ol-tabW), mask)))
	}

	if wc.resizable() {
		lines := fitContentLines(content, cw, ch, wc.centerContent(), wc.contentPadTop())
		if !wc.tabOnBorder() {
			topBox := borderSt.Render("┌" + strings.Repeat("─", cw) + "┐")
			out = append(out, padLineWithMask(topBox, modalW, mask))
		}
		for _, line := range lines {
			row := borderSt.Render("│") + line + borderSt.Render("│")
			out = append(out, padLineWithMask(row, modalW, mask))
		}
		bottomBox := borderSt.Render("└" + strings.Repeat("─", cw) + "┘")
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
		prefix = borderSt.Render("┌" + strings.Repeat("─", ol-1))
	}
	tabPart := borderSt.Render("├") + tabSt.Render(inner) + borderSt.Render("┴")
	mid := prefix + tabPart
	used := lipgloss.Width(mid)
	hlineLen := boxW - used - 1 // remaining ─ before ┐
	if hlineLen < 0 {
		hlineLen = 0
	}
	right := borderSt.Render(strings.Repeat("─", hlineLen) + "┐")
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
	OriginTop, OriginLeft int
	OriginInitialized     bool
	Dragging              bool
	DragOffsetX           int
	DragOffsetY           int
	ContentWidth, ContentHeight int
	ContentSizeInitialized    bool
	Resizing                  bool
	ResizeEdge                ResizeEdge
	ResizeStartX, ResizeStartY int
	ResizeStartW, ResizeStartH int
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
	if wc.resizable() {
		return renderTabFrame(modelView, cw, ch, wc)
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
		if wc.resizable() {
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
