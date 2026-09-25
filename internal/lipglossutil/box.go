package lipglossutil

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func BoxOutputString(s string, boxSymbol string) string {
	lines := strings.Split(s, "\n")
	output := BoxOutputSlices(lines, boxSymbol)
	return output
}

func BoxOutputSlices(lines []string, boxSymbol string) string {
	emptyBox := boxSymbol + boxSymbol + "\n" + boxSymbol + boxSymbol
	if len(lines) == 0 {
		return emptyBox
	}
	length := GetExpectedOutputBlockWidth(lines)
	if length == -1 {
		return emptyBox
	}

	lines = append([]string(nil), lines...) // duplicate lines to prevent modifying original lines

	for i := range lines {
		lines[i] = boxSymbol + lines[i] + boxSymbol
	}

	headerFooter := strings.Repeat(boxSymbol, length+2*lipgloss.Width(boxSymbol))

	lines = append([]string{headerFooter}, lines...) // insert first
	lines = append(lines, headerFooter)              // insert last
	body := strings.Join(lines, "\n")
	return body
}
