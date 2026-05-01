package overlay

type PlacementKind uint8

const (
	placementCenter PlacementKind = iota
	placementRightDrawer
	placementFixed
)

type Placement struct {
	kind      PlacementKind
	fixedTop  int
	fixedLeft int
}

func Center() Placement {
	return Placement{kind: placementCenter}
}

func RightDrawer() Placement {
	return Placement{kind: placementRightDrawer}
}

func Fixed(top, left int) Placement {
	return Placement{kind: placementFixed, fixedTop: top, fixedLeft: left}
}

// Origin returns the placement anchor before OverlayView overflow clamping: negative top/left are
// pinned to 0, but if the modal is wider or taller than the viewport the origin is not shifted
// until compositing—use ClampedOrigin or ClampOverlayOrigin for coordinates that must match painting
// (e.g. mouse hit-testing).
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

// ClampedOrigin returns Origin followed by the same overflow clamp as OverlayView. Use this when
// forwarding tea.MouseMsg or storing overlay bounds so geometry matches the compositor.
func (p Placement) ClampedOrigin(modalW, modalH, viewW, viewH int) (top, left int) {
	t, l := p.Origin(modalW, modalH, viewW, viewH)
	return ClampOverlayOrigin(modalW, modalH, viewW, viewH, t, l)
}

type OverlayConfig struct {
	Placement           Placement
	DimOpacity          float64
	CloseOnEscape       bool
	CloseOnClickOutside bool
}

func DefaultOverlayConfig() OverlayConfig {
	return OverlayConfig{
		Placement:           Center(),
		DimOpacity:          0.35,
		CloseOnEscape:       true,
		CloseOnClickOutside: false,
	}
}
