package overlay

import (
	"strings"
	"testing"
)

func TestClampOverlayOrigin_horizontalOverflow(t *testing.T) {
	top, left := ClampOverlayOrigin(40, 3, 30, 20, 5, 5)
	if top != 5 || left != 0 {
		t.Fatalf("got top=%d left=%d want top=5 left=0", top, left)
	}
}

func TestClampOverlayOrigin_verticalOverflow(t *testing.T) {
	top, left := ClampOverlayOrigin(10, 25, 80, 10, 0, 0)
	if top != 0 || left != 0 {
		t.Fatalf("got top=%d left=%d want top=0 left=0", top, left)
	}
}

func TestClampOverlayOrigin_idempotent_with_Placement(t *testing.T) {
	mw, mh := 10, 4
	viewW, viewH := 80, 25
	t0, l0 := Center().Origin(mw, mh, viewW, viewH)
	t1, l1 := ClampOverlayOrigin(mw, mh, viewW, viewH, t0, l0)
	t2, l2 := Center().ClampedOrigin(mw, mh, viewW, viewH)
	if t1 != t2 || l1 != l2 {
		t.Fatalf("ClampedOrigin mismatch: (%d,%d) vs (%d,%d)", t1, l1, t2, l2)
	}
}

func TestModalCellSize_matches_overlay_split(t *testing.T) {
	modal := "ab\ncd"
	w, h := ModalCellSize(modal)
	if w != 2 || h != 2 {
		t.Fatalf("got %dx%d want 2x2", w, h)
	}
}

func TestClampOverlayOriginAtPoint_matches_ModalCellSize_plus_clamp(t *testing.T) {
	modal := "Cut\nCopy\nPaste"
	viewW, viewH := 40, 15
	top, left := 4, 10
	t1, l1 := ClampOverlayOriginAtPoint(modal, viewW, viewH, top, left)
	mw, mh := ModalCellSize(modal)
	t2, l2 := ClampOverlayOrigin(mw, mh, viewW, viewH, top, left)
	if t1 != t2 || l1 != l2 {
		t.Fatalf("got (%d,%d) want (%d,%d)", t1, l1, t2, l2)
	}
}

func TestClampMenuOrigin_matches_ClampOverlayOriginAtPoint(t *testing.T) {
	modal := "a\nb"
	vw, vh := 50, 20
	for _, tc := range []struct{ top, left int }{{0, 0}, {3, 10}, {-2, 99}} {
		a, b := ClampMenuOrigin(modal, vw, vh, tc.top, tc.left)
		c, d := ClampOverlayOriginAtPoint(modal, vw, vh, tc.top, tc.left)
		if a != c || b != d {
			t.Fatalf("MenuOrigin (%d,%d) vs AtPoint (%d,%d)", a, b, c, d)
		}
	}
}

func TestOverlayViewAtPoint_matches_clamp_plus_OverlayView(t *testing.T) {
	main := "....\n....\n"
	modal := "XX"
	vw, vh := 4, 3
	top, left := 1, 1
	got := OverlayViewAtPoint(main, modal, vw, vh, top, left)
	tCl, lCl := ClampOverlayOriginAtPoint(modal, vw, vh, top, left)
	want := OverlayView(main, modal, vw, vh, tCl, lCl)
	if got != want {
		t.Fatalf("AtPoint vs manual overlay")
	}
}

func TestOverlayViewInCenterWithOffset_zero_delta_matches_InCenter(t *testing.T) {
	main := strings.Repeat("-", 20) + "\n" + strings.Repeat("-", 20)
	modal := "hi"
	vw, vh := 20, 2
	a := OverlayViewInCenterWithOffset(main, modal, vw, vh, 0, 0)
	b := OverlayViewInCenter(main, modal, vw, vh)
	if a != b {
		t.Fatalf("offset 0 should match InCenter")
	}
}

func TestOverlayViewInCenterInMain_matches_explicit_viewport(t *testing.T) {
	main := "hello\nworld\n"
	modal := "(.)"
	w, h := ModalCellSize(main)
	got := OverlayViewInCenterInMain(main, modal)
	want := OverlayViewInCenter(main, modal, w, h)
	if got != want {
		t.Fatalf("InMain mismatch with explicit ModalCellSize(main)")
	}
}

func TestOverlayViewInCenter_clamped_placement_matches_CellInModal(t *testing.T) {
	cases := []struct {
		name       string
		viewW, viewH int
		modal      string
	}{
		{"wide_modal", 14, 5, strings.Repeat("W", 80)},
		{"tall_modal", 8, 4, "l\nl\nl\nl\nl\nl"},
		{"narrow_square", 3, 3, "ABCDE\nFG"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mw, mh := ModalCellSize(tc.modal)
			top := (tc.viewH - mh) / 2
			left := (tc.viewW - mw) / 2
			tClamped, lClamped := ClampOverlayOriginAtPoint(tc.modal, tc.viewW, tc.viewH, top, left)
			ct, cl := Center().ClampedOrigin(mw, mh, tc.viewW, tc.viewH)
			if tClamped != ct || lClamped != cl {
				t.Fatalf("center+clamp vs ClampedOrigin: got (%d,%d) want (%d,%d)", tClamped, lClamped, ct, cl)
			}
			for y := tClamped; y < tClamped+mh && y < tc.viewH; y++ {
				for x := lClamped; x < lClamped+mw && x < tc.viewW; x++ {
					if !CellInModal(x, y, tClamped, lClamped, mw, mh) {
						t.Fatalf("visible cell (%d,%d) not in modal rect", x, y)
					}
				}
			}
		})
	}
}

