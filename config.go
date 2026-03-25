package overlay

// PlacementKind selects how the top-left corner of an overlay is chosen
// relative to the viewport.
type PlacementKind uint8

const (
	placementCenter PlacementKind = iota
	placementRightDrawer
	placementFixed
)

// Placement describes where to anchor the overlay rectangle.
type Placement struct {
	kind      PlacementKind
	fixedTop  int
	fixedLeft int
}

// Center anchors the modal so it is centered in the viewport.
func Center() Placement {
	return Placement{kind: placementCenter}
}

// RightDrawer anchors the modal flush to the right edge (top row 0).
// Height/width follow the rendered modal string.
func RightDrawer() Placement {
	return Placement{kind: placementRightDrawer}
}

// Fixed anchors the modal with its top-left at (top, left) in cells.
func Fixed(top, left int) Placement {
	return Placement{kind: placementFixed, fixedTop: top, fixedLeft: left}
}

// Origin returns the top-left cell for the modal given its size and the viewport.
func (p Placement) Origin(modalW, modalH, viewW, viewH int) (top, left int) {
	switch p.kind {
	case placementCenter:
		top = (viewH - modalH) / 2
		left = (viewW - modalW) / 2
	case placementRightDrawer:
		top = 0
		left = max(0, viewW-modalW)
	case placementFixed:
		top, left = p.fixedTop, p.fixedLeft
	}
	if top < 0 {
		top = 0
	}
	if left < 0 {
		left = 0
	}
	return top, left
}

// OverlayConfig controls dimming, dismissal, and placement for one stack frame.
type OverlayConfig struct {
	Placement Placement

	// DimOpacity in (0,1] applies a faint style to the surface under this overlay.
	// 0 disables dimming for this frame.
	DimOpacity float64

	// CloseOnEscape pops this overlay when Escape is pressed (handled before the
	// top model receives the key).
	CloseOnEscape bool

	// CloseOnClickOutside pops when the user presses the primary button outside
	// the overlay rectangle. Requires tea.WithMouseCellMotion() (or similar).
	CloseOnClickOutside bool
}

// DefaultOverlayConfig returns sensible defaults: centered, light dim, Escape closes.
func DefaultOverlayConfig() OverlayConfig {
	return OverlayConfig{
		Placement:           Center(),
		DimOpacity:          0.35,
		CloseOnEscape:       true,
		CloseOnClickOutside: false,
	}
}
