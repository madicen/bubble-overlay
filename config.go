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
