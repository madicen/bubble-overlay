Output "screenshots/draggable.gif"
Set Theme "Dracula"
Set FontSize 20
Set Width 800
Set Height 600
Set TypingSpeed 50ms
Source "vhs/_env.tape"

Hide
Type "go mod download"
Enter
Sleep 3s
Type "go run ./examples/draggable"
Enter
Sleep 1s
Show
Sleep 1s

# Open window
Type " "
Sleep 500ms

# Move (VHS Ctrl+Alt+arrow; app also accepts Alt+arrow)
Ctrl+Alt+Right
Sleep 80ms
Ctrl+Alt+Right
Sleep 80ms
Ctrl+Alt+Right
Sleep 80ms
Ctrl+Alt+Right
Sleep 80ms
Ctrl+Alt+Right
Sleep 80ms
Ctrl+Alt+Right
Sleep 80ms
Ctrl+Alt+Down
Sleep 80ms
Ctrl+Alt+Down
Sleep 80ms
Ctrl+Alt+Down
Sleep 80ms
Ctrl+Alt+Left
Sleep 80ms
Ctrl+Alt+Left
Sleep 80ms
Ctrl+Alt+Left
Sleep 80ms
Ctrl+Alt+Left
Sleep 80ms
Ctrl+Alt+Left
Sleep 80ms
Ctrl+Alt+Left
Sleep 80ms
Ctrl+Alt+Left
Sleep 80ms
Ctrl+Alt+Left
Sleep 80ms

# Resize wider, then taller
Ctrl+Alt+Shift+Right
Sleep 80ms
Ctrl+Alt+Shift+Right
Sleep 80ms
Ctrl+Alt+Shift+Right
Sleep 80ms
Ctrl+Alt+Shift+Right
Sleep 80ms
Ctrl+Alt+Shift+Right
Sleep 80ms
Ctrl+Alt+Shift+Right
Sleep 80ms
Ctrl+Alt+Shift+Right
Sleep 80ms
Ctrl+Alt+Shift+Right
Sleep 80ms
Ctrl+Alt+Shift+Right
Sleep 80ms
Ctrl+Alt+Shift+Right
Sleep 80ms
Ctrl+Alt+Shift+Down
Sleep 80ms
Ctrl+Alt+Shift+Down
Sleep 80ms
Ctrl+Alt+Shift+Down
Sleep 80ms
Ctrl+Alt+Shift+Down
Sleep 3s

# Close window, then quit
Escape
Sleep 500ms
Type "q"
Sleep 1s
