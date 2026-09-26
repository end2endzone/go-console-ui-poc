package textpanelwithscrollbar

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
	"github.com/end2endzone/go-console-ui-poc/internal/tui"
	"github.com/end2endzone/go-console-ui-poc/internal/tui/components/titledborderedpanels"
	"github.com/end2endzone/go-console-ui-poc/internal/tui/lipglossutil"
)

const borderWidth = 1

// Force Model to always implements interface titledborderedpanels.TitledBorderedPanel
var _ titledborderedpanels.TitledBorderedPanel = (*Model)(nil)

type Model struct {
	Title      string
	TitleStyle lipgloss.Style
	PanelStyle lipgloss.Style
	ScrollBar  lipglossutil.ScrollBarModel
	viewport   viewport.Model
	width      int
	height     int
	focused    bool
	content    string // raw content
}

func New() Model {
	// Initialize with zero size safely
	return Model{
		viewport:  viewport.New(0, 0),
		ScrollBar: lipglossutil.NewScrollBarModel(),
	}
}

/////////////////////////////////////
// Focus interface
/////////////////////////////////////

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

/////////////////////////////////////
// Viewport interface
/////////////////////////////////////

// GotoTop sets the viewport to the top position.
func (m *Model) GotoTop() (lines []string) {
	return m.viewport.GotoTop()
}

// GotoBottom sets the viewport to the bottom position.
func (m *Model) GotoBottom() (lines []string) {
	return m.viewport.GotoBottom()
}

// ScrollLeft moves the viewport to the left by the given number of columns.
func (m *Model) ScrollLeft(n int) {
	m.viewport.ScrollLeft(n)
}

// ScrollRight moves viewport to the right by the given number of columns.
func (m *Model) ScrollRight(n int) {
	m.viewport.ScrollRight(n)
}

// ScrollUp moves the view down by the given number of lines. Returns the new
// lines to show.
func (m *Model) ScrollUp(n int) (lines []string) {
	return m.viewport.ScrollUp(n)
}

// ScrollDown moves the view down by the given number of lines.
func (m *Model) ScrollDown(n int) (lines []string) {
	return m.viewport.ScrollDown(n)
}

// PageUp moves the view up by one height of the viewport.
func (m *Model) PageUp() []string {
	return m.viewport.PageUp()
}

// PageDown moves the view down by the number of lines in the viewport.
func (m *Model) PageDown() []string {
	return m.viewport.PageDown()
}

/////////////////////////////////////
// Other
/////////////////////////////////////

func (m *Model) SetTitle(title string) {
	m.Title = title
}

func (m *Model) SetTitleStyle(style lipgloss.Style) {
	m.TitleStyle = style
}

func (m *Model) SetPanelStyle(style lipgloss.Style) {
	m.PanelStyle = style

	// Update the internal viewport
	m.UpdateViewport()
}

func (m *Model) GetMinimumSize() (width int, height int) {
	width = 2*borderWidth + m.ScrollBar.Width()
	height = 2 * borderWidth
	return
}

func (m *Model) GetMaximumContentSize() (width int, height int) {
	if m.width == 0 || m.height == 0 {
		return 0, 0
	}

	// The viewport's matches the panel if the panel's style is border is unset.
	width = m.width
	height = m.height

	// Account for any internal borders
	if m.PanelStyle.GetBorderLeft() {
		width--
	}
	if m.PanelStyle.GetBorderRight() {
		width--
	}
	if m.PanelStyle.GetBorderTop() {
		height--
	}
	if m.PanelStyle.GetBorderBottom() {
		height--
	}

	// Account for any internal padding
	width -= m.PanelStyle.GetPaddingLeft()
	width -= m.PanelStyle.GetPaddingRight()
	height -= m.PanelStyle.GetPaddingTop()
	height -= m.PanelStyle.GetPaddingBottom()

	// Account the for scrollbar
	width -= m.ScrollBar.Width()

	return
}

// SetSize is called by the parent to allocate space
func (m *Model) SetSize(width int, height int) {
	m.width = width
	m.height = height

	minWidth, minHeight := m.GetMinimumSize()
	if m.width < minWidth {
		m.width = minWidth
	}
	if m.height < minHeight {
		m.height = minHeight
	}

	// Update the internal viewport
	m.UpdateViewport()
}

// SetContent set the internal content of the panel
func (m *Model) SetContent(content string) {
	m.content = content

	// Update the internal viewport
	m.UpdateViewport()
}

// UpdateViewport is called by the parent to resize the viewport and update its content
func (m *Model) UpdateViewport() {
	m.viewport.Width, m.viewport.Height = m.GetMaximumContentSize()

	// Wrap the text to the viewport's specific width so it
	// knows exactly how many vertical lines are being rendered.
	wrappedText := lipgloss.NewStyle().Width(m.viewport.Width).Render(m.content)

	m.viewport.SetContent(wrappedText)
}

/////////////////////////////////////
// tea.Model interface
/////////////////////////////////////

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) Update(msg tea.Msg) (titledborderedpanels.TitledBorderedPanel, tea.Cmd) {
	var cmd tea.Cmd
	// Forward keys/mouse messages to the viewport
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m *Model) View() string {
	viewportView := m.viewport.View()

	// Force PanelStyle to be bordered
	if !lipglossutil.IsStyleBordered(&m.PanelStyle) {
		m.PanelStyle = m.PanelStyle.Border(lipgloss.RoundedBorder())
	}

	// Insert the scrollbar on the right side
	var scrollbarView string
	{
		// Split by line to be able to manipulate lines individually
		lines := strings.Split(viewportView, "\n")
		vpHeight := len(lines) // number of line displayed
		if vpHeight == 0 {
			// If viewport content is empty, return immediately
			return viewportView
		}

		// Determine scroll thumb position
		scrollPercent := m.viewport.ScrollPercent()
		thumbIndex := int(scrollPercent * float64(vpHeight-1))

		//totalLinesCount := m.viewport.TotalLineCount()
		//visibleLinesCount := m.viewport.VisibleLineCount()

		// Fix thumbPos if we are at the top most or botto mmost viewport
		if m.viewport.AtBottom() {
			thumbIndex = vpHeight - 1
		}
		if m.viewport.AtTop() {
			thumbIndex = 0
		}

		scrollbarView = m.ScrollBar.RenderLinesContent(lines, thumbIndex)
	}

	view := tui.RenderBorderWithTitle(m.PanelStyle, m.Title, m.TitleStyle, scrollbarView)
	return view
}
