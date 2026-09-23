package viewportwithverticalscrollbar

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Theme display constants
const (
	scrollBarWidth = 2 // For example " █", note the space before the scroll bar cursor
)

type Model struct {
	CursorStyle    lipgloss.Style
	ScrollBarStyle lipgloss.Style
	Style          lipgloss.Style
	Width          int
	Height         int
	viewport       viewport.Model
	focused        bool
	//content        string
}

// Focused returns the focus state of the component.
func (m *Model) Focused() bool {
	return m.focused
}

// Focus focuses the component, allowing the user to move around the rows and interact.
func (m *Model) Focus() {
	m.focused = true
}

// Blur blurs the component, preventing selection or movement.
func (m *Model) Blur() {
	m.focused = false
}

func (m *Model) SetContent(text string) {
	// Wraps text to fit inside the viewport content area.
	// The function shinks text by scrollBarWidth characters to reserve space for the scrollbar string at the end of each line.
	tmpWidth := m.Width - scrollBarWidth
	wrappedContent := lipgloss.NewStyle().Width(tmpWidth).Render(text)

	m.viewport.SetContent(wrappedContent)
}

// Renders the current model's viewport content side-by-side with a vertical dynamic scrollbar.
func (m *Model) renderViewportWithScrollbar() string {
	// Render viewport content normally.

	// Even if we already called SetContent() to trim the content to 2 characters less than the width of the viewport,
	// the code from go\pkg\mod\github.com\charmbracelet\lipgloss@v1.1.0\style.go adds padding at the end
	// of each the strings to match the width of the viewport. See function lipgloss.Style.Render() with the line
	// `str = alignTextHorizontal(str, horizontalAlign, width, st)`.
	// To work around this, we temporary shrink the width of the viewport to prevent padding.
	m.viewport.Width = m.Width - scrollBarWidth
	viewportView := m.viewport.View()
	m.viewport.Width = m.Width

	// Split by line to be able to manipulate lines individually
	lines := strings.Split(viewportView, "\n")
	vpHeight := len(lines) // number of line displayed
	if vpHeight == 0 {
		// If viewport content is empty, return immediately
		return viewportView
	}

	// Styles for track and scroll handle
	trackStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("238"))
	thumbStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("63")).Bold(true)

	// Determine scroll thumb position
	scrollPercent := m.viewport.ScrollPercent()
	thumbPos := int(scrollPercent * float64(vpHeight-1))

	// Fix thumbPos if we are at the top most or botto mmost viewport
	if m.viewport.AtBottom() {
		thumbPos = vpHeight - 1
	}
	if m.viewport.AtTop() {
		thumbPos = 0
	}

	// Append scrollbar characters at the end of each displayed line
	var output strings.Builder
	for i, line := range lines {
		var scrollChar string
		if i == thumbPos {
			scrollChar = thumbStyle.Render("█")
		} else {
			scrollChar = trackStyle.Render("│")
		}

		// Print the line itself and then the scrollbar
		output.WriteString(fmt.Sprintf("%s %s", line, scrollChar))

		isLast := (i + 1) == len(lines)
		if !isLast {
			output.WriteString("\n")
		}
	}

	return output.String()
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	if m.focused {
		m.viewport, cmd = m.viewport.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m *Model) View() string {
	m.viewport.Width = m.Width
	m.viewport.Height = m.Height
	m.viewport.Style = m.Style

	renderedContent := m.renderViewportWithScrollbar()
	return renderedContent
}
