package ui

import (
	"github.com/charmbracelet/lipgloss"
)

// View holds a collection of panels and its own dimensions.
type OptionSelector struct {
	Options         []any            `json:"values"`               // The list of option values to select from.
	OptionsRenderer func(any) string ``                            // An external function to convert an Option to a display string.
	Size            Size             `json:"size,omitempty"`       // Size of the rendered area
	CursorIcon      rune             `json:"cursorIcon,omitempty"` // Cursor icon highlighting selected option
	CursorIndex     int              `json:"cursor,omitempty"`     // Index of selected option in Values
	CursorStyle     *lipgloss.Style  ``                            // An optional rendering style for the cursor icon
	SelectionStyle  *lipgloss.Style  ``                            // An optional rendering style for the selected option
	scrollOffset    int              //`json:"scrollOffset,omitempty"` // For list scrolling subwindow. When all values can not be rendered into the Size area.
}

func NewOptionSelector() OptionSelector {
	s := OptionSelector{
		CursorIcon: '>',
	}
	s.Reset()
	return s
}

func (s *OptionSelector) Reset() {
	s.Options = []any{}
	s.OptionsRenderer = nil
	s.Size.Width = 0
	s.Size.Height = 0
	s.CursorIndex = 0
	s.scrollOffset = 0
}

func (s *OptionSelector) Refresh() {
	if s.CursorIndex >= len(s.Options) {
		s.CursorIndex = len(s.Options) - 1
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
	if s.CursorIndex < len(s.Options)-1 {
		s.CursorIndex++
	}
	s.UpdateScrollWindow()
}

func (s *OptionSelector) GetSelectedOption() any {
	index := s.CursorIndex
	if index >= len(s.Options) {
		index = len(s.Options) - 1
	}
	return s.Options[index]
}

func (s *OptionSelector) Render() string {
	endIdx := s.scrollOffset + s.Size.Height
	if endIdx > len(s.Options) {
		endIdx = len(s.Options)
	}

	cursorWidth := 2 // 1 character for the cursor and 1 space after the icon

	content := ""

	// Render each options.
	// We skip some options (scrollOffset) if the number of options exceeds how many value can fit in the rendering area (Size).
	// We then only render a subwindow of the options.
	for i := s.scrollOffset; i < endIdx; i++ {
		// Render the option's display value

		// Get the option displayed value
		value := s.OptionsRenderer(s.Options[i])

		// Truncate text if too long
		maxLen := s.Size.Width - cursorWidth // Account for indicator
		value = TruncateTextWidth(value, maxLen)

		// Render the cursor
		cursor := "  "

		// If this option is the selected option...
		if s.CursorIndex == i {
			if s.CursorStyle != nil {
				// Override the selected cursor text and colorize with the configured style
				cursor = s.CursorStyle.Render((string)(s.CursorIcon)) + " "
			}
			if s.SelectionStyle != nil {
				// Override the selected displayed value text and colorize with the configured style
				value = s.SelectionStyle.Render(value)
			}
		}

		// Render the option's display value
		content += cursor + value + "\n"
	}

	style := lipgloss.NewStyle().
		Width(s.Size.Width).
		Height(s.Size.Height)
	output := style.Render(content)

	return output
}
