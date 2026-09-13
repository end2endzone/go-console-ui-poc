package main

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

// Bounds defines spacing around an element.
type Bounds struct {
	Top    int `json:"top"`
	Right  int `json:"right"`
	Bottom int `json:"bottom"`
	Left   int `json:"left"`
}

// Size defines dimensions.
type Size struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

// Panel represents a specific UI component.
type Panel struct {
	Padding       Bounds `json:"padding,omitempty"`
	Margin        Bounds `json:"margin,omitempty"`
	BorderSize    Size   `json:"borderSize,omitempty"` // size of the panel
	Content       string `json:"content,omitempty"`
	MinimumWidth  int    `json:"minimumWidth,omitempty"`
	MinimumHeight int    `json:"minimumHeight,omitempty"`
}

// View holds a collection of panels and its own dimensions.
type View struct {
	Panels []*Panel `json:"panels"`
	Size   Size     `json:"size,omitempty"`
}

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

func (v *View) SplitPanelsVerticalyByRatio(ratios []float32) {
	if len(ratios) != len(v.Panels) {
		err := fmt.Errorf("Number of panels and split ratios do not match!")
		panic(err)
	}

	// Split the view size into the given panels
	remainingWidth := v.Size.Width
	for i, panel := range v.Panels {
		isLast := (i + 1) == len(v.Panels)

		var newSize Size

		// width
		if !isLast {
			newSize.Width = int(ratios[i] * float32(v.Size.Width))
		} else {
			newSize.Width = remainingWidth
		}
		remainingWidth -= newSize.Width

		// height
		newSize.Height = v.Size.Height

		panel.SetSize(newSize)
	}
}

func (p *Panel) Style() lipgloss.Style {
	size := p.GetStyleSize()

	style := lipgloss.NewStyle().
		Width(size.Width).
		Height(size.Height).
		Border(lipgloss.RoundedBorder()).
		Padding(p.Padding.Top, p.Padding.Right, p.Padding.Bottom, p.Padding.Left)

	return style
}

func (p *Panel) SetSize(size Size) {
	// Do not allow size < 0
	if size.Width < 0 {
		size.Width = 0
	}
	if size.Height < 0 {
		size.Height = 0
	}

	// Checks minimums
	if size.Width < p.MinimumWidth {
		size.Width = p.MinimumWidth
	}
	if size.Height < p.MinimumHeight {
		size.Height = p.MinimumHeight
	}

	p.BorderSize = size
}

func (p *Panel) GetBorderSize() Size {
	return p.BorderSize
}

// GetStyleSize returns the dimension of the box style inside the panel.
// This size is defined by the border size minus the width and height of the borders.
func (p *Panel) GetStyleSize() Size {
	size := p.BorderSize

	// Remove the width and height of the borders
	size.Width -= 2
	size.Height -= 2

	// Do not allow size < 0
	if size.Width < 0 {
		size.Width = 0
	}
	if size.Height < 0 {
		size.Height = 0
	}

	return size
}

// GetRenderSize returns the maximum dimension of the renderable area inside the panel.
// This size is defined by the border size minus the borders width/height and the inner padding.
func (p *Panel) GetRenderSize() Size {
	size := p.BorderSize

	// Remove the width and height of the borders
	size.Width -= 2
	size.Height -= 2

	// Remove padding
	size.Width -= (p.Padding.Left + p.Padding.Right)
	size.Height -= (p.Padding.Top + p.Padding.Bottom)

	// Do not allow size < 0
	if size.Width < 0 {
		size.Width = 0
	}
	if size.Height < 0 {
		size.Height = 0
	}

	return size
}

// GetOutterSize returns the total dimension of the panel.
// This dimensions is defined by the border size plus the outter margins.
func (p *Panel) GetOutterSize() Size {
	size := p.BorderSize

	size.Width += (p.Margin.Left + p.Margin.Right)
	size.Height += (p.Margin.Top + p.Margin.Bottom)

	return size
}

// ViewDebug renders a sample debugging view.
// The following code and styles shows lipgloss packages behavior.
// It renders the following text:
//
// ╭──────────╮╭──────────╮╭──────────╮╭──────────╮
// │1234567890││          ││    12345 ││If we     │
// │2         ││  text    ││    2aaaa ││render too│
// │3         ││  auto    ││    3bbbb ││much text │
// │4   10x8  ││  wrappe  ││    4cccc ││in the    │
// │5         ││  d in    ││    5dddd ││box, the  │
// │6         ││  paddin  ││    6eeee ││height of │
// │7         ││  g       ││          ││the box   │
// │8         ││          ││          ││will      │
// ╰──────────╯╰──────────╯╰──────────╯│expand to │
// .                                   │fit the   │
// .                                   │content.  │
// .                                   │Text must │
// .                                   │be        │
// .                                   │truncated │
// .                                   │manually. │
// .                                   ╰──────────╯
// ↑/↓: Navigate  |  Enter: Select  |  /: Search  |  Esc: Clear  |  q: Quit
// p3.borders(10,8), p3.inner(5,6)
// [EMPTY LINE]
func ViewDebug() string {
	/*
		panel1 := lipgloss.NewStyle().
			Padding(0, 0, 0, 0).
			Width(10).
			Height(8).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#FF0000"))

		panel2 := lipgloss.NewStyle().
			Padding(1, 2).
			Width(10).
			Height(8).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#00FF00"))

		panel3 := lipgloss.NewStyle().
			Padding(1).
			Width(10).
			Height(8).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#0000FF"))

		panel4 := lipgloss.NewStyle().
			Padding(0, 0, 0, 0).
			Width(10).
			Height(8).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("204"))
	*/

	panel1 := Panel{
		BorderSize: Size{12, 10}, // 10x8 style
	}

	panel2 := Panel{
		BorderSize: Size{12, 10}, // 10x8 style
		Padding:    Bounds{1, 2, 1, 2},
	}

	panel3 := Panel{
		BorderSize: Size{12, 10}, // 10x8 style
		Padding:    Bounds{0, 1, 2, 4},
	}

	panel4 := Panel{
		BorderSize: Size{12, 10}, // 10x8 style
	}

	s1 := panel1.Style().
		BorderForeground(lipgloss.Color("#FF0000"))
	s2 := panel2.Style().
		BorderForeground(lipgloss.Color("#00FF00"))
	s3 := panel3.Style().
		BorderForeground(lipgloss.Color("#0000FF"))
	s4 := panel4.Style().
		BorderForeground(lipgloss.Color("204"))

	// Join all columns
	columns := lipgloss.JoinHorizontal(
		lipgloss.Top,
		s1.Render("1234567890\n2\n3\n4   10x8\n5\n6\n7\n8"),
		s2.Render("text auto wrapped in padding"),
		s3.Render("123452aaaa3bbbb4cccc5dddd6eeee"),
		s4.Render("If we render too much text in the box, the height of the box will expand to fit the content. Text must be truncated manually."),
	)

	helpText := "↑/↓: Navigate  |  Enter: Select  |  /: Search  |  Esc: Clear  |  q: Quit"

	sizes := fmt.Sprintf("p3.borders(%d,%d), p3.style(%d,%d), p3.render(%d,%d)\n",
		panel3.GetBorderSize().Width,
		panel3.GetBorderSize().Height,
		panel3.GetStyleSize().Width,
		panel3.GetStyleSize().Height,
		panel3.GetRenderSize().Width,
		panel3.GetRenderSize().Height,
	)
	return columns + "\n" + helpText + "\n" + sizes + "\n"
}
