package lipglossutil

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

// StripStyles removes all ANSI escape sequences from a rendered string
func StripStyles(renderedView string) string {
	return string(ansi.Strip(renderedView))
}

func IsStyleBordered(s *lipgloss.Style) bool {
	if s.GetBorderTop() ||
		s.GetBorderBottom() ||
		s.GetBorderLeft() ||
		s.GetBorderRight() {
		return true
	}
	return false
}
