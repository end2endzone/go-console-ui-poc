package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
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
