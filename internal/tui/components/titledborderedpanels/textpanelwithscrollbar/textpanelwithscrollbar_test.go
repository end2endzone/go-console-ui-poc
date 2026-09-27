package textpanelwithscrollbar_test

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/end2endzone/go-console-ui-poc/internal/tui/components/titledborderedpanels/textpanelwithscrollbar"
	"github.com/end2endzone/go-console-ui-poc/internal/tui/lipglossutil"
	"github.com/stretchr/testify/require"
)

var loremIpsumDescription = "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum."

func TestTextPanel(t *testing.T) {
	t.Run("Test width and height, default title, custom scrollbar symbols", func(t *testing.T) {
		m := textpanelwithscrollbar.New()

		m.Title = "mytitle"
		m.TitleStyle = lipgloss.NewStyle().Bold(true)
		m.PanelStyle = lipgloss.NewStyle().Border(lipgloss.DoubleBorder())

		m.ScrollBar.ThumbSymbol = "▒▒▒"
		m.ScrollBar.TrackSymbol = " ┊ "
		//m.ScrollBar.ThumbStyle = lipgloss.NewStyle().Padding(0, 2)
		//m.ScrollBar.TrackStyle = lipgloss.NewStyle().Padding(0, 2)

		m.SetSize(40, 10)
		m.SetContent(loremIpsumDescription)

		// Act
		actualOutput := m.View()
		actualOutput = lipglossutil.StripStyles(actualOutput)

		actualOutputWidth := lipgloss.Width(actualOutput)
		actualOutputHeight := lipgloss.Height(actualOutput)

		expectedOutput := []string{
			"╔══mytitle═════════════════════════════╗",
			"║Lorem ipsum dolor sit amet,        ▒▒▒║",
			"║consectetur adipiscing elit, sed    ┊ ║",
			"║do eiusmod tempor incididunt ut     ┊ ║",
			"║labore et dolore magna aliqua. Ut   ┊ ║",
			"║enim ad minim veniam, quis nostrud  ┊ ║",
			"║exercitation ullamco laboris nisi   ┊ ║",
			"║ut aliquip ex ea commodo            ┊ ║",
			"║consequat. Duis aute irure dolor    ┊ ║",
			"╚══════════════════════════════════════╝",
		}

		boxSymbol := "•"
		actualOutputBoxed := lipglossutil.BoxOutputString(actualOutput, boxSymbol)
		expectedOutputBoxed := lipglossutil.BoxOutputSlices(expectedOutput, boxSymbol)
		t.Logf("\nExpected test output:\n%s\n\nActual test output:\n%s", expectedOutputBoxed, actualOutputBoxed)

		// Expected width/height should be 2 characters wider than the style's width/height.
		// When redering a bordered style, the style width/height dimension property is applied to the content, not the borders.
		// Border are added around the rendered content which makes the border 2 characters.
		const borderWidth = 1
		expectedWidth := 40 + 2*borderWidth
		expectedHeight := 10 + 2*borderWidth
		require.Equal(t, expectedWidth, expectedWidth, m.PanelStyle.GetWidth()+2*borderWidth)
		require.Equal(t, expectedHeight, expectedHeight, m.PanelStyle.GetHeight()+2*borderWidth)

		// Assert width/height of rendered output matches the expected size
		require.Equal(t, expectedWidth, expectedWidth, actualOutputWidth)
		require.Equal(t, expectedHeight, expectedHeight, actualOutputHeight)

		// Assert rendered output
		err := lipglossutil.AssertViewOutputLines(expectedOutput, actualOutput)
		require.NoError(t, err)
	})

	t.Run("Test scrollbar with padding, scrolled 4 down", func(t *testing.T) {
		m := textpanelwithscrollbar.New()

		m.Title = "mytitle"
		m.TitleStyle = lipgloss.NewStyle().Bold(true)
		m.PanelStyle = lipgloss.NewStyle().Border(lipgloss.RoundedBorder())

		m.SetSize(40, 10)
		m.SetContent(loremIpsumDescription)

		m.ScrollDown(4)

		m.ScrollBar.ThumbSymbol = "☺"
		m.ScrollBar.ThumbStyle = lipgloss.NewStyle().Padding(0, 3, 0, 0)
		m.ScrollBar.TrackStyle = lipgloss.NewStyle().Padding(0, 3, 0, 0)

		// Act
		actualOutput := m.View()
		actualOutput = lipglossutil.StripStyles(actualOutput)

		actualOutputWidth := lipgloss.Width(actualOutput)
		actualOutputHeight := lipgloss.Height(actualOutput)

		expectedOutput := []string{
			"╭──mytitle────────────────────────────────╮",
			"│minim veniam, quis nostrud           │   │",
			"│exercitation ullamco laboris nisi ut │   │",
			"│aliquip ex ea commodo consequat.     │   │",
			"│Duis aute irure dolor in             │   │",
			"│reprehenderit in voluptate velit     ☺   │",
			"│esse cillum dolore eu fugiat nulla   │   │",
			"│pariatur. Excepteur sint occaecat    │   │",
			"│cupidatat non proident, sunt in      │   │",
			"╰─────────────────────────────────────────╯",
		}

		boxSymbol := "•"
		actualOutputBoxed := lipglossutil.BoxOutputString(actualOutput, boxSymbol)
		expectedOutputBoxed := lipglossutil.BoxOutputSlices(expectedOutput, boxSymbol)
		t.Logf("\nExpected test output:\n%s\n\nActual test output:\n%s", expectedOutputBoxed, actualOutputBoxed)

		// Expected width/height should be 2 characters wider than the style's width/height.
		// When redering a bordered style, the style width/height dimension property is applied to the content, not the borders.
		// Border are added around the rendered content which makes the border 2 characters.
		const borderWidth = 1
		expectedWidth := 40 + 2*borderWidth
		expectedHeight := 10 + 2*borderWidth
		require.Equal(t, expectedWidth, expectedWidth, m.PanelStyle.GetWidth()+2*borderWidth)
		require.Equal(t, expectedHeight, expectedHeight, m.PanelStyle.GetHeight()+2*borderWidth)

		// Assert width/height of rendered output matches the expected size
		require.Equal(t, expectedWidth, expectedWidth, actualOutputWidth)
		require.Equal(t, expectedHeight, expectedHeight, actualOutputHeight)

		// Assert rendered output
		err := lipglossutil.AssertViewOutputLines(expectedOutput, actualOutput)
		require.NoError(t, err)
	})
}
