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

func TestViewportBehavior(t *testing.T) {
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

	// Setting content again will not move the viewport's scrolled cursor.
	vp.SetContent(wrappedText)

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
		Width(vp.Width - scrollBarWidth /*- borderWidthIfPresent*/).
		//Height(vp.Width - borderHeightIfPresent).
		Render(content)

	// Set the pre-wrapped multi-line text into the viewport
	/*before := vp.ScrollPercent()
	if before == 123.4567 {
		return ""
	}*/
	vp.SetContent(wrappedText)
	/*after := vp.ScrollPercent()
	if after == 123.4567 {
		return ""
	}*/

	// Temporary patch the viewport's width and height to prevent the viewport from adding additionnal padding.
	// Then render the View().
	vp.Width -= scrollBarWidth /*+ borderWidthIfPresent*/
	//vp.Height -= borderHeightIfPresent
	tmpViewportView := vp.View()
	vp.Width += scrollBarWidth /*+ borderWidthIfPresent*/
	//vp.Height += borderHeightIfPresent

	tmpViewportViewWidth := lipgloss.Width(tmpViewportView)
	if tmpViewportViewWidth == 1234567 {
		return ""
	}

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
	newViewportViewWidth := lipgloss.Width(newViewportView)
	if newViewportViewWidth == 1234567 {
		return ""
	}

	// Add a border if the original viewport was bordered
	if bordered {
		style := vp.Style.
			Border(previousBorder).
			Width(vp.Width).
			Height(vp.Height)

		// Add a border around previously rendered text
		newViewportView = style.Render(newViewportView)
	}

	finalViewportViewWidth := lipgloss.Width(newViewportView)
	if finalViewportViewWidth == 1234567 {
		return ""
	}

	return newViewportView
}

