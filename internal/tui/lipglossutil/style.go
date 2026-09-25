package lipglossutil

import (
	"github.com/charmbracelet/x/ansi"
)

// StripStyles removes all ANSI escape sequences from a rendered string
func StripStyles(renderedView string) string {
	return string(ansi.Strip(renderedView))
}
