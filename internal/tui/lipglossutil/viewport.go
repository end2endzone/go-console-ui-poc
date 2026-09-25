package lipglossutil

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
)

func IsViewportBordered(vp *viewport.Model) bool {
	if vp.Style.GetBorderTop() ||
		vp.Style.GetBorderBottom() ||
		vp.Style.GetBorderLeft() ||
		vp.Style.GetBorderRight() {
		return true
	}
	return false
}

func ViewportViewWithVerticalScrollBar(vp *viewport.Model, content string) string {
	if vp.Width == 0 || vp.Height == 0 {
		return ""
	}

	// Backup the viewport's
	previousStyle := vp.Style
	previousBorder := vp.Style.GetBorderStyle()
	previousWidth := vp.Width
	previousHeight := vp.Height
	bordered := IsViewportBordered(vp)

	const borderWidth = 1 // actual width of each side of the border

	// Remember that we need to render the inside text smaller to account for final border to match the viewport's original width.
	borderWidthIfPresent := 0
	borderHeightIfPresent := 0

	// Remove the borders from the viewport's style
	if bordered {
		// Reduce the size of the viewport to be able to add a border at a later time.
		if vp.Style.GetBorderTop() {
			vp.Height -= borderWidth
			borderHeightIfPresent++
		}
		if vp.Style.GetBorderBottom() {
			vp.Height -= borderWidth
			borderHeightIfPresent++
		}
		if vp.Style.GetBorderLeft() {
			vp.Width -= borderWidth
			borderWidthIfPresent++
		}
		if vp.Style.GetBorderRight() {
			vp.Width -= borderWidth
			borderWidthIfPresent++
		}
		if vp.Height < 0 {
			vp.Height = 0
		}
		if vp.Width < 0 {
			vp.Width = 0
		}

		// Remove the border
		vp.Style = vp.Style.UnsetBorderStyle()
	}

	// Restore the original style and size at the end
	defer func() {
		if bordered {
			vp.Width = previousWidth
			vp.Height = previousHeight
			vp.Style = previousStyle
		}
	}()

	ThumbSymbol := "█"
	TrackSymbol := "|"
	scrollBarWidth := 1 + max(lipgloss.Width(ThumbSymbol), lipgloss.Width(TrackSymbol)) // a space then the symbol

	ThumbStyle := lipgloss.NewStyle().Foreground(lipgloss.ANSIColor(8)).Bold(true)
	TrackStyle := lipgloss.NewStyle().Foreground(lipgloss.ANSIColor(8))

	// Explicitly wrap the text to the viewport's target width - scrollBarWidth to add the right scrollbar.
	// Note: Do not set a specific height to limit the rendered output to a maximum of lines.
	// When rendering smaller/shorter content, the rendering process will add empty rows.
	// Sending these empty rows to the viewport will make them actually scrollable which is undesired.
	wrappedText := lipgloss.NewStyle().
		Width(vp.Width - scrollBarWidth).
		Render(content)

	// Set the pre-wrapped multi-line text into the viewport
	//before := vp.ScrollPercent()
	//if before == 123.4567 {
	//	return ""
	//}
	vp.SetContent(wrappedText)
	//after := vp.ScrollPercent()
	//if after == 123.4567 {
	//	return ""
	//}

	// Temporary patch the viewport's width to match the content style's width.
	// Without this, the viewport will add additionnal padding at the end of each row to match its width.
	// Then render the View().
	vp.Width -= scrollBarWidth
	tmpViewportView := vp.View()
	vp.Width += scrollBarWidth

	//tmpViewportViewWidth := lipgloss.Width(tmpViewportView)
	//if tmpViewportViewWidth == 1234567 {
	//	return ""
	//}

	// Split by line to be able to manipulate lines individually
	lines := strings.Split(tmpViewportView, "\n")
	vpHeight := len(lines) // number of line displayed
	if vpHeight == 0 {
		// If viewport content is empty, return immediately
		return tmpViewportView
	}

	// Determine scroll thumb position
	scrollPercent := vp.ScrollPercent()
	thumbPos := int(scrollPercent * float64(vpHeight-1))

	// Fix thumbPos if we are at the top most or botto mmost viewport
	if vp.AtBottom() {
		thumbPos = vpHeight - 1
	}
	if vp.AtTop() {
		thumbPos = 0
	}

	// Append scrollbar characters at the end of each displayed line
	var output strings.Builder
	for i, line := range lines {
		var scrollSymbol string
		if i == thumbPos {
			scrollSymbol = ThumbStyle.Render(ThumbSymbol)
		} else {
			scrollSymbol = TrackStyle.Render(TrackSymbol)
		}

		// Print the line itself and then the scrollbar
		output.WriteString(fmt.Sprintf("%s %s", line, scrollSymbol))

		isLast := (i + 1) == len(lines)
		if !isLast {
			output.WriteString("\n")
		}
	}

	newViewportView := output.String()
	//newViewportViewWidth := lipgloss.Width(newViewportView)
	//if newViewportViewWidth == 1234567 {
	//	return ""
	//}

	// Add a border if the original viewport was bordered
	if bordered {
		style := vp.Style.
			Border(previousBorder).
			Width(vp.Width).
			Height(vp.Height)

		// Add a border around previously rendered text
		newViewportView = style.Render(newViewportView)
	}

	//finalViewportViewWidth := lipgloss.Width(newViewportView)
	//if finalViewportViewWidth == 1234567 {
	//	return ""
	//}

	return newViewportView
}
