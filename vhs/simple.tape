Output "screenshots/simple.gif"
Source "vhs/_env.tape"
Set Theme "Dracula"
Set FontSize 20
Set Width 800
Set Height 600

Hide
Type "go mod download"
Enter
Sleep 3s
Type "go run ./examples/simple"
Enter
Sleep 1s
Show
Sleep 1s
Type " "
Sleep 2s
