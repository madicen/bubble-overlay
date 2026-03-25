# ADR: Bubble Tea v2 bridge (overlayv2)

## Status

Accepted for `github.com/madicen/bubble-overlay/v2` (package `overlayv2`), dual-track with v1.

## Context

- Bubble Tea v2 uses `View() tea.View`, `charm.land/bubbletea/v2`, and different input types (`KeyPressMsg`, structured mouse messages) per [UPGRADE_GUIDE_V2](https://github.com/charmbracelet/bubbletea/blob/main/UPGRADE_GUIDE_V2.md).
- The existing `overlay` package keeps string `OverlayView` and v1 `OverlayStack` for current consumers.

## Decision

**R1 — String pipeline adapter:** `overlayv2.Stack` composes modals by flattening each child’s `tea.View` to a string (`ViewString` → `Content`), then applying `overlay.DimSurface` and `overlay.OverlayView` in order, finally `tea.NewView(content)`.

**R2 (deferred):** Native lipgloss v2 / `tea.Layer` stacking would replace string composition only after parity tests match `OverlayView` for ANSI edge cases.

## Consequences

- v2 apps get stack semantics and `tea.View` output without reimplementing focus order in the compositor.
- Z-order is logical (paint order in one string), not a separate runtime layer tree.
- Go toolchain follows v2’s requirement (currently 1.25.x in go.mod).

## Alternatives considered

- v2-only module: rejected for this phase to avoid breaking v1 dependents.
- `tea.Window` as the primary API name: rejected — v2 docs center on `tea.View`; revisit if Charm exports a higher-level window type later.
