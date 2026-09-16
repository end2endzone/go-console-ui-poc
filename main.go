package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"

	//"charm.land/bubbles/v2/textinput"
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

type ActiveComponent int

const (
	Unknown       ActiveComponent = iota // 0
	LeftTable                            // 1
	RightViewport                        // 2
	SearchText                           // 3
)

type theme struct {
	focusedBorderColor   lipgloss.Color
	unfocusedBorderColor lipgloss.Color
	panelsPadding        []int
	headerTextStyle      lipgloss.Style
}

type model struct {
	theme           theme
	books           []Book
	filteredBooks   []*Book
	table           table.Model
	viewport        viewport.Model
	viewportFocused bool
	searchText      textinput.Model
	ready           bool
	width           int
	height          int
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

func NewTheme() theme {
	theme := theme{
		focusedBorderColor:   lipgloss.Color("63"),
		unfocusedBorderColor: lipgloss.Color("240"),
		panelsPadding:        []int{0, 1, 0, 1}, // top, right, bottom, left
		headerTextStyle:      lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205")),
	}

	return theme
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

// ShrinkTableLastColumn removes the last column of the given table.
// The function can be used to auto-shrink a table (hide non-important columns) to match a targetted width.
// This allows a tables to be "responsive" and react dynamically based on available screen width.
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

// FindBookIndex finds the given book in the given list of books
// Returns the index where the book is found
// Returns -1 if the book is not found
func FindBookIndex(query *Book, books []*Book) int {
	for i := range books {
		tmp := books[i]
		if query == tmp {
			return i
		}
	}
	return -1
}

func TruncatValueAsPerTableColumn(value string, table table.Model, columnIdx int) string {
	columns := table.Columns()

	// Assert columnIdx not out of range
	if columnIdx < 0 || columnIdx >= len(columns) {
		err := fmt.Errorf("invalid column index %d on a table with %d columns", columnIdx, len(columns))
		panic(err)
	}

	column := columns[columnIdx]

	width := column.Width
	substring := value[0:min(len(value), width)]

	// It is truncated? For example "Harry Potter and the Phi…"
	if len(substring) == width {
		// Potentially truncated.
		if len(value) > len(substring) {
			// Yes it is
			substring = value[0:width-1] + "…"
		}
	}

	return substring
}

func hasBookChanged(before *Book, after *Book) bool {
	if before == nil && after == nil {
		return false
	}

	if before == nil && after != nil {
		return true
	} else if before != nil && after == nil {
		return true
	}

	// Both books instance are non-nil
	if before.Title == after.Title {
		return false
	}

	return true
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

	t := table.New(
		table.WithColumns(columns),
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

	searchText := textinput.New()
	searchText.Placeholder = "filter"
	searchText.CharLimit = 156
	searchText.Width = 40

	m := model{
		theme:      NewTheme(),
		books:      books,
		table:      t,
		searchText: searchText,
	}

	m.FocusComponent(LeftTable)

	// Fill Table
	m.FillBooksTable("")

	return m
}

// ActiveComponent return the active focused component in the main UI.
func (m *model) ActiveComponent() ActiveComponent {
	if m.table.Focused() {
		return LeftTable
	} else if m.viewportFocused {
		return RightViewport
	} else if m.searchText.Focused() {
		return SearchText
	}

	return Unknown
}

// FocusComponent focuses the given component and blur other components.
func (m *model) FocusComponent(c ActiveComponent) {
	switch c {
	case LeftTable:
		m.table.Focus()
		m.viewportFocused = false
		m.searchText.Blur()
	case RightViewport:
		m.table.Blur()
		m.viewportFocused = true
		m.searchText.Blur()
	case SearchText:
		m.table.Blur()
		m.viewportFocused = false
		m.searchText.Focus()
	default:
		m.table.Blur()
		m.viewportFocused = false
		m.searchText.Blur()
	}
}

// SelectedBook returns the current Book selected in the left table.
// Returns nil if no book is selected.
func (m *model) SelectedBook() *Book {
	cursor := m.table.Cursor()
	if cursor >= 0 && cursor < len(m.filteredBooks) {
		book := m.filteredBooks[cursor]
		return book
	}
	return nil
}

// onSelectedBookChanged refreshes the current model's UI elements when the user changes the current selected book in the table in the left panel.
func (m *model) onSelectedBookChanged() {
	bookPtr := m.SelectedBook()
	if bookPtr != nil {
		description := bookPtr.Description
		wrappedContent := m.formatViewportContent(description)
		m.viewport.SetContent(wrappedContent)
	} else {
		// There is no book selected, clear the right panel content
		m.viewport.SetContent("")
	}
}

// onFilterChanged refreshes the current model's UI elements when the user changes the current filter.
func (m *model) onFilterChanged() {

	previousBookPtr := m.SelectedBook()

	// Rebuild the books left table
	filter := m.searchText.Value()
	m.FillBooksTable(filter)

	// Try to restore the previous selection
	if previousBookPtr != nil {
		// Search for the previous book in the new selection
		index := FindBookIndex(previousBookPtr, m.filteredBooks)
		if index != -1 {
			// The new index matching the same previous books is found
			m.table.SetCursor(index)

			// Fix a specific bug:
			truncatedTitle := TruncatValueAsPerTableColumn(previousBookPtr.Title, m.table, 0)
			content := m.table.View()
			if !strings.Contains(content, truncatedTitle) {
				// This is a bug where even if we forced the SetCursor() to properly select our value,
				// the table's viewport does not update properly to show our selected value.

				// Try to fix the issue in a dirty way
				m.table.MoveUp(1)
				m.table.MoveDown(1)

				// and test again
				content := m.table.View()
				if !strings.Contains(content, truncatedTitle) {
					err := fmt.Errorf("The selected book named '%s' is selected but not in the table's internal viewport!\n"+
						"The table's output is the following:\n%s", previousBookPtr.Title, content)

					panic(err)
				}
			}
		}

		// Force the right panel to update for one of the following reasons:
		// 1. The previous element is found so we called m.table.SetCursor(). But the m.table.SetRows() in FillBooksTable() has resetted the selection to 0 and also resetted the right panel content.
		// 2. The previous element is not found anymore. We got a new selection.
		// 3. The previous element is not found anymore. There is no result that matches the search pattern. The table is empty

		// Refresh the right panel
		m.onSelectedBookChanged()
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

// FilterBooks filters the list of books based on the given filter
func (m *model) FilterBooks(filter string) []*Book {
	// Filter using case unsensitive
	filter = strings.ToLower(filter)

	filteredBooks := []*Book{}
	for i, b := range m.books {
		if strings.Contains(strings.ToLower(b.Title), filter) ||
			strings.Contains(strings.ToLower(b.Author), filter) {
			filteredBooks = append(filteredBooks, &m.books[i])
		}
	}
	return filteredBooks
}

// FillBooksTable updates the left table's rows based on the given filter
func (m *model) FillBooksTable(filter string) {
	var filteredBooks []*Book
	var rows []table.Row

	if len(filter) <= 1 {
		// No filter specified or filter too small and ignored
		filteredBooks = make([]*Book, len(m.books))
		rows = make([]table.Row, len(m.books))
		for i := range m.books {
			bookPtr := &m.books[i]
			filteredBooks[i] = bookPtr
			rows[i] = table.Row{bookPtr.Title, bookPtr.Author, strconv.Itoa(bookPtr.Year)}
		}
	} else {
		// Filter the list based on the filter
		filteredBooks = m.FilterBooks(filter)
		rows = make([]table.Row, len(filteredBooks))
		for i, b := range filteredBooks {
			rows[i] = table.Row{b.Title, b.Author, strconv.Itoa(b.Year)}
		}
	}

	// Apply
	m.filteredBooks = filteredBooks
	m.table.SetRows(rows)

	// Note that cursor is modified by the table when we change the rows
	m.table.SetCursor(0) // force first element to be selected

	// Force the right panel to update to the selection (or "unselection").
	m.onSelectedBookChanged()
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	// Get current component
	activeComponent := m.ActiveComponent()

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
			}

			// Height computation of both panels
			{
				contentHeight = m.height - 5 // 2 lines for the top and bottom borders, 1 search line, 2 lines for the help string (the help string itself and a final \n)
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

		// The right viewport dimensions have changed.
		// Force updating the right viewport with new automatically wrapped content.
		m.onSelectedBookChanged()

	case tea.KeyMsg:
		switch msg.String() {
		case "q":
			if activeComponent != SearchText {
				// only allow q to quit when its not the search string that has focus
				return m, tea.Quit
			}
		case "esc", "ctrl+c":
			return m, tea.Quit

		case "tab":
			// Focus the next component
			switch activeComponent {
			case LeftTable:
				m.FocusComponent(RightViewport)
			case RightViewport:
				m.FocusComponent(SearchText)
			case SearchText:
				m.FocusComponent(LeftTable)
			default:
				m.FocusComponent(LeftTable)
			}
		case "right", "left":
			// Quickly change from between left and right panels
			switch activeComponent {
			case LeftTable:
				m.FocusComponent(RightViewport)
			case RightViewport:
				m.FocusComponent(LeftTable)
			}
		}
	}

	// The message was not consumed by previous code.
	// Delegate the msg to the active panel
	switch activeComponent {
	case LeftTable:
		previousBookPtr := m.SelectedBook()

		m.table, cmd = m.table.Update(msg)
		cmds = append(cmds, cmd)

		// Check if the selected book has changed
		newBookPtr := m.SelectedBook()
		if hasBookChanged(previousBookPtr, newBookPtr) {
			// The selected book have changed.
			// Force updating the right viewport with new automatically wrapped content.
			m.onSelectedBookChanged()

			// And move the right viewport to the top of the view
			m.viewport.GotoTop()
		}
	case RightViewport:
		m.viewport, cmd = m.viewport.Update(msg)
		cmds = append(cmds, cmd)
	case SearchText:
		previousFilter := m.searchText.Value()

		m.searchText, cmd = m.searchText.Update(msg)
		cmds = append(cmds, cmd)

		newFilter := m.searchText.Value()

		// If the filter has changed
		if previousFilter != newFilter {
			m.onFilterChanged()
		}

	default:
	}

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	if !m.ready {
		return "Initializing UI..."
	}

	// Get current component
	activeComponent := m.ActiveComponent()

	// Define colors for left and right borders.
	// Set both borders as unfocused by default
	leftBorderColor := m.theme.unfocusedBorderColor
	rightBorderColor := m.theme.unfocusedBorderColor
	// Set active border to the focused style
	switch activeComponent {
	case LeftTable, SearchText:
		leftBorderColor = m.theme.focusedBorderColor
	case RightViewport:
		fallthrough
	default:
		rightBorderColor = m.theme.focusedBorderColor
	}

	leftStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(leftBorderColor).
		Padding(m.theme.panelsPadding...)

	rightStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(rightBorderColor).
		Padding(m.theme.panelsPadding...)

	leftTitle := m.theme.headerTextStyle.Render("Books")
	rightTitle := m.theme.headerTextStyle.Render("Summary")

	positionIndicatorText := fmt.Sprintf("[%d/%d]", m.table.Cursor()+1, len(m.table.Rows()))

	leftPanel := leftStyle.Render(leftTitle + "\n\n" + m.table.View() + "\n\n" + positionIndicatorText)
	rightPanel := rightStyle.Render(rightTitle + "\n\n" + m.renderViewportWithScrollbar())

	body := lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, rightPanel)

	searchLabel := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF")).Render("Search: ")
	searchLabel += " " + m.searchText.View()

	help := lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render(
		"Tab/←/→: Switch Active Panel  |  ↑/↓: Scroll  |  q: Quit",
	)

	return body + "\n" + searchLabel + "\n" + help + "\n"
}

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	_, err := p.Run()
	if err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}
