package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// RenderBorderWithTitle renders content inside a bordered box using boxStyle,
// but replaces the top border with one that has title embedded in it.
// For example:
//
//	╭─Title─────────────────────────────────────────────────────╮
//	│ ...content...                                             │
//	╰───────────────────────────────────────────────────────────╯
//
// Note that using a style that has a border changes the expected output width and height.
// A style with a border, a width and a height of 80x24 will result in an output
// that is 82x26 because the border is rendered around the 80x24 text.
func RenderBorderWithTitle(boxStyle lipgloss.Style, title string, titleStyle lipgloss.Style, content string) string {

	// Let lipgloss render the box normally but without the top border.
	bodyStyle := boxStyle.BorderTop(false)
	body := bodyStyle.Render(content)

	// Measure the real rendered width.
	// Use this width for renderoug out our manual top border line.
	width := lipgloss.Width(body)

	top := renderTopBorderWithTitle(boxStyle, title, titleStyle, width)

	return top + "\n" + body
}

type titledBorderInfo struct {
	borderTopLeftWidth  int // ╭
	leadWidth           int // -
	titleWidth          int // 11 for "hello world"
	trailWidth          int // -
	borderTopRightWidth int // ╮
}

func (info titledBorderInfo) Length() int {
	length := info.borderTopLeftWidth +
		info.leadWidth +
		info.titleWidth +
		info.trailWidth +
		info.borderTopRightWidth
	return length
}

// Shrink shrinks the length of 1 element in a given order.
// Returns true when succesfull. Returns false otherwise.
func (info *titledBorderInfo) Shrink() bool {
	// Shorten trail when trail is longer than lead
	if info.trailWidth > 0 && info.trailWidth > info.leadWidth {
		info.trailWidth--
		return true
	}

	// Shorten trail when lead and trail are the same length
	if info.trailWidth > 0 && info.trailWidth == info.leadWidth {
		info.trailWidth--
		return true
	}

	// Shorten lead when lead is longer than trail
	if info.leadWidth > 0 && info.leadWidth > info.trailWidth {
		info.leadWidth--
		return true
	}

	// Shorten the title
	if info.titleWidth > 0 {
		info.titleWidth--
		return true
	}

	// only left and right corners are remaining
	return false
}

// ShrinkFit shrinks the length until it fits in the given width.
// Returns true when succesfull. Returns false otherwise.
func (info *titledBorderInfo) ShrinkFit(width int) bool {
	copy := *info

	// while copy.Length() > width
	for copy.Length() > width {
		if !copy.Shrink() {
			break // stop when we failed to Shrink
		}
	}

	// Did we succeed ?
	if copy.Length() <= width {
		*info = copy
		return true
	}
	return false
}

// countEdgeSpaces counts consecutive leading and trailing spaces or tabs.
func countEdgeSpaces(s string) (leading int, trailing int) {
	// Count leading spaces/tabs
	for _, r := range s {
		if r == ' ' || r == '\t' {
			leading++
		} else {
			break
		}
	}

	// If the string is only spaces/tabs, leading equals the length of the string
	if leading == len(s) {
		return leading, leading
	}

	// Count trailing spaces/tabs by iterating backwards using rune indices
	runes := []rune(s)
	for i := len(runes) - 1; i >= 0; i-- {
		if runes[i] == ' ' || runes[i] == '\t' {
			trailing++
		} else {
			break
		}
	}

	return leading, trailing
}

// shrinkStyle shrinks an element of the style by 1.
// Returns true when succesfull. Returns false otherwise.
func shrinkStyle(style *lipgloss.Style) bool {
	// Remove left/right borders, if any
	if style.GetBorderLeft() || style.GetBorderRight() {
		*style = style.UnsetBorderLeft().UnsetBorderRight()
		return true
	}

	paddingLeft := style.GetPaddingLeft()
	paddingRight := style.GetPaddingRight()

	// Remove padding on right if right>left or left == right
	if paddingRight > 0 && (paddingRight > paddingLeft || paddingLeft == paddingRight) {
		paddingRight--
		*style = style.PaddingRight(paddingRight)
		return true
	}

	// Remove padding on left if left>right or left == right
	if paddingLeft > 0 && (paddingLeft > paddingRight || paddingLeft == paddingRight) {
		paddingLeft--
		*style = style.PaddingLeft(paddingLeft)
		return true
	}

	return false
}

