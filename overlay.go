// Package overlay provides OverlayView for compositing a modal over a main view
// so only the modal rectangle is replaced; the rest of the main view stays visible.
//
// OverlayStack composes multiple Bubble Tea models with optional dimming,
// declarative placement (OverlayConfig), Escape / click-outside dismissal, and
// focus-trapped updates (see OverlayStack.MainReceivesKeyMsg and FocusTrap).
//
// Hosts that forward mouse events should use ClampOverlayOrigin or Placement.ClampedOrigin
// so hit-testing matches OverlayView; see README “Consumer integration”.
//
// For Bubble Tea v2 (charm.land/bubbletea/v2), use import path
// github.com/madicen/bubble-overlay/v2 (package overlayv2). See docs/ADR-v2-bridge.md.

package overlay

import (
	"strings"
	"unicode/utf8"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/cellbuf"
)

// overlayMergeKind selects how modal cells replace main cells in the cellbuf compositor.
type overlayMergeKind uint8

const (
	mergeOpaque overlayMergeKind = iota
	mergeTransparentSpaces
	mergeMaskRune
)

// OverlayView composites modalView on top of mainView. Only the rectangle at (top, left)
// with the modal's size is replaced; all other cells show the main view. Returns a single
// string with viewHeight lines, each viewWidth cells wide (padding/truncation as needed).
//
// Main and modal strings may contain ANSI (e.g. from lipgloss); overlay uses display-cell
// width (grapheme-aware, matching lipgloss) so alignment is correct. After the modal,
// graphics state that originated under the modal is re-applied so background colors and
// other SGR attributes still apply to the visible tail of each line. A full SGR reset
// (and hyperlink reset) is inserted immediately before the modal so the main line’s
// active pen does not bleed into the first cells of the modal.
func OverlayView(mainView, modalView string, viewWidth, viewHeight, top, left int) string {
	return overlayViewInternal(mainView, modalView, viewWidth, viewHeight, top, left, mergeOpaque, 0)
}

// OverlayViewWithTransparency is like OverlayView but cells in the modal that are ASCII
// space (' ') are treated as transparent: the main view shows through at those positions.
// Non-space modal cells (including styled spaces from lipgloss that carry a background)
// still replace the main view.
func OverlayViewWithTransparency(mainView, modalView string, viewWidth, viewHeight, top, left int) string {
	return overlayViewInternal(mainView, modalView, viewWidth, viewHeight, top, left, mergeTransparentSpaces, 0)
}

// OverlayViewWithMask is like OverlayView but treats any cell whose rune equals maskRune
// as transparent (pass-through to the main view). Use 0 for maskRune to behave like OverlayView.
func OverlayViewWithMask(mainView, modalView string, viewWidth, viewHeight, top, left int, maskRune rune) string {
	if maskRune == 0 {
		return OverlayView(mainView, modalView, viewWidth, viewHeight, top, left)
	}
	return overlayViewInternal(mainView, modalView, viewWidth, viewHeight, top, left, mergeMaskRune, maskRune)
}

// ClampOverlayOrigin applies the same origin adjustment as OverlayView: if the modal rectangle
// would extend past the viewport edge, top and/or left are shifted so the rectangle fits; then
// negative coordinates are clamped to zero.
//
// modalW and modalH must match how OverlayView measures that modal string (see ModalCellSize).
func ClampOverlayOrigin(modalW, modalH, viewW, viewH, top, left int) (int, int) {
	if left+modalW > viewW {
		left = max(0, viewW-modalW)
	}
	left = max(left, 0)
	if top+modalH > viewH {
		top = max(0, viewH-modalH)
	}
	top = max(top, 0)
	return top, left
}

func overlayViewInternal(mainView, modalView string, viewWidth, viewHeight, top, left int, kind overlayMergeKind, maskRune rune) string {
	mainLines := strings.Split(mainView, "\n")
	modalLines := strings.Split(modalView, "\n")
	if len(modalLines) == 0 {
		out := make([]string, 0, viewHeight)
		for row := range viewHeight {
			line := ""
			if row < len(mainLines) {
				line = mainLines[row]
			}
			out = append(out, padOrTruncateLine(line, viewWidth))
		}
		return strings.Join(out, "\n")
	}
	modalH := len(modalLines)
	modalW := 0
	for _, l := range modalLines {
		if w := lipgloss.Width(l); w > modalW {
			modalW = w
		}
	}
	top, left = ClampOverlayOrigin(modalW, modalH, viewWidth, viewHeight, top, left)

	var out []string
	for row := range viewHeight {
		mainLine := ""
		if row < len(mainLines) {
			mainLine = mainLines[row]
		}
		if row < top || row >= top+modalH {
			out = append(out, padOrTruncateLine(mainLine, viewWidth))
			continue
		}
		modalLine := modalLines[row-top]
		combined := overlayLine(mainLine, modalLine, left, modalW, viewWidth, kind, maskRune)
		out = append(out, combined)
	}
	return strings.Join(out, "\n")
}

