package overlay

import "fmt"

func DevStackDepthFooter(depth int, dev bool) string {
	if !dev {
		return ""
	}
	return fmt.Sprintf("[overlay depth=%d]", depth)
}
