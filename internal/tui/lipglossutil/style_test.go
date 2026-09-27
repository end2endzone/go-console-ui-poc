package lipglossutil_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/end2endzone/go-console-ui-poc/internal/tui/lipglossutil"
	"github.com/stretchr/testify/require"
)

const longInputText = "The Little Prince (Le Petit Prince) is a novella by French aristocrat, writer, and military aviator Antoine de Saint-Exupéry. First published in English and French in the United States in April 1943, it is one of the most translated and best-selling books in history. The story follows a young prince who visits various planets in space, including Earth, addressing themes of loneliness, friendship, love, and loss. Despite its style as a children's book, The Little Prince makes observations about life and human nature that resonate deeply with adult readers."

/*
Style rules:
	1. Content that has too many line and exceeds the maximum renderable area will overfill the height but not the width.
	2. A Style with a given Width and Height will fill the rendered output string with spaces to fill the given Width/Height.
	3. A Style with padding will shrink the content so that padding + content matches the style's Width and Height.
	   The style will try to split between words. If a phrase or a word is too long, it will cut the word in half and continue on the next line.
	4. A Border wraps the style's Width and Height rendering an output that is Width+2 wide and Height+2 high.
	5. A Style that has a MaxWidth will truncate whatever output that is longer than its MaxWidth, including the border,padding,etc.
	6. A Style that has a MaxHeight will truncate whatever output that is higher than its MaxHeight, including the border,padding,etc.
	7. A Style with a Margin will increase the size of a style over its border. It fill fill the margin area with spaces.
	8. A Style without a Width or a Height will not force word wrapping potentially resulting very wide output.
	   If padding is defined, the padding will be added arround the text area.
	   If a border is defined, the border will be added arround the padding and the text area.
	9. The Bubble Tea framework will truncate the right section of a rendering so that each line of a View() fit the terminal's width dimension.
   10. The Bubble Tea framework will truncate the top section of a rendering so that the last lines of a View() fit the terminal's height dimension.
*/

// testCase defines the structure for our table-driven style tests
type testCase struct {
	name           string
	input          string
	setupStyle     func() lipgloss.Style
	expectedWidth  int
	expectedHeight int
	output         []string
}

func runStyleTest(t *testing.T, tc testCase) {
	t.Run(tc.name, func(t *testing.T) {
		style := tc.setupStyle()

		output := style.Render(tc.input)

		// Strip ANSI colors/styles so we can perform reliable structural string assertions
		output = lipglossutil.StripStyles(output)

		boxSymbol := "•"
		boxExpected := lipglossutil.BoxOutputSlices(tc.output, boxSymbol)
		boxActual := lipglossutil.BoxOutputString(output, boxSymbol)
		t.Logf("\nExpected test output:\n%s\n\nActual test output:\n%s", boxExpected, boxActual)

		lines := strings.Split(output, "\n")

		// Set expected width/height values from the expected output
		expectedOutputWidth := lipglossutil.GetExpectedOutputBlockWidth(tc.output)

		// Validate expected width/height values from the expected output
		require.Equal(t, tc.expectedHeight, len(tc.output), fmt.Sprintf("expected height %d does not match the expected output height %d", tc.expectedHeight, len(tc.output)))
		require.Equal(t, tc.expectedWidth, expectedOutputWidth, fmt.Sprintf("expected width %d does not match the expected output width %d", tc.expectedWidth, expectedOutputWidth))

		// Assert height
		require.Equal(t, tc.expectedHeight, len(lines), fmt.Sprintf("expected height %d, got %d", tc.expectedHeight, len(lines)))

		// Assert width across all lines
		for i, line := range lines {
			actualWidth := lipgloss.Width(line)
			require.Equal(t, tc.expectedWidth, actualWidth, fmt.Sprintf("expected output line %d: expected width %d, got %d", i, tc.expectedWidth, actualWidth))
		}

		// Assert actual output text
		for i := range tc.output {
			expectedOutput := tc.output[i]
			actualOutput := lines[i]
			require.Equal(t, expectedOutput, actualOutput, fmt.Sprintf("expected output line %d is `%s`, got `%s`", i, expectedOutput, actualOutput))
		}
	})
}

func TestIndividualStylePropertiesWithBorders(t *testing.T) {
	tests := []testCase{
		{
			name:  "Rule 2 & 4: Simple word with space padding and Rounded Border structural check",
			input: "Go",
			setupStyle: func() lipgloss.Style {
				return lipgloss.NewStyle().Width(6).Height(3).Border(lipgloss.RoundedBorder())
			},
			expectedWidth:  8,
			expectedHeight: 5,
			output: []string{
				"╭──────╮",
				"│Go    │", // Padded with spaces to match required text width
				"│      │", // Padded with blank lines to match required text height
				"│      │",
				"╰──────╯",
			},
		},
		{
			name:  "Rule 3: Padding shrinks inner word-wrap bounds",
			input: "Wrapping",
			setupStyle: func() lipgloss.Style {
				return lipgloss.NewStyle().Width(10).Height(4).Padding(0, 2)
			},
			expectedWidth:  10,
			expectedHeight: 4,
			output: []string{
				"  Wrappi  ",
				"  ng      ",
				"          ",
				"          ",
			},
		},
		{
			name:  "Rule 7: Margin pushes borders inward and fills edges with spaces",
			input: "Box",
			setupStyle: func() lipgloss.Style {
				return lipgloss.NewStyle().Width(5).Height(1).Border(lipgloss.RoundedBorder()).Margin(1, 1, 1, 1)
			},
			expectedWidth:  9,
			expectedHeight: 5,
			output: []string{
				"         ",
				" ╭─────╮ ",
				" │Box  │ ",
				" ╰─────╯ ",
				"         ",
			},
		},
	}

	for _, tc := range tests {
		runStyleTest(t, tc)
	}
}

func TestCombinedStyleCompositionWithBorders(t *testing.T) {
	tests := []testCase{
		{
			name:  "Full Property Combination - Structural Validation",
			input: longInputText,
			setupStyle: func() lipgloss.Style {
				return lipgloss.NewStyle().
					Width(80).
					Height(12).
					Padding(1, 3, 2, 4).
					Border(lipgloss.RoundedBorder()).
					Margin(4, 2, 3, 1)
			},
			expectedWidth:  85, // 80 + 2 (borders) + 2 (margin right) + 1 (margin left)
			expectedHeight: 21,
			output: []string{
				"                                                                                     ",
				"                                                                                     ",
				"                                                                                     ",
				"                                                                                     ",
				" ╭────────────────────────────────────────────────────────────────────────────────╮  ",
				" │                                                                                │  ",
				" │    The Little Prince (Le Petit Prince) is a novella by French aristocrat,      │  ",
				" │    writer, and military aviator Antoine de Saint-Exupéry. First published in   │  ",
				" │    English and French in the United States in April 1943, it is one of the     │  ",
				" │    most translated and best-selling books in history. The story follows a      │  ",
				" │    young prince who visits various planets in space, including Earth,          │  ",
				" │    addressing themes of loneliness, friendship, love, and loss. Despite its    │  ",
				" │    style as a children's book, The Little Prince makes observations about      │  ",
				" │    life and human nature that resonate deeply with adult readers.              │  ",
				" │                                                                                │  ",
				" │                                                                                │  ",
				" │                                                                                │  ",
				" ╰────────────────────────────────────────────────────────────────────────────────╯  ",
				"                                                                                     ",
				"                                                                                     ",
				"                                                                                     ",
			},
		},
	}

	for _, tc := range tests {
		runStyleTest(t, tc)
	}
}
