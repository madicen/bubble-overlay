package overlay

import "fmt"

// DevStackDepthFooter returns a short footer line for debugging overlay depth when dev is true.
func DevStackDepthFooter(depth int, dev bool) string {
	if !dev {
		return ""
	}
	return fmt.Sprintf("[overlay depth=%d]", depth)
}
