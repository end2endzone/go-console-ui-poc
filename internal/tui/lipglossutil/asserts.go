package lipglossutil

import (
	"fmt"
	"strings"
)

func AssertViewOutputString(expected string, actual string) error {
	lines := strings.Split(expected, "\n")
	err := AssertViewOutputLines(lines, actual)
	return err
}

func AssertViewOutputLines(expectedLines []string, actual string) error {
	actualLines := strings.Split(actual, "\n")

	expectedWidth := GetLipglossWidthFromLines(expectedLines)
	actualWidth := GetLipglossWidthFromLines(actualLines)

	expectedHeight := GetLipglossHeightFromLines(expectedLines)
	actualHeight := GetLipglossHeightFromLines(actualLines)

	if expectedWidth != actualWidth {
		return fmt.Errorf("expectedWidth != actualWidth, expecting %d, got %d", expectedWidth, actualWidth)
	}
	if expectedHeight != actualHeight {
		return fmt.Errorf("expectedHeight != actualHeight, expecting %d, got %d", expectedHeight, actualHeight)
	}

	// compare line by line
	for i := range actualLines {
		expectedLine := expectedLines[i]
		actualLine := actualLines[i]
		if expectedLine != actualLine {
			return fmt.Errorf("line %d, expectedLine != actualLine, expecting `%s`, got `%s`", i, expectedLine, actualLine)
		}
	}

	return nil
}
