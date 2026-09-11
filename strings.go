package main

import "strings"

// IndexN returns the index of the nth occurrence of substr in s.
// n is 1-based (e.g., n=1 for the first occurrence).
// It returns -1 if n <= 0, substr is empty, or the nth occurrence is not found.
func StringsIndexN(s, substr string, n int) int {
	if n <= 0 || substr == "" {
		return -1
	}

	currentIndex := 0
	for i := 0; i < n; i++ {
		// Search the remaining part of the string
		matchOffset := strings.Index(s[currentIndex:], substr)
		if matchOffset == -1 {
			return -1 // Substring not found enough times
		}

		// Update the cumulative index position
		if i < n-1 {
			currentIndex += matchOffset + len(substr)
		} else {
			currentIndex += matchOffset
		}
	}

	return currentIndex
}

// truncateTextWidth Truncates text with trailing ellipsis to fit a given maximum length.
func truncateTextWidth(text string, maxLen int) string {
	if maxLen <= 3 {
		return "..."
	}
	runes := []rune(text)
	if len(runes) > maxLen {
		return string(runes[:maxLen-3]) + "..."
	}
	return text
}

// truncateTextHeight Truncates text with trailing ellipsis to fit a given maximum number of lines.
func truncateTextHeight(text string, maxHeight int) string {
	if maxHeight <= 1 {
		return "[...]"
	}

	// Do we need to truncate?
	linesCount := strings.Count(text, "\n") + 1 // +1 since the last sentence is a line after the last newline character
	if linesCount <= maxHeight {
		return text
	}

	// Too many lines of text
	lines := strings.Split(text, "\n")

	// Truncate lines over maxHeight, including the maxHeight line
	lines = lines[:maxHeight-1] // -1 since we want to override the latest visible line

	// Add a "continue" line endding
	lines = append(lines, "[...]")

	newText := strings.Join(lines, "\n")
	return newText
}
