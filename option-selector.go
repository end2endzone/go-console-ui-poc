package main

import "github.com/charmbracelet/lipgloss"

// View holds a collection of panels and its own dimensions.
type OptionSelector struct {
	Values       []string `json:"values"`
	Size         Size     `json:"size,omitempty"`       // Size of the rendered area
	CursorIcon   rune     `json:"cursorIcon,omitempty"` // Cursor icon highlighting selected option
	CursorIndex  int      `json:"cursor,omitempty"`     // Index of selected option in Values
	scrollOffset int      //`json:"scrollOffset,omitempty"` // For list scrolling subwindow. When all values can not be rendered into the Size area.
}

func NewOptionSelector() OptionSelector {
	s := OptionSelector{
		CursorIcon: '>',
	}
	s.Reset()
	return s
}

func (s *OptionSelector) Reset() {
	s.Values = []string{}
	s.Size.Width = 0
	s.Size.Height = 0
	s.CursorIndex = 0
	s.scrollOffset = 0
}

func (s *OptionSelector) Refresh() {
	if s.CursorIndex >= len(s.Values) {
		s.CursorIndex = len(s.Values) - 1
	}
	if s.CursorIndex < 0 {
		s.CursorIndex = 0
	}
	s.scrollOffset = 0
}

func (s *OptionSelector) UpdateScrollWindow() {
	visibleItems := s.Size.Height
	if s.CursorIndex < s.scrollOffset {
		s.scrollOffset = s.CursorIndex
	} else if s.CursorIndex >= s.scrollOffset+visibleItems {
		s.scrollOffset = s.CursorIndex - visibleItems + 1
	}
}

func (s *OptionSelector) MoveUp() {
	if s.CursorIndex > 0 {
		s.CursorIndex--
	}
	s.UpdateScrollWindow()
}

func (s *OptionSelector) MoveDown() {
	if s.CursorIndex < len(s.Values)-1 {
		s.CursorIndex++
	}
	s.UpdateScrollWindow()
}

func (s *OptionSelector) GetSelection() string {
	index := s.CursorIndex
	if index >= len(s.Values) {
		index = len(s.Values) - 1
	}
	return s.Values[index]
}

func (s *OptionSelector) Render() string {
	endIdx := s.scrollOffset + s.Size.Height
	if endIdx > len(s.Values) {
		endIdx = len(s.Values)
	}

	cursorWidth := 2 // 1 character for the cursor and 1 space after the icon

	content := ""

	// Render each options.
	// We skip some options (scrollOffset) if the number of Values exceeds how many value can fit in the rendering area (Size).
	// We then only render a subwindow of the options Values.
	for i := s.scrollOffset; i < endIdx; i++ {
		// Truncate text if too long
		maxLen := s.Size.Width - cursorWidth // Account for indicator
		display := truncateTextWidth(s.Values[i], maxLen)

		// If this Books is the selected book...
		cursor := "  "
		if s.CursorIndex == i {
			// Set the selected cursor & colorize with a style this book's title
			style := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000")).Bold(true)
			cursor = style.Render((string)(s.CursorIcon) + " ")
		}

		// Render the Book
		content += cursor + display + "\n"
	}

	style := lipgloss.NewStyle().
		Width(s.Size.Width).
		Height(s.Size.Height)
	output := style.Render(content)

	return output
}
