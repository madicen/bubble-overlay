# ADR: Bubble Tea v2 bridge (overlayv2)

## Status

Accepted for `github.com/madicen/bubble-overlay/v2` (package `overlayv2`), dual-track with v1.

## Context

- Bubble Tea v2 uses `View() tea.View`, `charm.land/bubbletea/v2`, and different input types per the [upgrade guide](https://github.com/charmbracelet/bubbletea/blob/main/UPGRADE_GUIDE_V2.md).
- The `overlay` package keeps string `OverlayView` and v1 `OverlayStack`.

## Decision

**R1 — String pipeline adapter:** `overlayv2.Stack` flattens each child `tea.View` to a string, applies `overlay.DimSurface` and `overlay.OverlayView`, then `tea.NewView(content)`.

**R2 (deferred):** Native lipgloss v2 / `tea.Layer` stacking only after parity with `OverlayView` ANSI behavior.

## Consequences

- v2 apps get stack semantics and `tea.View` without reimplementing hole-punching.
- Z-order is logical (single composited string per frame) until R2.