// OverlayViewInCenter centers modalView in a viewport of viewWidth×viewHeight and composites with OverlayView.
// The viewport may be the full terminal, a tab/content region, or any rectangle you pass to OverlayView—this is
// the general “center in viewport” helper, not full-screen-only.
//
// When the background is not the entire terminal, pass the same width and height you use for that region.
// If the region matches the main view’s cell bounds, use ModalCellSize(mainView) or OverlayViewInCenterInMain.
//
// Centering uses ModalCellSize(modalView), matching OverlayView’s internal modal measurement (not lipgloss.Size).
func OverlayViewInCenter(mainView, modalView string, viewWidth, viewHeight int) string {
	modalW, modalH := ModalCellSize(modalView)
	top := (viewHeight - modalH) / 2
	left := (viewWidth - modalW) / 2
	return OverlayView(mainView, modalView, viewWidth, viewHeight, top, left)
}

// OverlayViewInCenterWithTransparency is like OverlayViewInCenter but uses OverlayViewWithTransparency.
func OverlayViewInCenterWithTransparency(mainView, modalView string, viewWidth, viewHeight int) string {
	modalW, modalH := ModalCellSize(modalView)
	top := (viewHeight - modalH) / 2
	left := (viewWidth - modalW) / 2
	return OverlayViewWithTransparency(mainView, modalView, viewWidth, viewHeight, top, left)
}

// OverlayViewInCenterWithMask is like OverlayViewInCenter but uses OverlayViewWithMask.
func OverlayViewInCenterWithMask(mainView, modalView string, viewWidth, viewHeight int, maskRune rune) string {
	modalW, modalH := ModalCellSize(modalView)
	top := (viewHeight - modalH) / 2
	left := (viewWidth - modalW) / 2
	return OverlayViewWithMask(mainView, modalView, viewWidth, viewHeight, top, left, maskRune)
}

// OverlayViewInCenterInMain derives the viewport size from ModalCellSize(mainView) and centers modalView over mainView.
// Use when the overlay applies to exactly the main string’s cell bounds (e.g. an inner panel or log region).
func OverlayViewInCenterInMain(mainView, modalView string) string {
	viewW, viewH := ModalCellSize(mainView)
	return OverlayViewInCenter(mainView, modalView, viewW, viewH)
}

// OverlayViewInCenterWithOffset centers modalView, adds deltaTop and deltaLeft (e.g. nudge a loading banner upward),
// then composites. Overflow clamping matches OverlayView (applied inside OverlayView).
func OverlayViewInCenterWithOffset(mainView, modalView string, viewWidth, viewHeight, deltaTop, deltaLeft int) string {
	modalW, modalH := ModalCellSize(modalView)
	top := (viewHeight-modalH)/2 + deltaTop
	left := (viewWidth-modalW)/2 + deltaLeft
	return OverlayView(mainView, modalView, viewWidth, viewHeight, top, left)
}

// OverlayViewInCenterWithOffsetWithTransparency is like OverlayViewInCenterWithOffset but uses OverlayViewWithTransparency.
func OverlayViewInCenterWithOffsetWithTransparency(mainView, modalView string, viewWidth, viewHeight, deltaTop, deltaLeft int) string {
	modalW, modalH := ModalCellSize(modalView)
	top := (viewHeight-modalH)/2 + deltaTop
	left := (viewWidth-modalW)/2 + deltaLeft
	return OverlayViewWithTransparency(mainView, modalView, viewWidth, viewHeight, top, left)
}

// OverlayViewInCenterWithOffsetWithMask is like OverlayViewInCenterWithOffset but uses OverlayViewWithMask.
func OverlayViewInCenterWithOffsetWithMask(mainView, modalView string, viewWidth, viewHeight, deltaTop, deltaLeft int, maskRune rune) string {
	modalW, modalH := ModalCellSize(modalView)
	top := (viewHeight-modalH)/2 + deltaTop
	left := (viewWidth-modalW)/2 + deltaLeft
	return OverlayViewWithMask(mainView, modalView, viewWidth, viewHeight, top, left, maskRune)
}

