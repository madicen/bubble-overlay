package overlayv2

import (
	tea "charm.land/bubbletea/v2"
	bov "github.com/madicen/bubble-overlay"
	"github.com/madicen/bubble-overlay/internal/layout"
)

// ViewAdapter composes stacked modal layers into a single tea.View.
// The default implementation ([StringPipelineAdapter]) uses the v1 string
// OverlayView/DimSurface pipeline (R1) for ANSI-correct hole punching.
type ViewAdapter interface {
	Adapt(base string, frames []FrameView, w, h int) tea.View
}

// FrameView is one stack layer: flattened modal content and overlay config.
type FrameView struct {
	Modal string
	Cfg   bov.OverlayConfig
}

// StringPipelineAdapter implements R1 compositing via bov.OverlayView.
type StringPipelineAdapter struct{}

// Adapt implements [ViewAdapter].
func (StringPipelineAdapter) Adapt(base string, frames []FrameView, w, h int) tea.View {
	cur := base
	for _, fr := range frames {
		if fr.Cfg.DimOpacity > 0 {
			cur = bov.DimSurface(cur, fr.Cfg.DimOpacity)
		}
		mw, mh := layout.ModalCellSize(fr.Modal)
		top, left := fr.Cfg.Placement.Origin(mw, mh, w, h)
		cur = bov.OverlayView(cur, fr.Modal, w, h, top, left)
	}
	return tea.NewView(cur)
}
