// Package overlayv2 provides OverlayStack for Bubble Tea v2 (charm.land/bubbletea/v2).
//
// Rendering uses the R1 "string pipeline": modal layers are flattened with
// github.com/madicen/bubble-overlay's OverlayView and DimSurface, then wrapped in tea.NewView.
// See docs/ADR-v2-bridge.md for rationale and future R2 (native layers) options.
package overlayv2