// OverlayViewAtPoint composites modalView with its top-left anchored at (anchorTop, anchorLeft) after the same
// overflow clamp as OverlayView. Use Bubble Tea v1 mouse coordinates as anchorTop = msg.Y (row) and anchorLeft = msg.X (column).
func OverlayViewAtPoint(mainView, modalView string, viewWidth, viewHeight, anchorTop, anchorLeft int) string {
	top, left := ClampOverlayOriginAtPoint(modalView, viewWidth, viewHeight, anchorTop, anchorLeft)
	return OverlayView(mainView, modalView, viewWidth, viewHeight, top, left)
}

// OverlayViewAtPointWithTransparency is like OverlayViewAtPoint but uses OverlayViewWithTransparency.
func OverlayViewAtPointWithTransparency(mainView, modalView string, viewWidth, viewHeight, anchorTop, anchorLeft int) string {
	top, left := ClampOverlayOriginAtPoint(modalView, viewWidth, viewHeight, anchorTop, anchorLeft)
	return OverlayViewWithTransparency(mainView, modalView, viewWidth, viewHeight, top, left)
}

// OverlayViewAtPointWithMask is like OverlayViewAtPoint but uses OverlayViewWithMask.
func OverlayViewAtPointWithMask(mainView, modalView string, viewWidth, viewHeight, anchorTop, anchorLeft int, maskRune rune) string {
	top, left := ClampOverlayOriginAtPoint(modalView, viewWidth, viewHeight, anchorTop, anchorLeft)
	return OverlayViewWithMask(mainView, modalView, viewWidth, viewHeight, top, left, maskRune)
}

func overlayLine(mainLine, modalLine string, left, modalW, viewWidth int, kind overlayMergeKind, maskRune rune) string {
	if kind == mergeOpaque {
		prefix := prefixCells(mainLine, left)
		if w := widthCells(prefix); w < left {
			prefix += strings.Repeat(" ", left-w)
		}
		suffix := skipCells(mainLine, left+modalW)
		st, lk := penAtCellOffset(mainLine, left+modalW)
		resume := ansiResumeAfterOverlay(st, lk)
		beforeModal := ansi.ResetStyle + ansi.ResetHyperlink()
		line := prefix + beforeModal + modalLine + resume + suffix
		return padOrTruncateLine(line, viewWidth)
	}

	buf := cellbuf.NewBuffer(viewWidth, 1)
	setString(buf, 0, 0, mainLine)

	mBuf := cellbuf.NewBuffer(modalW, 1)
	setString(mBuf, 0, 0, modalLine)

	for x := 0; x < modalW; x++ {
		if left+x >= viewWidth {
			break
		}
		mCell := mBuf.Cell(x, 0)
		if mCell == nil {
			continue
		}
		switch kind {
		case mergeTransparentSpaces:
			if mCell.Rune == ' ' {
				continue
			}
		case mergeMaskRune:
			if mCell.Rune == maskRune {
				continue
			}
		}
		buf.SetCell(left+x, 0, mCell)
	}

	return renderLine(buf, viewWidth)
}

func renderLine(buf *cellbuf.Buffer, width int) string {
	var b strings.Builder
	var curStyle cellbuf.Style
	var curLink cellbuf.Link

	b.WriteString(ansi.ResetStyle)
	b.WriteString(ansi.ResetHyperlink())

	var skip int
	for x := 0; x < width; x++ {
		if skip > 0 {
			skip--
			continue
		}
		c := buf.Cell(x, 0)
		if c == nil {
			continue
		}
		if c.Style != curStyle {
			if c.Style.Empty() {
				b.WriteString(ansi.ResetStyle)
			} else {
				// DiffSequence updates the active pen from curStyle so omitted attributes
				// (e.g. no Bg on modal lipgloss text) reset attributes left over from the
				// previous cell — Sequence() alone does not emit \033[49m, so main-layer
				// backgrounds could bleed across overlay merges (issue seen when the
				// speech bubble overlaps the styled status row).
				b.WriteString(c.Style.DiffSequence(curStyle))
			}
			curStyle = c.Style
		}
		if c.Link != curLink {
			if c.Link.Empty() {
				b.WriteString(ansi.ResetHyperlink())
			} else {
				b.WriteString(ansi.SetHyperlink(c.Link.URL, c.Link.Params))
			}
			curLink = c.Link
		}
		if c.Rune != 0 {
			b.WriteRune(c.Rune)
			w := ansi.StringWidth(string(c.Rune))
			if w > 1 {
				skip = w - 1
			}
		} else {
			b.WriteByte(' ')
		}
	}

	b.WriteString(ansi.ResetStyle)
	b.WriteString(ansi.ResetHyperlink())
	return b.String()
}

