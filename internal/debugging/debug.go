package debugging

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

// DumpStringToFile writes a debug string to the given file path.
func DumpStringToFile(filename string, data string) {
	// 0644 is the file permission mode (read/write for owner, read-only for others)
	err := os.WriteFile(filename, []byte(data), 0644)
	if err != nil {
		err2 := fmt.Errorf("Failed to dump debug string to file: %v", err)
		panic(err2)
	}
}

// AppendStringToFile appends a debug string to the given file path.
func AppendStringToFile(filename string, data string) {
	// 0644 gives read/write permissions to the owner, and read-only to others
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	// Write the string to the file
	if _, err := file.WriteString(data); err != nil {
		panic(err)
	}
}

// GetCurrentLocation returns the file name and line number of where it was called.
func GetCurrentLocation() string {
	// skip = 1 looks at the function calling getCurrentLocation
	_, file, line, ok := runtime.Caller(1)
	if !ok {
		return "unknown:0"
	}

	// Optional: Use filepath.Base to get "main.go" instead of the full absolute path
	shortFile := filepath.Base(file)

	return fmt.Sprintf("%s:%d", shortFile, line)
}

func RenderAllColors() string {
	var out string
	out = "Backgrounds: "
	for i := 0; i < 16; i++ {
		style := lipgloss.NewStyle().
			Background(lipgloss.ANSIColor(i)).
			Bold(true)
		out += " " + style.Render(fmt.Sprintf(" %d ", i)) + " "
	}
	out += "\n"
	out += "Foreground: "
	for i := 0; i < 16; i++ {
		style := lipgloss.NewStyle().
			Foreground(lipgloss.ANSIColor(i)).
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

// DumpRenderingWithoutStylesToFile remove styles and writes a debug string to the given file path.
func DumpRenderingWithoutStylesToFile(filename string, data string) {
	noStyles := StripStyles(data)

	// 0644 is the file permission mode (read/write for owner, read-only for others)
	err := os.WriteFile(filename, []byte(noStyles), 0644)
	if err != nil {
		err2 := fmt.Errorf("Failed to dump debug string to file: %v", err)
		panic(err2)
	}
}
