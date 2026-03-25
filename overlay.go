// Package overlay provides OverlayView for compositing a modal over a main view
// so only the modal rectangle is replaced; the rest of the main view stays visible.

package overlay

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/cellbuf"
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
	mainLines := strings.Split(mainView, "\n")
	modalLines := strings.Split(modalView, "\n")
	if len(modalLines) == 0 {
		var out []string
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
	// Clamp so modal fits
	if left+modalW > viewWidth {
		left = max(0, viewWidth-modalW)
	}
	left = max(left, 0)
	if top+modalH > viewHeight {
		top = max(0, viewHeight-modalH)
	}
	top = max(top, 0)

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
		combined := overlayLine(mainLine, modalLine, left, modalW, viewWidth)
		out = append(out, combined)
	}
	return strings.Join(out, "\n")
}

// overlayLine returns mainLine with modalLine overlaid at column left for modalW cells.
// When mainLine has fewer than left cells (e.g. main view has fewer rows), prefix is
// padded so the modal stays aligned at column left.
func overlayLine(mainLine, modalLine string, left, modalW, viewWidth int) string {
	prefix := prefixCells(mainLine, left)
	if w := widthCells(prefix); w < left {
		prefix += strings.Repeat(" ", left-w)
	}
	suffix := skipCells(mainLine, left+modalW)
	st, lk := penAtCellOffset(mainLine, left+modalW)
	resume := ansiResumeAfterOverlay(st, lk)
	// Clear pen after prefix so the modal is not drawn on top of the main line’s bg/fg.
	beforeModal := ansi.ResetStyle + ansi.ResetHyperlink()
	line := prefix + beforeModal + modalLine + resume + suffix
	return padOrTruncateLine(line, viewWidth)
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

// ansiResumeAfterOverlay resets terminal attributes (as after a typical lipgloss modal),
// then reapplies pen so the following suffix bytes render like the original stream.
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

// prefixCells returns the prefix of s that spans at most n display cells (ANSI preserved).
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

// skipCells returns the substring of s starting at the first display cell index n (ANSI preserved).
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
