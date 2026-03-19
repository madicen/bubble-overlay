package overlay

import (
	"fmt"
	"strings"
	"testing"
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
		line := lines[row]
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

	out := OverlayView(mainView, modalView, 20, 5, 1, 5)
	lines := strings.Split(out, "\n")

	// Row 1 (first modal row): main has A(0-4), then overlay 5-14, then main B(15-19).
	// Overlay region: modal "MM  MM    " -> where modal has space, main shows (A or B). So we expect A and B to show through.
	row1 := lines[1]
	// Left of overlay (cols 0-4): A
	if !strings.Contains(row1[:5], "A") {
		t.Errorf("row 1 left: want A from main, got %q", row1[:5])
	}
	// Overlay (cols 5-14): M where modal has M, main (A or B) where modal has space. Modal "MM  MM    " -> positions 2,3 are space (main B from 7,8), 6,7,8,9 are space (main A from 11-14? No, overlay 5-14: main region is mainLine[5:15] = "BBBBBAAAAA". So cell 0 of overlay = main col 5 = B, cell 1 = B, cell 2 = B, cell 3 = B, cell 4 = A, cell 5 = A, cell 6 = A, cell 7 = A, cell 8 = A, cell 9 = A. Modal "MM  MM    " = M M space space M M space space space space. So we want: M M B B M M A A A A.
	if !strings.Contains(row1[5:15], "M") {
		t.Errorf("row 1 overlay region: want modal content, got %q", row1[5:15])
	}
	if !strings.Contains(row1[15:], "B") {
		t.Errorf("row 1 right margin: want B from main, got %q", row1[15:])
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
	if !strings.HasPrefix(lines[5], "Line 05: T") { // "Line 05: " is 9 chars. "T" is 10th char (index 9). Modal starts at 10.
		t.Errorf("Row 5 left background mismatch. Got %q", lines[5])
	}
}
