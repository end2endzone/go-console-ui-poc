package lipglossutil

import "github.com/charmbracelet/lipgloss"

func GetExpectedOutputBlockWidth(output []string) int {
	totalWidth := -1
	for i := range output {
		w := lipgloss.Width(output[i])
		if w > totalWidth {
			totalWidth = w
		}
	}
	return totalWidth
}

func GetLipglossWidthFromLines(lines []string) int {
	width := 0
	for i := range lines {
		w := lipgloss.Width(lines[i])
		if w > width {
			width = w
		}
	}
	return width
}

func GetLipglossHeightFromLines(lines []string) int {
	height := len(lines)
	return height
}
