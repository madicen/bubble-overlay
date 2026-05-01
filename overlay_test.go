package overlay

import (
	"fmt"
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

func TestOverlayView_replacesOnlyModalRect(t *testing.T) {
	mainLines := []string{
		"LLLLLLLLLLLLLLLLLLLLLLLLLLLLLL",
		"LLLLLLLLLLLLLLLLLLLLLLLLLLLLLL",
		"LLLLLLLLLLLLLLLLLLLLLLLLLLLLLL",
		"LLLLLLLLLLLLLLLLLLLLLLLLLLLLLL",
		"LLLLLLLLLLLLLLLLLLLLLLLLLLLLLL",
		"LLLLLLLLLLLLLLLLLLLLLLLLLLLLLL",
		"LLLLLLLLLLLLLLLLLLLLLLLLLLLLLL",
		"LLLLLLLLLLLLLLLLLLLLLLLLLLLLLL",
		"LLLLLLLLLLLLLLLLLLLLLLLLLLLLLL",
		"LLLLLLLLLLLLLLLLLLLLLLLLLLLLLL",
	}
	mainView := strings.Join(mainLines, "\n")
	modalLines := []string{
		"MMMMMMMMMM",
		"MMMMMMMMMM",
		"MMMMMMMMMM",
		"MMMMMMMMMM",
	}
	modalView := strings.Join(modalLines, "\n")

	viewWidth, viewHeight := 30, 10
	top, left := 3, 10
	out := OverlayView(mainView, modalView, viewWidth, viewHeight, top, left)
	lines := strings.Split(out, "\n")

	if len(lines) != viewHeight {
		t.Fatalf("got %d lines, want %d", len(lines), viewHeight)
	}
	for row := 0; row < top; row++ {
		if !strings.Contains(lines[row], "L") || strings.Contains(lines[row], "M") {
			t.Errorf("row %d: should be main only, got %q", row, lines[row])
		}
	}
	for row := top; row < top+4; row++ {
		line := ansi.Strip(lines[row])
		leftPart := line[:10]
		midPart := line[10:20]
		rightPart := line[20:30]
		if !strings.Contains(leftPart, "L") {
			t.Errorf("row %d: left 10 cols should be main, got %q", row, leftPart)
		}
		if !strings.Contains(midPart, "M") {
			t.Errorf("row %d: middle 10 cols should be modal, got %q", row, midPart)
		}
		if !strings.Contains(rightPart, "L") {
			t.Errorf("row %d: right 10 cols should be main, got %q", row, rightPart)
		}
	}
	for row := top + 4; row < viewHeight; row++ {
		if !strings.Contains(lines[row], "L") || strings.Contains(lines[row], "M") {
			t.Errorf("row %d: should be main only, got %q", row, lines[row])
		}
	}
}

func TestOverlayView_modalReplacesRegion(t *testing.T) {
	mainLines := []string{
		"AAAAABBBBBAAAAABBBBB",
		"AAAAABBBBBAAAAABBBBB",
		"AAAAABBBBBAAAAABBBBB",
		"AAAAABBBBBAAAAABBBBB",
		"AAAAABBBBBAAAAABBBBB",
	}
	mainView := strings.Join(mainLines, "\n")
	modalLines := []string{
		"MM  MM    ",
		"  MMMM    ",
		"MM    MM  ",
	}
	modalView := strings.Join(modalLines, "\n")

	out := OverlayView(mainView, modalView, 20, 5, 1, 5)
	lines := strings.Split(out, "\n")

	row1 := ansi.Strip(lines[1])
	if !strings.Contains(row1[:5], "A") {
		t.Errorf("row 1 left: want A from main, got %q", row1[:5])
	}
	if !strings.Contains(row1[5:15], "M") {
		t.Errorf("row 1 overlay region: want modal content, got %q", row1[5:15])
	}
	if !strings.Contains(row1[15:], "B") {
		t.Errorf("row 1 right margin: want B from main, got %q", row1[15:])
	}
}

func TestOverlayViewWithTransparency_modalRow(t *testing.T) {
	mainLines := []string{
		"AAAAABBBBBAAAAABBBBB",
		"AAAAABBBBBAAAAABBBBB",
		"AAAAABBBBBAAAAABBBBB",
		"AAAAABBBBBAAAAABBBBB",
		"AAAAABBBBBAAAAABBBBB",
	}
	mainView := strings.Join(mainLines, "\n")
	modalLines := []string{
		"MM  MM    ",
		"  MMMM    ",
		"MM    MM  ",
	}
	modalView := strings.Join(modalLines, "\n")

	out := OverlayViewWithTransparency(mainView, modalView, 20, 5, 1, 5)
	lines := strings.Split(out, "\n")
	row1 := ansi.Strip(lines[1])
	wantMid := "MMBBMMAAAA"
	if gotMid := row1[5:15]; gotMid != wantMid {
		t.Errorf("row 1 overlay region: want %q, got %q", wantMid, gotMid)
	}
}

func TestOverlayView_opaqueIsTrulyOpaque(t *testing.T) {
	mainView := "AAAAAAAAAA"
	modalView := "  MM  "
	out := OverlayView(mainView, modalView, 10, 1, 0, 2)
	stripped := ansi.Strip(out)
	want := "AA  MM  AA"
	if stripped != want {
		t.Errorf("opaque overlay: want %q, got %q", want, stripped)
	}
}

func TestOverlayViewWithMask_resetsBackgroundWhenModalOmitsBG(t *testing.T) {
	// Explicit SGR: lipgloss may omit colors without a TTY (CI), which would skip this path entirely.
	mainRow := "\x1b[48;5;236m\x1b[38;5;252m" + strings.Repeat("x", 40) + "\x1b[0m"
	modal := "\x1b[38;5;201mHello" // foreground only, no explicit background
	out := OverlayViewWithMask(mainRow, modal, 40, 1, 0, 10, '\ufffc')
	// renderLine uses DiffSequence so default background is set; x/ansi may emit 49 alone or with fg (e.g. "...;49m").
	if !strings.Contains(out, "\x1b[49m") && !strings.Contains(out, ";49m") {
		t.Fatalf("expected SGR 49 (default background) when painting FG-only modal after main BG (got %q)", out)
	}
}

func TestOverlayView_resetsPenBeforeModal(t *testing.T) {
	main := "\x1b[42m" + strings.Repeat("G", 20) + "\x1b[0m"
	modal := "MODALMODAL"
	out := OverlayView(main, modal, 30, 1, 0, 5)
	i := strings.Index(out, "MODAL")
	if i < 0 {
		t.Fatal("modal not found")
	}
	before := out[:i]
	if !strings.Contains(before, ansi.ResetStyle) {
		t.Fatalf("expected SGR reset before modal, prefix ends with: %q", before[len(before)-min(40, len(before)):])
	}
}

func TestOverlayView_reappliesStyleHiddenByModal(t *testing.T) {
	const plain = "012345678901234"
	purple := "\x1b[48;5;57m" + strings.Repeat("x", 36) + "\x1b[0m"
	mainView := plain + purple
	modal := strings.Repeat("M", 8)

	out := OverlayView(mainView, modal, 80, 1, 0, 10)
	if !strings.Contains(out, "MMMMMMMM") {
		t.Fatalf("modal content missing: %q", out)
	}
	i := strings.Index(out, "MMMMMMMM")
	afterModal := out[i+len("MMMMMMMM"):]
	if !strings.Contains(afterModal, "\x1b[48;5;57m") {
		n := 120
		if len(afterModal) < n {
			n = len(afterModal)
		}
		t.Fatalf("expected purple SGR after modal to style visible tail, got: %q", afterModal[:n])
	}
}

func TestOverlayView_examplesBehavior(t *testing.T) {
	var mainLines []string
	for i := range 25 {
		mainLines = append(mainLines, fmt.Sprintf("Line %02d: This is the background content that should be visible.", i))
	}
	mainView := strings.Join(mainLines, "\n")

	modalLines := []string{
		"┌──────────────────┐",
		"│      MODAL       │",
		"│      POPUP       │",
		"│      TEST        │",
		"└──────────────────┘",
	}
	modalView := strings.Join(modalLines, "\n")

	out := OverlayView(mainView, modalView, 80, 25, 5, 10)
	lines := strings.Split(out, "\n")

	if !strings.Contains(lines[24], "Line 24") {
		t.Errorf("Row 24 should contain background content, got %q", lines[24])
	}

	if !strings.HasPrefix(ansi.Strip(lines[5]), "Line 05: T") {
		t.Errorf("Row 5 left background mismatch. Got %q", lines[5])
	}
}