func setString(buf *cellbuf.Buffer, x, y int, s string) {
	var st cellbuf.Style
	var lk cellbuf.Link
	var state byte
	p := ansi.GetParser()
	defer ansi.PutParser(p)
	pos := 0
	curX := x
	for pos < len(s) {
		seq, width, nRead, newState := ansi.DecodeSequence(s[pos:], state, p)
		state = newState
		if width == 0 {
			if ansi.HasCsiPrefix(seq) && byte(p.Command()&0xff) == 'm' {
				cellbuf.ReadStyle(p.Params(), &st)
			} else if ansi.HasOscPrefix(seq) && p.Command() == 8 {
				cellbuf.ReadLink(p.Data(), &lk)
			}
			pos += nRead
			continue
		}
		if curX < buf.Width() {
			r, _ := utf8.DecodeRuneInString(seq)
			buf.SetCell(curX, y, &cellbuf.Cell{
				Rune:  r,
				Style: st,
				Link:  lk,
			})
			for i := 1; i < width && curX+i < buf.Width(); i++ {
				buf.SetCell(curX+i, y, &cellbuf.Cell{Rune: 0, Style: st, Link: lk})
			}
		}
		curX += width
		pos += nRead
	}
}

// penAtCellOffset returns the SGR + hyperlink pen after consuming n display cells of s
// (same cell boundaries as skipCells(s, n)).
func penAtCellOffset(s string, n int) (cellbuf.Style, cellbuf.Link) {
	var st cellbuf.Style
	var lk cellbuf.Link
	var cellCount int
	var state byte
	p := ansi.GetParser()
	defer ansi.PutParser(p)
	pos := 0
	for pos < len(s) {
		seq, width, nRead, newState := ansi.DecodeSequence(s[pos:], state, p)
		state = newState
		if width == 0 {
			if ansi.HasCsiPrefix(seq) && byte(p.Command()&0xff) == 'm' {
				cellbuf.ReadStyle(p.Params(), &st)
			} else if ansi.HasOscPrefix(seq) && p.Command() == 8 {
				cellbuf.ReadLink(p.Data(), &lk)
			}
			pos += nRead
			continue
		}
		if cellCount >= n {
			return st, lk
		}
		cellCount += width
		pos += nRead
	}
	return st, lk
}

func ansiResumeAfterOverlay(st cellbuf.Style, lk cellbuf.Link) string {
	var b strings.Builder
	b.WriteString(ansi.ResetStyle)
	b.WriteString(ansi.ResetHyperlink())
	if !lk.Empty() {
		b.WriteString(ansi.SetHyperlink(lk.URL, lk.Params))
	}
	if !st.Empty() {
		b.WriteString(st.Sequence())
	}
	return b.String()
}

func prefixCells(s string, n int) string {
	if n <= 0 {
		return ""
	}
	var cellCount int
	var state byte
	p := ansi.GetParser()
	defer ansi.PutParser(p)
	pos := 0
	for pos < len(s) {
		_, width, nRead, newState := ansi.DecodeSequence(s[pos:], state, p)
		state = newState
		if width == 0 {
			pos += nRead
			continue
		}
		if cellCount+width > n {
			break
		}
		cellCount += width
		pos += nRead
	}
	return s[:pos]
}

// skipCells returns the substring of s starting after the first n display cells (ANSI preserved).
func skipCells(s string, n int) string {
	var cellCount int
	var state byte
	p := ansi.GetParser()
	defer ansi.PutParser(p)
	pos := 0
	for pos < len(s) {
		_, width, nRead, newState := ansi.DecodeSequence(s[pos:], state, p)
		state = newState
		if width == 0 {
			pos += nRead
			continue
		}
		if cellCount >= n {
			return s[pos:]
		}
		cellCount += width
		pos += nRead
	}
	return ""
}

func widthCells(s string) int {
	return ansi.StringWidth(s)
}

func padOrTruncateLine(line string, viewWidth int) string {
	w := widthCells(line)
	if w < viewWidth {
		return line + strings.Repeat(" ", viewWidth-w)
	}
	if w > viewWidth {
		return prefixCells(line, viewWidth)
	}
	return line
}
