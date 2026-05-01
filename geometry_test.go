package overlay

import "testing"

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
