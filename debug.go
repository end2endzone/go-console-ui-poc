package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

// dumpToFile writes a debug string to the given file path.
func dumpToFile(filename string, data string) {
	// 0644 is the file permission mode (read/write for owner, read-only for others)
	err := os.WriteFile(filename, []byte(data), 0644)
	if err != nil {
		err2 := fmt.Errorf("Failed to dump debug string to file: %v", err)
		panic(err2)
	}
}

func RenderAllColors() string {
	var out string
	for i := 0; i < 16; i++ {
		style := lipgloss.NewStyle().
			Background(lipgloss.ANSIColor(i)).
			Bold(true)
		out += " " + style.Render(fmt.Sprintf(" %d ", i)) + " "
	}
	return out
}

func GetLongestLineInText(text string) (int, string) {
	lines := strings.Split(text, "\n")
	lenght := -1
	line := ""
	for _, tmp := range lines {
		if len([]rune(tmp)) > lenght {
			lenght = len([]rune(tmp))
			line = tmp
		}
	}
	return lenght, line
}

// StripStyles removes all ANSI escape sequences from a rendered string
func StripStyles(renderedView string) string {
	return string(ansi.Strip(renderedView))
}

// StripStylesAndDumpToFile remove styles and writes a debug string to the given file path.
func StripStylesAndDumpToFile(filename string, data string) {
	noStyles := StripStyles(data)

	// 0644 is the file permission mode (read/write for owner, read-only for others)
	err := os.WriteFile(filename, []byte(noStyles), 0644)
	if err != nil {
		err2 := fmt.Errorf("Failed to dump debug string to file: %v", err)
		panic(err2)
	}
}
