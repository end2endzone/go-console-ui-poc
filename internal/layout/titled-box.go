package layout

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

	topLeft := borderRender.Render(border.TopLeft)
	topRight := borderRender.Render(border.TopRight)

	// Compute width of multiple elements
	// totalWidth := width
	cornersWidth := lipgloss.Width(border.TopLeft) + lipgloss.Width(border.TopRight)
	middleWidth := width - cornersWidth
	if title == "" || middleWidth < 0 {
		// If border is too narrow do not insert a title
		return topLeft + borderRender.Render(strings.Repeat(border.Top, middleWidth)) + topRight
	}
	if middleWidth < 0 {
		middleWidth = 0
	}
	titleWidth := lipgloss.Width(titleStyle.Render(title))

	// Can we render the full title (including its offset position) on the top border?
	titlePositionOnBorder := 2 // a value of 1 means there's top-left corner, a single `─` character then the title
	trailWidth := middleWidth - titlePositionOnBorder - titleWidth
	if trailWidth < 0 {
		// Too short

		// Try to shorten titlePositionOnBorder first?
		// While (trailWidth < 0 && titlePositionOnBorder > 0)
		for trailWidth < 0 && titlePositionOnBorder > 0 {
			titlePositionOnBorder--
			trailWidth = middleWidth - titlePositionOnBorder - titleWidth
		}

		// Still too short ?
		if trailWidth < 0 {
			// We must truncate the title until it does.
			maxTitleWidth := middleWidth - titlePositionOnBorder
			if maxTitleWidth < 0 {
				maxTitleWidth = 0
			}

			// Truncate title
			title = lipgloss.NewStyle().MaxWidth(maxTitleWidth).Render(title)
			titleWidth = lipgloss.Width(title)
			trailWidth = middleWidth - titlePositionOnBorder - titleWidth
			if trailWidth < 0 {
				trailWidth = 0
			}
		}
	}

	// define strings printed before and after the title
	lead := borderRender.Render(strings.Repeat(border.Top, titlePositionOnBorder))
	trail := borderRender.Render(strings.Repeat(border.Top, trailWidth))

	styledTitle := titleStyle.Render(title)

	return topLeft + lead + styledTitle + trail + topRight
}
