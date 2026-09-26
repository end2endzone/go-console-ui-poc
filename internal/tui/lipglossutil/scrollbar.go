package lipglossutil

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var ThumbSymbol = "█"
var TrackSymbol = "│"

type ScrollBarModel struct {
	ThumbSymbol string
	TrackSymbol string
	ThumbStyle  lipgloss.Style
	TrackStyle  lipgloss.Style
}

func NewScrollBarModel() ScrollBarModel {
	// Initialize with zero size safely
	return ScrollBarModel{
		ThumbSymbol: ThumbSymbol,
		TrackSymbol: TrackSymbol,
		ThumbStyle:  lipgloss.NewStyle().Foreground(lipgloss.ANSIColor(8)).Bold(true),
		TrackStyle:  lipgloss.NewStyle().Foreground(lipgloss.ANSIColor(8)),
	}
}

func (m *ScrollBarModel) Width() int {
	renderedThumbSymbol := m.ThumbStyle.Render(m.ThumbSymbol)
	renderedTrackSymbol := m.TrackStyle.Render(m.TrackSymbol)
	thumbWidth := lipgloss.Width(renderedThumbSymbol)
	trackWidth := lipgloss.Width(renderedTrackSymbol)

	scrollBarWidth := 1 + max(thumbWidth, trackWidth) // a space then the scrollbar

	return scrollBarWidth
}

func (m *ScrollBarModel) RenderLinesContent(lines []string, thumbIndex int) string {
	renderedThumbSymbol := m.ThumbStyle.Render(m.ThumbSymbol)
	renderedTrackSymbol := m.TrackStyle.Render(m.TrackSymbol)

	// Append scrollbar characters at the end of each displayed line
	var buffer strings.Builder
	for i, line := range lines {
		var scrollSymbol string
		if i == thumbIndex {
			scrollSymbol = renderedThumbSymbol
		} else {
			scrollSymbol = renderedTrackSymbol
		}

		// Print the line itself and then the scrollbar
		buffer.WriteString(fmt.Sprintf("%s %s", line, scrollSymbol))

		isLast := (i + 1) == len(lines)
		if !isLast {
			buffer.WriteString("\n")
		}
	}

	output := buffer.String()

	return output
}
