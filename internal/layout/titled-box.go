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
	titleWidth := lipgloss.Width(title)

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