// shrinkString shrinks an element of the string by 1.
// Returns true when succesfull. Returns false otherwise.
func shrinkString(s *string) bool {
	// Is there spaces in the string itself ? If so, trim it
	whiteSpaceLeft, whiteSpaceRight := countEdgeSpaces(*s)

	// Remove white space on right if right>left or left == right
	if whiteSpaceRight > 0 && (whiteSpaceRight > whiteSpaceLeft || whiteSpaceLeft == whiteSpaceRight) {
		runes := []rune(*s)
		*s = string(runes[:len(runes)-1]) // remove last rune
		return true
	}

	// Remove white space on left if left>right or left == right
	if whiteSpaceLeft > 0 && (whiteSpaceLeft > whiteSpaceRight || whiteSpaceLeft == whiteSpaceRight) {
		runes := []rune(*s)
		*s = string(runes[1:]) // remove first rune
		return true
	}

	// Truncate the string
	runes := []rune(*s)
	if len(runes) > 0 {
		*s = string(runes[:len(runes)-1]) // remove last rune
		return true
	}

	return false
}

// shrinkFit shrinks the style then the title to match the given witdh.
// Returns true when succesfull. Returns false otherwise.
func shrinkFit(style *lipgloss.Style, title *string, width int) bool {
	// Check now
	renderedTitle := style.Render(*title)
	if lipgloss.Width(renderedTitle) <= width {
		return true
	}

	// while shrinkStyle(style) == true
	for shrinkStyle(style) == true {
		// Check now
		renderedTitle := style.Render(*title)
		if lipgloss.Width(renderedTitle) <= width {
			return true
		}
	}

	// while shrinkStyle(style) == true
	for shrinkString(title) == true {
		// Check now
		renderedTitle := style.Render(*title)
		if lipgloss.Width(renderedTitle) <= width {
			return true
		}
	}

	return false
}

// renderTopBorderWithTitle renders just the top border line of a bordered style using a custom the styled title.
func renderTopBorderWithTitle(boxStyle lipgloss.Style, title string, titleStyle lipgloss.Style, width int) string {
	// Make the titleStyle safe for a 1 liner title
	titleStyle = titleStyle.
		MaxHeight(1).         // force 1 liner output
		UnsetWidth().         // do not automatically add padding to match a target width
		UnsetHeight().        // do not automatically add padding to match a target height
		UnsetBorderTop().     // disable top border
		UnsetBorderBottom().  // disable bottom border
		UnsetPaddingTop().    // disable top padding
		UnsetPaddingBottom(). // disable bottom padding
		UnsetMargins()        // disable margins

	border := boxStyle.GetBorderStyle()

	// Style for the plain parts of the border (corners + fill runes).
	borderRender := lipgloss.NewStyle().
		Foreground(boxStyle.GetBorderTopForeground()).
		Background(boxStyle.GetBorderTopBackground())

	borderTopLeft := borderRender.Render(border.TopLeft)
	borderTopRight := borderRender.Render(border.TopRight)

	var info titledBorderInfo

	// Compute width of multiple elements
	// totalWidth := width
	info.borderTopLeftWidth = lipgloss.Width(borderTopLeft)
	info.borderTopRightWidth = lipgloss.Width(borderTopRight)
	cornersWidth := info.borderTopLeftWidth + info.borderTopRightWidth
	middleWidth := width - cornersWidth
	if middleWidth < 0 {
		middleWidth = 0
	}

	// Define constants & settings
	const defaultTitlePositionOnBorder = 2 // a value of 1 means there's top-left corner, a single `─` character then the title

	// Disable the title if no title
	// or when total width is too narrow
	if title == "" || middleWidth < 0 {
		output := borderTopLeft + borderRender.Render(strings.Repeat(border.Top, middleWidth)) + borderTopRight
		return output
	}

	info.titleWidth = lipgloss.Width(titleStyle.Render(title))

	// Setting default values
	info.leadWidth = defaultTitlePositionOnBorder
	info.trailWidth = width - cornersWidth - info.leadWidth - info.titleWidth
	if info.trailWidth < 0 {
		info.trailWidth = 0
	}

	// Force trail to be as big as the lead for cosmetic reasons
	if info.trailWidth < info.leadWidth {
		// Make the trail as big as the lead.
		// This potentially make it bigger than the target width.
		// The shrinking algorithm will take care of this.
		info.trailWidth = info.leadWidth
	}

	// Can we render the full title (including its offset position) on the top border?
	if info.Length() > width {
		// We can't. Target Width is too short.
		// Shrink something

		success := info.ShrinkFit(width)
		if !success {
			// render the minimum broder possible
			output := borderTopLeft + borderTopRight
			return output
		}

		// Yes we can.

		// Is the title target width is smaller than our native rendered title width ?
		if info.titleWidth < lipgloss.Width(titleStyle.Render(title)) {
			// No. We need to shorten the title's rendered width

			success = shrinkFit(&titleStyle, &title, info.titleWidth)
			if !success {
				// render the minimum broder possible
				output := borderTopLeft + borderTopRight
				return output
			}
		}
	}

	// define strings printed before and after the title
	lead := borderRender.Render(strings.Repeat(border.Top, info.leadWidth))
	trail := borderRender.Render(strings.Repeat(border.Top, info.trailWidth))

	renderedTitle := titleStyle.Render(title)

	return borderTopLeft + lead + renderedTitle + trail + borderTopRight
}
