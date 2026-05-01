package overlayv2

import (
	tea "charm.land/bubbletea/v2"
	bov "github.com/madicen/bubble-overlay"
	"github.com/madicen/bubble-overlay/internal/layout"
)

type ViewAdapter interface {
	Adapt(base string, frames []FrameView, w, h int) tea.View
}

type FrameView struct {
	Modal string
	Cfg   bov.OverlayConfig
}

type StringPipelineAdapter struct{}

func (StringPipelineAdapter) Adapt(base string, frames []FrameView, w, h int) tea.View {
	cur := base
	for _, fr := range frames {
		if fr.Cfg.DimOpacity > 0 {
			cur = bov.DimSurface(cur, fr.Cfg.DimOpacity)
		}
		mw, mh := layout.ModalCellSize(fr.Modal)
		top, left := fr.Cfg.Placement.ClampedOrigin(mw, mh, w, h)
		cur = bov.OverlayView(cur, fr.Modal, w, h, top, left)
	}
	return tea.NewView(cur)
}
