package titledborderedpanels

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// TitledBorderedPanel interface defining required behaviors
type TitledBorderedPanel interface {
	// Focus methods
	Focused() bool
	Focus()
	Blur()

	// Panel methods
	SetTitle(title string)
	SetTitleStyle(style lipgloss.Style)
	SetPanelStyle(style lipgloss.Style)
	SetSize(width int, height int)
	GetMinimumSize() (width int, height int)
	GetMaximumContentSize() (width int, height int)
	SetContent(content string)

	// Bubble tea.Model interface
	Init() tea.Cmd
	Update(msg tea.Msg) (TitledBorderedPanel, tea.Cmd)
	View() string
}
