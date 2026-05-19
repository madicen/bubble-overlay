package overlayv2

import (
	tea "charm.land/bubbletea/v2"
	bov "github.com/madicen/bubble-overlay"
)

type ViewAdapter interface {
	Adapt(base string, frames []FrameView, w, h int) tea.View
}

type FrameView struct {
	ModelView string
	Cfg       bov.OverlayConfig
	Layer     *bov.LayerState
}

type StringPipelineAdapter struct{}

func (StringPipelineAdapter) Adapt(base string, frames []FrameView, w, h int) tea.View {
	cur := base
	for _, fr := range frames {
		if fr.Cfg.DimOpacity > 0 {
			cur = bov.DimSurface(cur, fr.Cfg.DimOpacity)
		}
		modal := bov.RenderEntryModal(fr.ModelView, fr.Cfg, fr.Layer)
		top, left := bov.EntryClampedOrigin(fr.Cfg, fr.Layer, modal, w, h)
		cur = bov.ComposeModalLayer(cur, modal, fr.Cfg, top, left, w, h)
	}
	return tea.NewView(cur)
}
