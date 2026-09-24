package main_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/stretchr/testify/require"
)

var harryPotterBookDescription = "Harry Potter and the Philosopher's Stone (published in the United States as Harry Potter and the Sorcerer's Stone) is the fantasy novel that launched the globally acclaimed series by British author J. K. Rowling. The novel introduces Harry Potter, a young boy who discovers on his eleventh birthday that he is an orphaned wizard with a mysterious past, setting the stage for one of the most successful franchises in literary and cinematic history.\n\nRaised by his abusive aunt, uncle, and cousin, Harry has lived a miserable existence sleeping in a cupboard under the stairs. Everything changes when he receives a acceptance letter to Hogwarts School of Witchcraft and Wizardry, delivered by a half-giant named Rubeus Hagrid. Harry learns that his parents were powerful magical figures murdered by the dark wizard Lord Voldemort, and that Harry miraculously survived Voldemort's killing curse as an infant, leaving him with a lightning-bolt scar and legendary status in the wizarding world. At Hogwarts, Harry makes lifelong friends in Ron Weasley and Hermione Granger, and begins his education in magic. However, strange events at the school lead the trio to discover that the Philosopher's Stone—a magical object granting immortality—is hidden within the castle and under threat. Believing a hostile professor is attempting to steal it for the weakened Voldemort, Harry and his friends navigate a series of deadly magical obstacles to protect the stone, confronting the true agent of evil in a dramatic final showdown."

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

func TestFooBar(t *testing.T) {
	/*
		Viewport rules
			1. A viewport truncates its content horizontally and vertically to the size of the viewport.
			   Content that is wider than the viewport width or extends past the viewport height is simply clipped.
			2. If a Style matching the viewport's width/height is set, the viewport's View() method still won't
			   wrap text. Instead, Lipgloss will enforce those dimensions as a hard truncation boundary on the rendered output.
			3. Keep the viewport's style strictly for background colors, borders, or text colors if needed.
			4. Explicitly wrap the text manually to the viewport's target width then call SetContent().
	*/

	vp := viewport.New(80, 10)

	// Do not set a .Width() and .Height() otherwise Lipgloss will hard-clip the right edge.
	// Keep vp.Style strictly for background colors, borders, or text colors if needed.
	vp.Style = lipgloss.NewStyle()
	//	.Border(lipgloss.RoundedBorder())

	// Explicitly wrap the text to the viewport's target width first
	wrappedText := lipgloss.NewStyle().Width(vp.Width).Render(harryPotterBookDescription)

	// Set the pre-wrapped multi-line text into the viewport
	vp.SetContent(wrappedText)

	// Move down 1 row
	keyDownMsg := tea.KeyMsg{
		Type:  tea.KeyDown,
		Runes: []rune{},
		Alt:   false,
	}
	//var cmd tea.Cmd
	vp /*cmd*/, _ = vp.Update(keyDownMsg)
	vp.ScrollDown(1) // again

	// Render the view
	view := vp.View()

	// Strip ANSI colors/styles so we can perform reliable structural string assertions
	actualOutput := StripStyles(view)

	expectedOutput := []string{
		"globally acclaimed series by British author J. K. Rowling. The novel introduces ",
		"Harry Potter, a young boy who discovers on his eleventh birthday that he is an  ",
		"orphaned wizard with a mysterious past, setting the stage for one of the most   ",
		"successful franchises in literary and cinematic history.                        ",
		"                                                                                ",
		"Raised by his abusive aunt, uncle, and cousin, Harry has lived a miserable      ",
		"existence sleeping in a cupboard under the stairs. Everything changes when he   ",
		"receives a acceptance letter to Hogwarts School of Witchcraft and Wizardry,     ",
		"delivered by a half-giant named Rubeus Hagrid. Harry learns that his parents    ",
		"were powerful magical figures murdered by the dark wizard Lord Voldemort, and   ",
	}

	boxSymbol := "•"
	actualOutputBoxed := BoxOutputString(actualOutput, boxSymbol)
	expectedOutputBoxed := BoxOutputSlices(expectedOutput, boxSymbol)
	t.Logf("\nExpected test output:\n%s\n\nActual test output:\n%s", expectedOutputBoxed, actualOutputBoxed)

	err := AssertViewOutputLines(expectedOutput, actualOutput)
	require.NoError(t, err)
}
