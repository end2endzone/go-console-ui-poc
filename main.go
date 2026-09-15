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

// Theme display constants
const (
	leftRightBorderWidth  = 1
	leftRightPaddingWidth = 1
	scrollBarWidth        = 2                                     // For example " █", note the space before the scroll bar cursor
	minTableItems         = 2                                     // Minimum number of data rows (excluding table's header rows)
	minTableHeight        = minTableItems + 2                     // A full table height includes 2 line table header
	minContentHeight      = 6 + minTableItems                     // For right panel, that is: 2 lines for "Books" header + 2 lines for the table, minTableItems, 2 lines cursor indicator footer
	minViewportTextWidth  = 1                                     // Minimum width of the text (exclusing the scroll bars characters)
	minRightViewportWidth = minViewportTextWidth + scrollBarWidth // 1 character wide + scroll bar
	minRightWidth         = minRightViewportWidth +
		2*leftRightBorderWidth +
		2*leftRightPaddingWidth // 2 characters for border, 2 characters for padding
)

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

func getTableColumnsWidth(t *table.Model) int {
	width := 0
	columns := t.Columns()
	for _, c := range columns {
		width += 1 + c.Width + 1 // each column is padded with 1 space. In other words, a 25 characters wide columns is 27 characters long.
	}
	return width
}

func getRenderedTextMaximumWidth(text string) int {
	maxLen := -1

	lines := strings.Split(text, "\n")
	for _, line := range lines {
		if len(line) > maxLen {
			maxLen = len(line)
		}
	}

	return maxLen
}

func setRenderedTextLineValue(text *string, linenumber int, value string) {
	lines := strings.Split(*text, "\n")

	// Assert linenumber
	if linenumber >= len(lines) {
		err := fmt.Errorf("failed to set line %d to value %s in a text string that is only %d lines", linenumber, value, len(lines))
		panic(err)
	}

	lines[linenumber] = value

	*text = strings.Join(lines, "\n")
}

func ShrinkTableLastColumn(t *table.Model) {
	// First remove the last column in rows.
	// Without this, there is an index out of range runtime error.
	rows := t.Rows()
	for i := range rows {
		// remove 1 column in the row
		rows[i] = rows[i][:len(rows[i])-1]
	}
	t.SetRows(rows)

	// Then remove the actual last column
	columns := t.Columns()
	columns = columns[:len(columns)-1]
	t.SetColumns(columns)
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

// onSelectedBookChanged refreshes the current model's UI elements when the user changes the current selected book in the table in the left panel.
func (m *model) onSelectedBookChanged() {
	cursor := m.table.Cursor()
	if cursor < len(m.books) {
		description := m.books[cursor].Description
		wrappedContent := m.formatViewportContent(description)
		m.viewport.SetContent(wrappedContent)
	}
}

// Wraps text to fit inside the viewport content area.
// The function shinks text by scrollBarWidth characters to reserve space for the scrollbar string at the end of each line.
func (m *model) formatViewportContent(text string) string {
	textWithoutScrollBarWidth := m.viewport.Width - scrollBarWidth

	// Check for minimum length size
	if textWithoutScrollBarWidth < minViewportTextWidth {
		textWithoutScrollBarWidth = minViewportTextWidth
	}

	wrappedContent := lipgloss.NewStyle().Width(textWithoutScrollBarWidth).Render(text)

	//DEBUG
	//maxTextLength := getRenderedTextMaximumWidth(wrappedContent)
	//setRenderedTextLineValue(&wrappedContent, 4, fmt.Sprintf("MTL-1=%d", maxTextLength))

	return wrappedContent
}

// Renders the current model's viewport content side-by-side with a vertical dynamic scrollbar.
func (m *model) renderViewportWithScrollbar() string {
	// Render viewport content normally.

	// Even if we already called SetContent() to trim the content to 2 characters less than the width of the viewport,
	// the code from go\pkg\mod\github.com\charmbracelet\lipgloss@v1.1.0\style.go adds padding at the end
	// of each the strings to match the width of the viewport. See function lipgloss.Style.Render() with the line
	// `str = alignTextHorizontal(str, horizontalAlign, width, st)`.
	// To work around this, we temporary shrink the width of the viewport to prevent padding.
	m.viewport.Width -= scrollBarWidth
	viewportView := m.viewport.View()
	m.viewport.Width += scrollBarWidth

	//DEBUG
	//maxTextLength := getRenderedTextMaximumWidth(viewportView)
	//setRenderedTextLineValue(&viewportView, 5, fmt.Sprintf("MTL-2=%d", maxTextLength))
	//setRenderedTextLineValue(&viewportView, 6, fmt.Sprintf("vp.w=%d", m.viewport.Width))

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

		var leftWidth int
		var rightWidth int
		var contentHeight int
		var rightViewportWidth int
		{
			// Width computation of both panels
			{
				tableWidth := getTableColumnsWidth(&m.table)
				leftWidth = tableWidth + 4 // +2 for padding (1 on each side), +2 borders
				rightWidth = m.width - leftWidth

				if rightWidth < minRightWidth {
					rightWidth = minRightWidth
				}

				//DEBUG
				//rightWidth = 15
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

				//// Debuging code to auto-shrink the table to match a targetted leftWidth ?
				//// This makes the table width responsive based on available screen width.
				//if len(m.table.Columns()) == 3 {
				//	ShrinkTableLastColumn(&m.table)
				//	ShrinkTableLastColumn(&m.table)
				//}

				// Leave table's width to default value 0 so that is uses the minimum required width
				m.table.SetHeight(tableHeight)
			}

			// Right panel calculations
			{
				rightViewportWidth = rightWidth - 2*leftRightBorderWidth - 2*leftRightPaddingWidth
			}
		}

		if !m.ready {
			m.viewport = viewport.New(rightViewportWidth, contentHeight-2) //right side has a 2 lines non-scrollable header
			m.ready = true
		} else {
			m.viewport.Width = rightViewportWidth
			m.viewport.Height = contentHeight - 2 //right side has a 2 lines non-scrollable header
		}

		//DEBUG
		//_, right, _, left := m.viewport.Style.GetPadding()
		//m.table.Rows()[0] = table.Row{
		//	fmt.Sprintf("rightViewportWidth=%d", rightViewportWidth),
		//	fmt.Sprintf("rightWidth=%d", rightWidth),
		//	"",
		//}
		//m.table.Rows()[2] = table.Row{
		//	fmt.Sprintf("minRightViewportWidth=%d", minRightViewportWidth),
		//	fmt.Sprintf("minRightWidth=%d", minRightWidth),
		//	"",
		//}
		//m.table.Rows()[1] = table.Row{
		//	fmt.Sprintf("padding: %d,%d", right, left),
		//	"",
		//	"",
		//}

		// The right viewport dimensions have changed.
		// Force updating the right viewport with new automatically wrapped content.
		m.onSelectedBookChanged()

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

				// The selected book have changed.
				// Force updating the right viewport with new automatically wrapped content.
				m.onSelectedBookChanged()

				// And move the viewport to the top of the view
				m.viewport.GotoTop()

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
	rightTitle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205")).Render("Summary")

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
	_, err := p.Run()
	if err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}
