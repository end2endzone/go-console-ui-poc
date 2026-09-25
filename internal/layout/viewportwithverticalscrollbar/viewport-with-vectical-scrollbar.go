package viewportwithverticalscrollbar

import (
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/end2endzone/go-console-ui-poc/internal/lipglossutil"
)

// Theme display constants
const (
	scrollBarWidth = 2 // For example " █", note the space before the scroll bar cursor
)

type Model struct {
	ThumbStyle lipgloss.Style
	TrackStyle lipgloss.Style
	Style      lipgloss.Style
	width      int
	height     int
	viewport   viewport.Model
	focused    bool
	rawText    string
}

// NewModel creates a new model for the viewport with vectical scrollbar widget.
func NewModel() Model {
	m := Model{
		ThumbStyle: lipgloss.NewStyle().Foreground(lipgloss.ANSIColor(8)).Bold(true),
		TrackStyle: lipgloss.NewStyle().Foreground(lipgloss.ANSIColor(8)),
		Style:      lipgloss.NewStyle(),
		width:      0,
		height:     0,
		viewport:   viewport.New(0, 0),
		focused:    false,
	}

	return m
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

// GotoTop sets the viewport to the top position.
func (m *Model) GotoTop() (lines []string) {
	return m.viewport.GotoTop()
}

// GotoBottom sets the viewport to the bottom position.
func (m *Model) GotoBottom() (lines []string) {
	return m.viewport.GotoBottom()
}

// SetSize is called by the parent to allocate space
func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height

	// Account for any internal padding, borders or scrollbar the child has
	m.viewport.Width = width - scrollBarWidth
	m.viewport.Height = height

	m.updateViewport()
}

func (m *Model) SetContent(text string) {
	m.rawText = text

	m.updateViewport()
}

func (m *Model) updateViewport() {
	// If size is already set, update the viewport content immediately
	if m.viewport.Width > 0 {
		// Wraps text to fit inside the viewport content area.
		// Use the viewport's width which is already shrunk by scrollBarWidth characters to reserve space for the scrollbar string at the end of each line.
		wrappedContent := lipgloss.NewStyle().Width(m.viewport.Width).Render(m.rawText)

		// DEBUG
		/*longest, actualLine := debugging.GetLongestLineInText(wrappedContent) // DEBUG
		if longest > 6543 || actualLine == "123456789" {
			return
		}
		debugging.DumpRenderingWithoutStylesToFile("m.viewport.SetContent().txt", wrappedContent)*/

		m.viewport.SetContent(wrappedContent)
	}
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

// Renders the current model's viewport content side-by-side with a vertical dynamic scrollbar.
func (m *Model) renderViewportWithScrollbar() string {
	if m.viewport.Width == 0 || m.viewport.Height == 0 {
		return ""
	}

	view := lipglossutil.ViewportViewWithVerticalScrollBar(&m.viewport, m.rawText)
	return view
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	if m.focused {
		m.viewport, cmd = m.viewport.Update(msg)
		cmds = append(cmds, cmd)
	}

	return *m, tea.Batch(cmds...)
}

func (m *Model) View() string {
	if m.width == 0 || m.height == 0 {
		return ""
	}

	//m.viewport.Style = m.Style

	renderedContent := m.renderViewportWithScrollbar()

	// DEBUG
	/*longest, actualLine := debugging.GetLongestLineInText(renderedContent) // DEBUG
	if longest > 6543 || actualLine == "123456789" {
		return ""
	}
	longest, actualLine = debugging.GetLongestLineInText(debugging.StripStyles(renderedContent)) // DEBUG
	if longest > 6543 || actualLine == "123456789" {
		return ""
	}
	debugging.DumpRenderingWithoutStylesToFile("viewport-with-vectical-scrollbar.View().txt", renderedContent)*/

	return renderedContent
}
