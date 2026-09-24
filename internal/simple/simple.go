package main

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type mainModel struct {
	Width  int
	Height int
	Text   string
}

func (m mainModel) Init() tea.Cmd {
	return nil
}

func (m mainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// Parent calculates exactly how much room the embedded component gets
		m.Width = msg.Width / 3
		m.Height = msg.Height - 4 // Leave room for headers/footers
		cmd = nil
		cmds = append(cmds, cmd)
	case tea.KeyMsg:
		switch msg.String() {
		case "q":
			return m, tea.Quit
		case "esc", "ctrl+c":
			return m, tea.Quit
		}
	}

	return m, tea.Batch(cmds...)
}

func GetPropertyIfNonZero(indent int, name string, value int) string {
	if value == 0 {
		return ""
	}

	output := strings.Repeat(" ", indent) + fmt.Sprintf("%s=%d\n", name, value)
	return output
}

func bool2int(value bool) int {
	if value {
		return 1
	}
	return 0
}

func JointInts(values []int, sep string) string {
	output := ""
	for i := range values {
		isLast := (i + 1) == len(values)
		output += fmt.Sprintf("%d", values[i])
		if !isLast {
			output += sep
		}
	}
	return output
}

func IsAllIntZeros(values []int) bool {
	for i := range values {
		if values[i] != 0 {
			return false
		}
	}
	return true
}

func GetQuadIntPropertyIfNonZero(indent int, name string, values []int) string {
	if IsAllIntZeros(values) {
		return ""
	}

	output := strings.Repeat(" ", indent) + fmt.Sprintf("%s: %s\n", name, JointInts(values, ","))
	return output
}

func JointBools(values []bool, sep string) string {
	output := ""
	for i := range values {
		isLast := (i + 1) == len(values)
		output += fmt.Sprintf("%d", bool2int(values[i]))
		if !isLast {
			output += sep
		}
	}
	return output
}

func IsAllBoolFalse(values []bool) bool {
	for i := range values {
		if values[i] != false {
			return false
		}
	}
	return true
}

func GetQuadBoolPropertyIfNonZero(indent int, name string, values []bool) string {
	if IsAllBoolFalse(values) {
		return ""
	}

	output := strings.Repeat(" ", indent) + fmt.Sprintf("%s: %s\n", name, JointBools(values, ","))
	return output
}

func StyleDetails(s lipgloss.Style) string {
	output := "{\n"
	//output += GetPropertyIfNonZero(2, "width ", s.GetWidth()) + GetPropertyIfNonZero(0, "height", s.GetHeight())
	output += GetQuadIntPropertyIfNonZero(2, "width/height", []int{
		s.GetWidth(),
		s.GetHeight(),
	},
	)
	output += GetQuadIntPropertyIfNonZero(2, "padding (top, right, bottom, left)", []int{
		s.GetPaddingTop(),
		s.GetPaddingRight(),
		s.GetPaddingBottom(),
		s.GetPaddingLeft(),
	},
	)
	output += GetQuadIntPropertyIfNonZero(2, "margin (top, right, bottom, left)", []int{
		s.GetMarginTop(),
		s.GetMarginRight(),
		s.GetMarginBottom(),
		s.GetMarginLeft(),
	},
	)
	output += GetQuadBoolPropertyIfNonZero(2, "border (top, right, bottom, left)", []bool{
		s.GetBorderTop(),
		s.GetBorderRight(),
		s.GetBorderBottom(),
		s.GetBorderLeft(),
	},
	)
	output += "}"

	return output
}

func GetStyleExpectedOutputSize(s lipgloss.Style) (int, int) {
	width := s.GetWidth()
	height := s.GetHeight()
	if width == 0 || height == 0 {
		return -1, -1 // unknown
	}

	// The output will be at least width x height large

	// If there is a border, the border wraps the style's width and height area
	if s.GetBorderLeft() {
		width += 1
	}
	if s.GetBorderRight() {
		width += 1
	}

	if s.GetBorderTop() {
		height += 1
	}
	if s.GetBorderBottom() {
		height += 1
	}

	// Margin is added over the border
	marginLeft := s.GetMarginLeft()
	marginRight := s.GetMarginRight()
	marginTop := s.GetMarginTop()
	marginBottom := s.GetMarginBottom()
	if marginLeft > 0 {
		width += marginLeft
	}
	if marginRight > 0 {
		width += marginRight
	}
	if marginTop > 0 {
		height += marginTop
	}
	if marginBottom > 0 {
		height += marginBottom
	}

	return width, height
}

func (m mainModel) View() string {

	style := lipgloss.NewStyle().
		Width(80).
		Height(12).
		Padding(1, 3, 2, 4).
		Border(lipgloss.RoundedBorder()).
		Margin(4, 2, 3, 1)

	details := StyleDetails(style)
	rendered := style.Render(m.Text)

	renderedWidth := lipgloss.Width(rendered)
	renderedHeight := lipgloss.Height(rendered)

	output := ""
	output += fmt.Sprintf("%s\n", details)
	output += GetPropertyIfNonZero(0, "renderedWidth ", renderedWidth)
	output += GetPropertyIfNonZero(0, "renderedHeight", renderedHeight)
	output += rendered
	return output
}

func main() {
	model := mainModel{
		Text: "The Little Prince (Le Petit Prince) is a novella by French aristocrat, writer, and military aviator Antoine de Saint-Exupéry. First published in English and French in the United States in April 1943, it is one of the most translated and best-selling books in history. The story follows a young prince who visits various planets in space, including Earth, addressing themes of loneliness, friendship, love, and loss. Despite its style as a children's book, The Little Prince makes observations about life and human nature that resonate deeply with adult readers.", // \n\nThe narrative begins with an aviator stranded in the Sahara Desert after his plane crashes. While attempting to repair his engine, he meets a extraordinary boy dubbed 'the Little Prince.' The prince shares stories of his small home asteroid, B-612, where he spent his days raking out volcanoes and caring for a beautiful but proud rose. Feeling neglected by the rose, he set off on a journey across the cosmos. Along the way, he visits six other asteroids, each inhabited by an adult who embodies a flawed aspect of society: a king with no subjects, a vain man seeking admiration, a drunkard drinking to forget shame, a businessman counting stars he claims to own, a lamplighter bound by obsolete orders, and a geographer who knows nothing of his own world. Upon arriving on Earth, the prince learns vital life lessons about what truly matters from a wild fox, who teaches him that 'one sees clearly only with the heart; what is essential is invisible to the eye.'",
		//Text: "1234567890abcdefghijklmnopqrstuvwxyz1234567890abcdefghijklmnopqrstuvwxyz1234567890abcdefghijklmnopqrstuvwxyz1234567890abcdefghijklmnopqrstuvwxyz1234567890abcdefghijklmnopqrstuvwxyz1234567890abcdefghijklmnopqrstuvwxyz1234567890abcdefghijklmnopqrstuvwxyz1234567890abcdefghijklmnopqrstuvwxyz1234567890abcdefghijklmnopqrstuvwxyz1234567890abcdefghijklmnopqrstuvwxyz1234567890abcdefghijklmnopqrstuvwxyz1234567890abcdefghijklmnopqrstuvwxyz1234567890abcdefghijklmnopqrstuvwxyz1234567890abcdefghijklmnopqrstuvwxyz1234567890abcdefghijklmnopqrstuvwxyz1234567890abcdefghijklmnopqrstuvwxyz1234567890abcdefghijklmnopqrstuvwxyz",
	}

	p := tea.NewProgram(model, tea.WithAltScreen()) // WithAltScreen is recommended for viewports
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v", err)
		os.Exit(1)
	}
}
