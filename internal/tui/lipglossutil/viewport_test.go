package lipglossutil_test

import (
	"testing"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/end2endzone/go-console-ui-poc/internal/tui/lipglossutil"
	"github.com/stretchr/testify/require"
)

var harryPotterBookDescription = "Harry Potter and the Philosopher's Stone (published in the United States as Harry Potter and the Sorcerer's Stone) is the fantasy novel that launched the globally acclaimed series by British author J. K. Rowling. The novel introduces Harry Potter, a young boy who discovers on his eleventh birthday that he is an orphaned wizard with a mysterious past, setting the stage for one of the most successful franchises in literary and cinematic history.\n\nRaised by his abusive aunt, uncle, and cousin, Harry has lived a miserable existence sleeping in a cupboard under the stairs. Everything changes when he receives a acceptance letter to Hogwarts School of Witchcraft and Wizardry, delivered by a half-giant named Rubeus Hagrid. Harry learns that his parents were powerful magical figures murdered by the dark wizard Lord Voldemort, and that Harry miraculously survived Voldemort's killing curse as an infant, leaving him with a lightning-bolt scar and legendary status in the wizarding world. At Hogwarts, Harry makes lifelong friends in Ron Weasley and Hermione Granger, and begins his education in magic. However, strange events at the school lead the trio to discover that the Philosopher's Stone—a magical object granting immortality—is hidden within the castle and under threat. Believing a hostile professor is attempting to steal it for the weakened Voldemort, Harry and his friends navigate a series of deadly magical obstacles to protect the stone, confronting the true agent of evil in a dramatic final showdown."

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
	actualOutput := lipglossutil.StripStyles(view)

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
	actualOutputBoxed := lipglossutil.BoxOutputString(actualOutput, boxSymbol)
	expectedOutputBoxed := lipglossutil.BoxOutputSlices(expectedOutput, boxSymbol)
	t.Logf("\nExpected test output:\n%s\n\nActual test output:\n%s", expectedOutputBoxed, actualOutputBoxed)

	err := lipglossutil.AssertViewOutputLines(expectedOutput, actualOutput)
	require.NoError(t, err)
}

func TestViewportViewWithVerticalScrollBar(t *testing.T) {
	t.Run("Default viewport behavior with border", func(t *testing.T) {
		vp := viewport.New(80, 10)
		content := harryPotterBookDescription

		vp.Style = vp.Style.Border(lipgloss.RoundedBorder())
		vp.SetContent(content)

		// Get the expected bordered view output's width without any scrollbar
		officialView := vp.View()
		officialViewWidth := lipgloss.Width(officialView)

		// Assert width of View() output matches the viewport's size
		require.Equal(t, vp.Width, officialViewWidth)
	})

	t.Run("No scroll at all", func(t *testing.T) {
		vp := viewport.New(80, 10)
		content := harryPotterBookDescription

		// Render the viewport with scrollbar
		view := lipglossutil.ViewportViewWithVerticalScrollBar(&vp, content)

		// Strip ANSI colors/styles so we can perform reliable structural string assertions
		actualOutput := lipglossutil.StripStyles(view)

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
		actualOutputBoxed := lipglossutil.BoxOutputString(actualOutput, boxSymbol)
		expectedOutputBoxed := lipglossutil.BoxOutputSlices(expectedOutput, boxSymbol)
		t.Logf("\nExpected test output:\n%s\n\nActual test output:\n%s", expectedOutputBoxed, actualOutputBoxed)

		// Assert width of View() output matches the viewport's size
		require.Equal(t, vp.Width, lipgloss.Width(actualOutput))

		err := lipglossutil.AssertViewOutputLines(expectedOutput, actualOutput)
		require.NoError(t, err)
	})

	t.Run("Scrolled 3 lines", func(t *testing.T) {
		vp := viewport.New(80, 10)
		content := harryPotterBookDescription

		// Render the viewport with scrollbar
		view := lipglossutil.ViewportViewWithVerticalScrollBar(&vp, content)

		vp.ScrollDown(3)

		// (render again)
		view = lipglossutil.ViewportViewWithVerticalScrollBar(&vp, content)

		// Strip ANSI colors/styles so we can perform reliable structural string assertions
		actualOutput := lipglossutil.StripStyles(view)

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
		actualOutputBoxed := lipglossutil.BoxOutputString(actualOutput, boxSymbol)
		expectedOutputBoxed := lipglossutil.BoxOutputSlices(expectedOutput, boxSymbol)
		t.Logf("\nExpected test output:\n%s\n\nActual test output:\n%s", expectedOutputBoxed, actualOutputBoxed)

		// Assert width of View() output matches the viewport's size
		require.Equal(t, vp.Width, lipgloss.Width(actualOutput))

		err := lipglossutil.AssertViewOutputLines(expectedOutput, actualOutput)
		require.NoError(t, err)
	})

	t.Run("At the bottom", func(t *testing.T) {
		vp := viewport.New(80, 10)
		content := harryPotterBookDescription

		// Render the viewport with scrollbar
		view := lipglossutil.ViewportViewWithVerticalScrollBar(&vp, content)

		vp.GotoBottom()

		// (render again)
		view = lipglossutil.ViewportViewWithVerticalScrollBar(&vp, content)

		// Strip ANSI colors/styles so we can perform reliable structural string assertions
		actualOutput := lipglossutil.StripStyles(view)

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
		actualOutputBoxed := lipglossutil.BoxOutputString(actualOutput, boxSymbol)
		expectedOutputBoxed := lipglossutil.BoxOutputSlices(expectedOutput, boxSymbol)
		t.Logf("\nExpected test output:\n%s\n\nActual test output:\n%s", expectedOutputBoxed, actualOutputBoxed)

		// Assert width of View() output matches the viewport's size
		require.Equal(t, vp.Width, lipgloss.Width(actualOutput))

		err := lipglossutil.AssertViewOutputLines(expectedOutput, actualOutput)
		require.NoError(t, err)
	})

	t.Run("At the top, scrolled 3 lines, with borders", func(t *testing.T) {
		vp := viewport.New(80, 10)
		content := harryPotterBookDescription

		vp.Style = vp.Style.Border(lipgloss.RoundedBorder())

		// Render the viewport with scrollbar
		view := lipglossutil.ViewportViewWithVerticalScrollBar(&vp, content)

		vp.ScrollDown(3)

		// (render again)
		require.True(t, lipglossutil.IsViewportBordered(&vp)) // assert borders before the call
		view = lipglossutil.ViewportViewWithVerticalScrollBar(&vp, content)
		require.True(t, lipglossutil.IsViewportBordered(&vp)) // assert borders after the call

		// Strip ANSI colors/styles so we can perform reliable structural string assertions
		actualOutput := lipglossutil.StripStyles(view)

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
		actualOutputBoxed := lipglossutil.BoxOutputString(actualOutput, boxSymbol)
		expectedOutputBoxed := lipglossutil.BoxOutputSlices(expectedOutput, boxSymbol)
		t.Logf("\nExpected test output:\n%s\n\nActual test output:\n%s", expectedOutputBoxed, actualOutputBoxed)

		// Assert width of View() output matches the viewport's size
		require.Equal(t, vp.Width, lipgloss.Width(actualOutput))

		err := lipglossutil.AssertViewOutputLines(expectedOutput, actualOutput)
		require.NoError(t, err)
	})

}