func TestViewportViewWithVerticalScrollBar(t *testing.T) {
	vp := viewport.New(80, 10)

	content := harryPotterBookDescription

	t.Run("No scroll at all", func(t *testing.T) {
		// Render the viewport with scrollbar
		view := ViewportViewWithVerticalScrollBar(&vp, content)

		// Strip ANSI colors/styles so we can perform reliable structural string assertions
		actualOutput := StripStyles(view)

		expectedOutput := []string{
			"Harry Potter and the Philosopher's Stone (published in the United States as    █",
			"Harry Potter and the Sorcerer's Stone) is the fantasy novel that launched the  |",
			"globally acclaimed series by British author J. K. Rowling. The novel           |",
			"introduces Harry Potter, a young boy who discovers on his eleventh birthday    |",
			"that he is an orphaned wizard with a mysterious past, setting the stage for    |",
			"one of the most successful franchises in literary and cinematic history.       |",
			"                                                                               |",
			"Raised by his abusive aunt, uncle, and cousin, Harry has lived a miserable     |",
			"existence sleeping in a cupboard under the stairs. Everything changes when he  |",
			"receives a acceptance letter to Hogwarts School of Witchcraft and Wizardry,    |",
		}

		boxSymbol := "•"
		actualOutputBoxed := BoxOutputString(actualOutput, boxSymbol)
		expectedOutputBoxed := BoxOutputSlices(expectedOutput, boxSymbol)
		t.Logf("\nExpected test output:\n%s\n\nActual test output:\n%s", expectedOutputBoxed, actualOutputBoxed)

		// Assert width of View() output matches the viewport's size
		require.Equal(t, vp.Width, lipgloss.Width(actualOutput))

		err := AssertViewOutputLines(expectedOutput, actualOutput)
		require.NoError(t, err)
	})

	t.Run("Scrolled 3 lines", func(t *testing.T) {
		vp.ScrollDown(3)

		// Render the viewport with scrollbar
		view := ViewportViewWithVerticalScrollBar(&vp, content)

		// Strip ANSI colors/styles so we can perform reliable structural string assertions
		actualOutput := StripStyles(view)

		expectedOutput := []string{
			"introduces Harry Potter, a young boy who discovers on his eleventh birthday    |",
			"that he is an orphaned wizard with a mysterious past, setting the stage for    |",
			"one of the most successful franchises in literary and cinematic history.       █",
			"                                                                               |",
			"Raised by his abusive aunt, uncle, and cousin, Harry has lived a miserable     |",
			"existence sleeping in a cupboard under the stairs. Everything changes when he  |",
			"receives a acceptance letter to Hogwarts School of Witchcraft and Wizardry,    |",
			"delivered by a half-giant named Rubeus Hagrid. Harry learns that his parents   |",
			"were powerful magical figures murdered by the dark wizard Lord Voldemort, and  |",
			"that Harry miraculously survived Voldemort's killing curse as an infant,       |",
		}

		boxSymbol := "•"
		actualOutputBoxed := BoxOutputString(actualOutput, boxSymbol)
		expectedOutputBoxed := BoxOutputSlices(expectedOutput, boxSymbol)
		t.Logf("\nExpected test output:\n%s\n\nActual test output:\n%s", expectedOutputBoxed, actualOutputBoxed)

		// Assert width of View() output matches the viewport's size
		require.Equal(t, vp.Width, lipgloss.Width(actualOutput))

		err := AssertViewOutputLines(expectedOutput, actualOutput)
		require.NoError(t, err)
	})

	t.Run("At the bottom", func(t *testing.T) {
		vp.GotoBottom()

		// Render the viewport with scrollbar
		view := ViewportViewWithVerticalScrollBar(&vp, content)

		// Strip ANSI colors/styles so we can perform reliable structural string assertions
		actualOutput := StripStyles(view)

		expectedOutput := []string{
			"were powerful magical figures murdered by the dark wizard Lord Voldemort, and  |",
			"that Harry miraculously survived Voldemort's killing curse as an infant,       |",
			"leaving him with a lightning-bolt scar and legendary status in the wizarding   |",
			"world. At Hogwarts, Harry makes lifelong friends in Ron Weasley and Hermione   |",
			"Granger, and begins his education in magic. However, strange events at the     |",
			"school lead the trio to discover that the Philosopher's Stone—a magical object |",
			"granting immortality—is hidden within the castle and under threat. Believing a |",
			"hostile professor is attempting to steal it for the weakened Voldemort, Harry  |",
			"and his friends navigate a series of deadly magical obstacles to protect the   |",
			"stone, confronting the true agent of evil in a dramatic final showdown.        █",
		}

		boxSymbol := "•"
		actualOutputBoxed := BoxOutputString(actualOutput, boxSymbol)
		expectedOutputBoxed := BoxOutputSlices(expectedOutput, boxSymbol)
		t.Logf("\nExpected test output:\n%s\n\nActual test output:\n%s", expectedOutputBoxed, actualOutputBoxed)

		// Assert width of View() output matches the viewport's size
		require.Equal(t, vp.Width, lipgloss.Width(actualOutput))

		err := AssertViewOutputLines(expectedOutput, actualOutput)
		require.NoError(t, err)
	})

	t.Run("At the top, scrolled 3 lines, with borders", func(t *testing.T) {
		vp.GotoTop()
		vp.ScrollDown(3)

		vp.Style = vp.Style.Border(lipgloss.RoundedBorder())

		// Get the expected bordered view output's width without any scrollbar
		officialView := vp.View()
		officialViewWidth := lipgloss.Width(officialView)

		// Assert width of View() output matches the viewport's size
		require.Equal(t, vp.Width, officialViewWidth)

		// Render the viewport with scrollbar
		require.True(t, IsViewportBordered(&vp)) // assert borders before the call
		view := ViewportViewWithVerticalScrollBar(&vp, content)
		require.True(t, IsViewportBordered(&vp)) // assert borders after the call

		// Strip ANSI colors/styles so we can perform reliable structural string assertions
		actualOutput := StripStyles(view)

		expectedOutput := []string{
			"╭──────────────────────────────────────────────────────────────────────────────╮",
			"│introduces Harry Potter, a young boy who discovers on his eleventh birthday  |│",
			"│that he is an orphaned wizard with a mysterious past, setting the stage for  █│",
			"│one of the most successful franchises in literary and cinematic history.     |│",
			"│                                                                             |│",
			"│Raised by his abusive aunt, uncle, and cousin, Harry has lived a miserable   |│",
			"│existence sleeping in a cupboard under the stairs. Everything changes when   |│",
			"│he receives a acceptance letter to Hogwarts School of Witchcraft and         |│",
			"│Wizardry, delivered by a half-giant named Rubeus Hagrid. Harry learns that   |│",
			"╰──────────────────────────────────────────────────────────────────────────────╯",
		}

		boxSymbol := "•"
		actualOutputBoxed := BoxOutputString(actualOutput, boxSymbol)
		expectedOutputBoxed := BoxOutputSlices(expectedOutput, boxSymbol)
		t.Logf("\nExpected test output:\n%s\n\nActual test output:\n%s", expectedOutputBoxed, actualOutputBoxed)

		// Assert width of View() output matches the viewport's size
		require.Equal(t, vp.Width, lipgloss.Width(actualOutput))

		err := AssertViewOutputLines(expectedOutput, actualOutput)
		require.NoError(t, err)
	})

}
