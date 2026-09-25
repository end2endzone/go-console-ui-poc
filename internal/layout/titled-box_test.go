package layout_test

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/end2endzone/go-console-ui-poc/internal/layout"
	"github.com/end2endzone/go-console-ui-poc/internal/lipglossutil"
	"github.com/stretchr/testify/require"
)

var loremIpsumDescription = "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum."

func TestRenderBorderWithTitle(t *testing.T) {
	t.Run("Test width/height, default title", func(t *testing.T) {
		boxStyle := lipgloss.NewStyle().
			Width(40).
			Height(10).
			Border(lipgloss.RoundedBorder())

		title := "mytitle"
		titleStyle := lipgloss.NewStyle()
		content := loremIpsumDescription

		// Act
		actualOutput := layout.RenderBorderWithTitle(boxStyle, title, titleStyle, content)
		actualOutput = lipglossutil.StripStyles(actualOutput)

		actualOutputWidth := lipgloss.Width(actualOutput)
		actualOutputHeight := lipgloss.Height(actualOutput)

		expectedOutput := []string{
			"╭──mytitle───────────────────────────────╮",
			"│Lorem ipsum dolor sit amet, consectetur │",
			"│adipiscing elit, sed do eiusmod tempor  │",
			"│incididunt ut labore et dolore magna    │",
			"│aliqua. Ut enim ad minim veniam, quis   │",
			"│nostrud exercitation ullamco laboris    │",
			"│nisi ut aliquip ex ea commodo consequat.│",
			"│Duis aute irure dolor in reprehenderit  │",
			"│in voluptate velit esse cillum dolore eu│",
			"│fugiat nulla pariatur. Excepteur sint   │",
			"│occaecat cupidatat non proident, sunt in│",
			"│culpa qui officia deserunt mollit anim  │",
			"│id est laborum.                         │",
			"╰────────────────────────────────────────╯",
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
		require.Equal(t, expectedWidth, expectedWidth, boxStyle.GetWidth()+2*borderWidth)
		require.Equal(t, expectedHeight, expectedHeight, boxStyle.GetHeight()+2*borderWidth)

		// Assert width/height of rendered output matches the expected size
		require.Equal(t, expectedWidth, expectedWidth, actualOutputWidth)
		require.Equal(t, expectedHeight, expectedHeight, actualOutputHeight)

		// Assert rendered output
		err := lipglossutil.AssertViewOutputLines(expectedOutput, actualOutput)
		require.NoError(t, err)
	})
}
