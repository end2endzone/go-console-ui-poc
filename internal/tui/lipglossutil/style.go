package lipglossutil

import (
	"fmt"
	"strings"

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

// CompareStyles compares properties of two lipgloss.Style and returns a diff string.
func CompareStyles(s1 lipgloss.Style, s2 lipgloss.Style) string {
	var sb strings.Builder

	checkDiff := func(propName string, val1, val2 any) {
		str1 := fmt.Sprintf("%v", val1)
		str2 := fmt.Sprintf("%v", val2)
		if str1 != str2 {
			// diff := fmt.Sprintf("- %s:\n  Style 1: %s\n  Style 2: %s\n", propName, str1, str2)
			diff := fmt.Sprintf("- %s: '%s' vs '%s'\n", propName, str1, str2)
			sb.WriteString(diff)
		}
	}

	// Core Layout & Dimensions
	checkDiff("Width", s1.GetWidth(), s2.GetWidth())
	checkDiff("Height", s1.GetHeight(), s2.GetHeight())
	checkDiff("Foreground", s1.GetForeground(), s2.GetForeground())
	checkDiff("Background", s1.GetBackground(), s2.GetBackground())

	// Border
	checkDiff("BorderTop", s1.GetBorderTop(), s2.GetBorderTop())
	checkDiff("BorderRight", s1.GetBorderRight(), s2.GetBorderRight())
	checkDiff("BorderBottom", s1.GetBorderBottom(), s2.GetBorderBottom())
	checkDiff("BorderLeft", s1.GetBorderLeft(), s2.GetBorderLeft())

	// Padding
	checkDiff("PaddingTop", s1.GetPaddingTop(), s2.GetPaddingTop())
	checkDiff("PaddingRight", s1.GetPaddingRight(), s2.GetPaddingRight())
	checkDiff("PaddingBottom", s1.GetPaddingBottom(), s2.GetPaddingBottom())
	checkDiff("PaddingLeft", s1.GetPaddingLeft(), s2.GetPaddingLeft())

	// Margin
	checkDiff("MarginTop", s1.GetMarginTop(), s2.GetMarginTop())
	checkDiff("MarginRight", s1.GetMarginRight(), s2.GetMarginRight())
	checkDiff("MarginBottom", s1.GetMarginBottom(), s2.GetMarginBottom())
	checkDiff("MarginLeft", s1.GetMarginLeft(), s2.GetMarginLeft())

	// Text decorations & modifiers
	checkDiff("Bold", s1.GetBold(), s2.GetBold())
	checkDiff("Italic", s1.GetItalic(), s2.GetItalic())
	checkDiff("Faint", s1.GetFaint(), s2.GetFaint())
	checkDiff("Blink", s1.GetBlink(), s2.GetBlink())
	checkDiff("Reverse", s1.GetReverse(), s2.GetReverse())
	checkDiff("Strikethrough", s1.GetStrikethrough(), s2.GetStrikethrough())
	checkDiff("StrikethroughSpaces", s1.GetStrikethroughSpaces(), s2.GetStrikethroughSpaces())

	// Underline features
	checkDiff("Underline", s1.GetUnderline(), s2.GetUnderline())
	checkDiff("UnderlineSpaces", s1.GetUnderlineSpaces(), s2.GetUnderlineSpaces())

	// Borders & advanced properties
	checkDiff("BorderStyle", s1.GetBorderStyle(), s2.GetBorderStyle())
	checkDiff("Transform", s1.GetTransform() != nil, s2.GetTransform() != nil)

	if sb.Len() == 0 {
		return "" // Styles are identical.
	}
	return sb.String()
}
