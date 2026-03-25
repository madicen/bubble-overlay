Output "screenshots/stack.gif"
Source "vhs/_env.tape"
Set Theme "Dracula"
Set FontSize 20
Set Width 800
Set Height 600

Hide
Type "go mod download"
Enter
Sleep 3s
Type "OVERLAY_DEV=1 go run ./examples/stack"
Enter
Sleep 1s
Show
Sleep 1s
Type " "
Sleep 1s
Type "i"
Sleep 1s
Escape
Sleep 1s
Escape
Sleep 1s
Type "q"
Sleep 1s
