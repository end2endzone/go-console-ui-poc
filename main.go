package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Book struct {
	Title       string `json:"title"`
	Author      string `json:"author"`
	Year        int    `json:"year"`
	Description string `json:"description"`
}

// Force model to always implements interface tea.Model
var _ tea.Model = (*model)(nil)

type model struct {
	books       []Book
	table       table.Model
	viewport    viewport.Model
	activePanel int // 0 = Left (Table), 1 = Right (Viewport)
	ready       bool
	width       int
	height      int
}

// ReadBooksFromFile reads a JSON file and parses it into a slice of Books.
func ReadBooksFromFile(filePath string) ([]Book, error) {
	// Read the raw bytes from the file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Unmarshal the JSON array into a slice of Book structs
	var books []Book
	err = json.Unmarshal(data, &books)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	return books, nil
}

func getTableColumnsWidth(t table.Model) int {
	width := 0
	columns := t.Columns()
	for _, c := range columns {
		width += 1 + c.Width + 1 // each column is padded with 1 space. In other words, a 25 characters wide columns is 27 characters long.
	}
	return width
}

func initialModel() model {
	books, err := ReadBooksFromFile("books.json")
	if err != nil {
		panic(err)
	}

	columns := []table.Column{
		{Title: "Title", Width: 25},
		{Title: "Author", Width: 20},
		{Title: "Year", Width: 4},
	}

	rows := make([]table.Row, len(books))
	for i, b := range books {
		rows[i] = table.Row{b.Title, b.Author, strconv.Itoa(b.Year)}
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(true)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(true)
	t.SetStyles(s)

	return model{
		books:       books,
		table:       t,
		activePanel: 0,
	}
}

// Wraps text to fit inside the viewport content area (accounting for scrollbar width)
func (m model) formatViewportContent(text string, scrollBarWidth int) string {
	viewportWidth := m.viewport.Width - scrollBarWidth
	if viewportWidth <= 0 {
		viewportWidth = 20
	}
	return lipgloss.NewStyle().Width(viewportWidth).Render(text)
}

// Renders the viewport content side-by-side with a vertical dynamic scrollbar
func (m model) renderViewportWithScrollbar() string {
	viewportView := m.viewport.View()
	lines := strings.Split(viewportView, "\n")
	vpHeight := len(lines) // number of line displayed

	if vpHeight == 0 {
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

	// Attach scrollbar character at the end of each displayed line
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

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		const (
			minTableItems    = 2
			minTableHeight   = minTableItems + 2 // a full table height includes 2 line table header
			minRightWidth    = 20
			minContentHeight = 6 + minTableItems // for right panel, that is: 2 lines for "Books" header + 2 lines for the table, minTableItems, 2 lines cursor indicator footer
			scrollBarWidth   = 2                 // For example " █",  note the space before the scroll bar string
		)

		var leftWidth int
		var rightWidth int
		var contentHeight int
		{
			// Width computation of both panels
			{
				tableWidth := getTableColumnsWidth(m.table)
				leftWidth = tableWidth + 4           // +2 for padding (1 on each side), +2 borders
				rightWidth = m.width - leftWidth - 4 // don't really know why I need to remove these 4 mandatory spaces but if I don't each line are clipped.

				if rightWidth < minRightWidth {
					rightWidth = minRightWidth
				}
			}

			// Height computation of both panels
			{
				contentHeight = m.height - 4 // 2 lines for the top and bottom borders, 2 lines for the help string (the help string itself and a final \n)
				if contentHeight < minContentHeight {
					contentHeight = minContentHeight
				}
			}

			// Left panel calculations
			{
				tableHeight := contentHeight - 4 // 2 lines for "Books" header + 2 lines cursor indicator footer
				if tableHeight < minTableHeight {
					tableHeight = minTableHeight
				}

				m.table.SetHeight(tableHeight)
			}
		}

		if !m.ready {
			m.viewport = viewport.New(rightWidth-scrollBarWidth, contentHeight-2) //right side has a 2 lines non-scrollable header
			m.ready = true
		} else {
			m.viewport.Width = rightWidth - scrollBarWidth
			m.viewport.Height = contentHeight - 2 //right side has a 2 lines non-scrollable header
		}

		cursor := m.table.Cursor()
		if cursor < len(m.books) {
			m.viewport.SetContent(m.formatViewportContent(m.books[cursor].Description, scrollBarWidth))
		}

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit

		case "tab", "right", "left":
			if m.activePanel == 0 {
				m.activePanel = 1
				m.table.Blur()
			} else {
				m.activePanel = 0
				m.table.Focus()
			}

		case "up", "down", "j", "k":
			if m.activePanel == 0 {
				m.table, cmd = m.table.Update(msg)
				cmds = append(cmds, cmd)

				cursor := m.table.Cursor()
				if cursor < len(m.books) {
					wrappedContent := m.formatViewportContent(m.books[cursor].Description, 2)
					m.viewport.SetContent(wrappedContent)
					m.viewport.GotoTop()
				}
				return m, tea.Batch(cmds...)
			} else {
				m.viewport, cmd = m.viewport.Update(msg)
				cmds = append(cmds, cmd)
				return m, tea.Batch(cmds...)
			}
		}
	}

	if m.activePanel == 0 {
		m.table, cmd = m.table.Update(msg)
		cmds = append(cmds, cmd)
	} else {
		m.viewport, cmd = m.viewport.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	if !m.ready {
		return "Initializing UI..."
	}

	leftBorderColor := lipgloss.Color("240")
	rightBorderColor := lipgloss.Color("240")

	if m.activePanel == 0 {
		leftBorderColor = lipgloss.Color("63")
	} else {
		rightBorderColor = lipgloss.Color("63")
	}

	leftStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(leftBorderColor).
		Padding(0, 1)

	rightStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(rightBorderColor).
		Padding(0, 1)

	leftTitle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205")).Render("Books")
	rightTitle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205")).Render("Description / Summary")

	positionIndicatorText := fmt.Sprintf("[%d/%d]", m.table.Cursor()+1, len(m.table.Rows()))

	leftPanel := leftStyle.Render(leftTitle + "\n\n" + m.table.View() + "\n\n" + positionIndicatorText)
	rightPanel := rightStyle.Render(rightTitle + "\n\n" + m.renderViewportWithScrollbar())

	body := lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, rightPanel)

	help := lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render(
		" Tab/←/→: Switch Active Panel  |  ↑/↓: Scroll  |  q: Quit",
	)

	return body + "\n" + help + "\n"
}

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}
