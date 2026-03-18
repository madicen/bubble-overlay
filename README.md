# bubble-overlay

Composable, float-over-content modals for [Bubble Tea](https://github.com/charmbracelet/bubbletea).

![Simple Demo](screenshots/simple.gif)

## Gallery

| Confirm Dialog | Form Input | Async Spinner |
| :---: | :---: | :---: |
| ![Confirm](screenshots/confirm.gif) | ![Form](screenshots/form.gif) | ![Spinner](screenshots/spinner.gif) |

## Installation

```bash
go get github.com/diceguyd30/bubble-overlay
```

## Usage

`bubble-overlay` provides a simple way to overlay a string view (like a modal or dialog) on top of another string view (your main application interface) without destroying the background context. It handles ANSI escape codes correctly, ensuring styles in the background and foreground are preserved and alignment remains intact.

```go
import (
    overlay "github.com/diceguyd30/bubble-overlay"
)

// In your View() method:
func (m model) View() string {
    // Render your main application view
    mainView := m.mainContent()

    if m.showModal {
        // Render your modal content (e.g. using lipgloss)
        modalView := m.modalContent()
        
        // Composite the modal over the main view
        // Signature: OverlayView(main, modal, width, height, y, x)
        return overlay.OverlayView(mainView, modalView, m.width, m.height, 5, 10)
    }
    
    return mainView
}
```

## Running Examples

```bash
go run examples/simple/main.go
go run examples/confirm/main.go
go run examples/form/main.go
go run examples/spinner/main.go
```