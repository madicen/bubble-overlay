# bubble-overlay

Composable modals and overlay stacks for [Bubble Tea](https://github.com/charmbracelet/bubbletea): **v1** (`View() string` + `OverlayView` / `OverlayStack`) and **v2** (`View() tea.View` + [`overlayv2`](v2/)).

![Simple Demo](screenshots/simple.gif)

## Requirements

- **Go 1.25+** (this module pins `charm.land/bubbletea/v2` alongside Bubble Tea v1.)

## Installation

```bash
go get github.com/madicen/bubble-overlay
```

For the v2 helpers only, the same module includes import path **`github.com/madicen/bubble-overlay/v2`** (package `overlayv2`); you still `go get` the root module once.

## Which API should I use?

| You are on… | Use |
|-------------|-----|
| [Bubble Tea v1](https://github.com/charmbracelet/bubbletea) (`github.com/charmbracelet/bubbletea`), `View() string` | Package **`overlay`**: `OverlayView`, optional `OverlayStack`, `OverlayConfig`, `FocusTrap`. |
| [Bubble Tea v2](https://github.com/charmbracelet/bubbletea/blob/main/UPGRADE_GUIDE_V2.md) (`charm.land/bubbletea/v2`), `View() tea.View` | Package **`overlayv2`**: [`Stack`](v2/stack.go), `CompositeView`, same placement/dim config from **`overlay`**. |

Both paths share **string compositing** for the actual hole in the terminal (`overlay.OverlayView`, ANSI-safe). The v2 stack flattens each child `tea.View` to a string, composites, then wraps with `tea.NewView` (see [docs/ADR-v2-bridge.md](docs/ADR-v2-bridge.md)).

---

## Bubble Tea v1 — quick start

**Low-level:** composite a lipgloss (or plain) modal string over a main string without destroying cells under the rest of the screen.

```go
import overlay "github.com/madicen/bubble-overlay"

func (m model) View() string {
    mainView := m.renderMain()
    if !m.showModal {
        return mainView
    }
    modal := lipgloss.NewStyle().Width(40).Render("…")
    // OverlayView(main, modal, width, height, row, col)
    return overlay.OverlayView(mainView, modal, m.width, m.height, 5, 10)
}
```

**Why it exists:** grapheme-aware widths match lipgloss; SGR and hyperlinks under the modal are reset before the hole and **re-applied** after it so long colored lines look correct (try **`examples/colors`**).

**Higher-level:** [`OverlayStack`](stack.go) + [`OverlayConfig`](config.go) for nested modals, dimming, `Center` / `RightDrawer` / `Fixed`, Escape and optional click-outside, plus [`FocusTrap`](focus.go) so the base model skips keys/mouse while overlays are open. Wire **`Update`** like the [examples](#examples): forward `WindowSizeMsg` to the stack; when `!stack.MainReceivesKeyMsg()`, route `KeyMsg` / `MouseMsg` only to `stack.Update` (see `examples/simple`, `examples/confirm`, `examples/stack`).

---

## Bubble Tea v2 — quick start

Read Charm’s **[Bubble Tea v2 upgrade guide](https://github.com/charmbracelet/bubbletea/blob/main/UPGRADE_GUIDE_V2.md)** first (`KeyPressMsg`, `View() tea.View`, alt screen on the view, etc.).

```go
import (
    tea "charm.land/bubbletea/v2"
    bov "github.com/madicen/bubble-overlay"
    "github.com/madicen/bubble-overlay/v2"
)

func (m model) View() tea.View {
    v := m.stack.CompositeView(m.mainString, m.width, m.height)
    v.AltScreen = true
    return v
}
```

- Build overlay models with **`View() tea.View`** (e.g. `tea.NewView(lipgloss…)`).
- Use **`overlayv2.FocusTrap`** the same way as v1 `FocusTrap`: when the stack has focus, do not send interactive messages to your base layer.
- Optional: implement [`overlayv2.ViewAdapter`](v2/adapter.go) on `Stack.Adapter` to replace the default R1 string compositor later.

End-to-end sample: **`go run examples/v2simple/main.go`**.

---

## Package reference

| Symbol | Package | Role |
|--------|---------|------|
| `OverlayView` | `overlay` | Place modal rectangle over main string (`width`×`height` grid). |
| `DimSurface` | `overlay` | Faint-dim a multiline string (used by stacks / v2 adapter). |
| `OverlayConfig`, `Placement` (`Center`, `RightDrawer`, `Fixed`) | `overlay` | Per-frame dimming and anchor (shared by v1 stack and v2). |
| `OverlayStack` | `overlay` | v1 stack: `Push`/`Pop`/`View(string)`/`Update`. |
| `OverlayOnCloser`, `FocusTrap`, `DevStackDepthFooter` | `overlay` / `overlayv2` | Close hooks, focus helper, debug footer. |
| `Stack`, `ViewAdapter`, `StringPipelineAdapter`, `ViewString` | `overlayv2` | v2 stack + R1 compositing. |

---

## Examples

| Example | Bubble Tea | What it demonstrates |
|---------|--------------|----------------------|
| `examples/simple` | v1 | `OverlayStack`, center + dim, Escape / space |
| `examples/confirm` | v1 | Dialog keys, `dismissConfirmMsg` + `Pop` |
| `examples/stack` | v1 | Nested stack, `nestedOpenMsg`, optional `OVERLAY_DEV=1` footer |
| `examples/form` | v1 | `OverlayView` + `textinput` |
| `examples/spinner` | v1 | `OverlayView` + async spinner |
| `examples/colors` | v1 | `OverlayView` + ANSI through the modal cut |
| `examples/v2simple` | v2 | `overlayv2.Stack`, `CompositeView`, `View() tea.View` |

From the **repository root**:

```bash
go run examples/simple/main.go
go run examples/confirm/main.go
go run examples/form/main.go
go run examples/spinner/main.go
go run examples/colors/main.go
go run examples/stack/main.go
OVERLAY_DEV=1 go run examples/stack/main.go   # stack depth footer
go run examples/v2simple/main.go
```

---

## Gallery

| Confirm Dialog | Form Input | Async Spinner | ANSI / colors | Nested stack |
| :---: | :---: | :---: | :---: | :---: |
| ![Confirm](screenshots/confirm.gif) | ![Form](screenshots/form.gif) | ![Spinner](screenshots/spinner.gif) | ![Colors](screenshots/colors.gif) | ![Stack](screenshots/stack.gif) |

---

## Recording GIFs (VHS)

Lipgloss uses [termenv](https://github.com/muesli/termenv) to decide whether to emit ANSI. If **`CI`** or **`NO_COLOR`** is set, styles may strip to ASCII in recordings.

Tapes in `vhs/` **`Source "vhs/_env.tape"`**, which clears those variables and sets **`TERM=xterm-256color`** and **`COLORTERM=truecolor`**. Run VHS from the **repository root** (`make gifs`).

Each tape runs **`go mod download`** inside **`Hide`** before **`go run ./examples/...`** so download lines do not appear in GIFs.
