Output "screenshots/transparency.gif"
Source "vhs/_env.tape"
Set Theme "Dracula"
Set FontSize 20
Set Width 800
Set Height 600

Hide
Type "go mod download"
Enter
Sleep 3s
Type "go run ./examples/transparency"
Enter
Sleep 1s
Show

# Show the list and header clearly before any interaction
Sleep 2s

# Navigate the list to show standard behavior
Down 1
Sleep 500ms
Down 1
Sleep 500ms
Down 1
Sleep 1s

# Toggle the overlay - Header remains visible
Space
Sleep 1.5s

# Scroll with magnifying glass effect
Down 1
Sleep 800ms
Down 1
Sleep 800ms
Space
Sleep 2s