package overlay

import (
	"fmt"
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

func TestOverlayView_replacesOnlyModalRect(t *testing.T) {
	// Main view: 30 wide, 10 tall; content "MAIN" on left and "END" on right
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
	// Modal: 10 wide, 4 tall, content "MMMM"
	modalLines := []string{
		"MMMMMMMMMM",
		"MMMMMMMMMM",
		"MMMMMMMMMM",
		"MMMMMMMMMM",
	}
	modalView := strings.Join(modalLines, "\n")

	viewWidth, viewHeight := 30, 10
	top, left := 3, 10 // modal at row 3, col 10
	out := OverlayView(mainView, modalView, viewWidth, viewHeight, top, left)
	lines := strings.Split(out, "\n")

	if len(lines) != viewHeight {
		t.Fatalf("got %d lines, want %d", len(lines), viewHeight)
	}
	// Rows 0-2: full main (L)
	for row := 0; row < top; row++ {
		if !strings.Contains(lines[row], "L") || strings.Contains(lines[row], "M") {
			t.Errorf("row %d: should be main only, got %q", row, lines[row])
		}
	}
	// Rows 3-6: main left (cols 0-9) + modal (cols 10-19) + main right (cols 20-29)
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
	// Rows 7-9: full main again
	for row := top + 4; row < viewHeight; row++ {
		if !strings.Contains(lines[row], "L") || strings.Contains(lines[row], "M") {
			t.Errorf("row %d: should be main only, got %q", row, lines[row])
		}
	}
}

func TestOverlayView_modalReplacesRegion(t *testing.T) {
	// Main: 20 wide, 5 tall; content "A" on left, "B" on right
	mainLines := []string{
		"AAAAABBBBBAAAAABBBBB",
		"AAAAABBBBBAAAAABBBBB",
		"AAAAABBBBBAAAAABBBBB",
		"AAAAABBBBBAAAAABBBBB",
		"AAAAABBBBBAAAAABBBBB",
	}
	mainView := strings.Join(mainLines, "\n")
	// Modal: 10 wide, 3 tall; row 0 "MM  MM" (spaces in middle), row 1 "  MMMM  ", row 2 "MM    MM"
	modalLines := []string{
		"MM  MM    ",
		"  MMMM    ",
		"MM    MM  ",
	}
	modalView := strings.Join(modalLines, "\n")

	// Test with transparency enabled
	out := OverlayViewWithTransparency(mainView, modalView, 20, 5, 1, 5)
	lines := strings.Split(out, "\n")

	// Row 1 (first modal row): main has A(0-4), then overlay 5-14, then main B(15-19).
	// Overlay region: modal "MM  MM    " -> where modal has space, main shows (A or B). So we expect A and B to show through.
	row1 := ansi.Strip(lines[1])

	// mainLine[1]: "AAAAABBBBBAAAAABBBBB"
	// overlay at left=5, modalW=10.
	// main region covered: "BBBBBAAAAA" (index 5 to 14)
	// modalLine[0]: "MM  MM    " (10 chars: M, M, space, space, M, M, space, space, space, space)
	// expected result at indices 5-14: "MMBBMMAAAA"
	wantMid := "MMBBMMAAAA"
	if gotMid := row1[5:15]; gotMid != wantMid {
		t.Errorf("row 1 overlay region: want %q, got %q", wantMid, gotMid)
	}
}

func TestOverlayView_opaqueIsTrulyOpaque(t *testing.T) {
	mainView := "AAAAAAAAAA"
	modalView := "  MM  "
	// 10 wide, 1 tall. Modal at left 2, width 6.
	out := OverlayView(mainView, modalView, 10, 1, 0, 2)
	stripped := ansi.Strip(out)
	// Expected: AA  MM  AA (the spaces in modal overwrite the main view)
	want := "AA  MM  AA"
	if stripped != want {
		t.Errorf("opaque overlay: want %q, got %q", want, stripped)
	}
}

func TestOverlayView_resetsPenBeforeModal(t *testing.T) {
	main := "\x1b[42m" + strings.Repeat("G", 20) + "\x1b[0m" // green bg
	modal := "MODALMODAL"
	out := OverlayView(main, modal, 30, 1, 0, 5)
	// After 5 G cells, SGR reset + hyperlink reset must precede modal so border is not on green.
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
	// 15 plain cells, then purple background; opening SGR starts at column 15.
	// Modal covers columns 10–17, so the SGR bytes live under the modal while
	// purple "x" cells continue to the right — output must re-inject style after the modal.
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
	// Resume block resets then reapplies purple before the tail runes.
	if !strings.Contains(afterModal, "\x1b[48;5;57m") {
		n := 120
		if len(afterModal) < n {
			n = len(afterModal)
		}
		t.Fatalf("expected purple SGR after modal to style visible tail, got: %q", afterModal[:n])
	}
}

func TestOverlayView_examplesBehavior(t *testing.T) {
	// Ensure that when a full-screen main view is provided (as intended in examples),
	// the background remains visible around the modal.

	// Construct 80x25 main view (simulating typical example size)
	var mainLines []string
	for i := 0; i < 25; i++ {
		mainLines = append(mainLines, fmt.Sprintf("Line %02d: This is the background content that should be visible.", i))
	}
	mainView := strings.Join(mainLines, "\n")

	// Simple box modal: 20x5
	modalLines := []string{
		"┌──────────────────┐",
		"│      MODAL       │",
		"│      POPUP       │",
		"│      TEST        │",
		"└──────────────────┘",
	}
	modalView := strings.Join(modalLines, "\n")

	// Overlay at top 5, left 10
	out := OverlayView(mainView, modalView, 80, 25, 5, 10)
	lines := strings.Split(out, "\n")

	// Verify last line (24) is intact (was failing in examples due to short mainView)
	if !strings.Contains(lines[24], "Line 24") {
		t.Errorf("Row 24 should contain background content, got %q", lines[24])
	}

	// Verify line 5 (start of modal) has background on left
	if !strings.HasPrefix(ansi.Strip(lines[5]), "Line 05: T") { // "Line 05: " is 9 chars. "T" is 10th char (index 9). Modal starts at 10.
		t.Errorf("Row 5 left background mismatch. Got %q", lines[5])
	}
}
