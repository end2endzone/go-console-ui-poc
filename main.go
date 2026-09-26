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

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/end2endzone/go-console-ui-poc/internal/tui/components/titledborderedpanels"
	"github.com/end2endzone/go-console-ui-poc/internal/tui/components/titledborderedpanels/textpanel"
	"github.com/end2endzone/go-console-ui-poc/internal/tui/components/titledborderedpanels/textpanelwithscrollbar"
	"github.com/end2endzone/go-console-ui-poc/internal/tui/layout"
	"github.com/end2endzone/go-console-ui-poc/internal/tui/navigation"
)

type Book struct {
	Title       string `json:"title"`
	Author      string `json:"author"`
	Year        int    `json:"year"`
	Description string `json:"description"`
}

// Force model to always implements interface tea.Model
var _ tea.Model = (*model)(nil)

// Focusable elements of the UI
type UIComponent int

const (
	UnknownPanelId UIComponent = iota // 0
	BooksPanelId                      // 1
	SummaryPanelId                    // 2
	SearchPanelId                     // 3
	ComponentCount = 3
)

// Declare all panel names as constants
const (
	LeftColumnName   string = "LeftColumn"
	SearchPanelName  string = "SearchPanel"
	BooksPanelName   string = "BooksPanel"
	SummaryPanelName string = "SummaryPanel"
)

type PanelStyles struct {
	TitleStyle       lipgloss.Style
	BorderStyle      lipgloss.Style
	BooksTableStyles table.Styles
}

type Theme struct {
	FocusedPanel   PanelStyles
	UnfocusedPanel PanelStyles
}

func NewTheme() Theme {
	theme := Theme{
		FocusedPanel: PanelStyles{
			BorderStyle:      lipgloss.NewStyle().Padding(0, 1).Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("63")),
			TitleStyle:       lipgloss.NewStyle().Padding(0, 1).Background(lipgloss.Color("63")).Bold(true),
			BooksTableStyles: table.DefaultStyles(),
		},
		UnfocusedPanel: PanelStyles{
			BorderStyle:      lipgloss.NewStyle().Padding(0, 1).Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("240")),
			TitleStyle:       lipgloss.NewStyle().Padding(0, 1).Foreground(lipgloss.Color("205")),
			BooksTableStyles: table.DefaultStyles(),
		},
	}

	// Set table.Styles for focused tables
	{
		tmp := table.DefaultStyles()
		tmp.Header = tmp.Header.
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("240")).
			BorderBottom(true).
			Bold(true)
		tmp.Selected = tmp.Selected.
			Foreground(theme.FocusedPanel.TitleStyle.GetForeground()). // use same colors as the border
			Background(theme.FocusedPanel.TitleStyle.GetBackground()). // use same colors as the border
			Bold(true)
		theme.FocusedPanel.BooksTableStyles = tmp
	}

	// Set table.Styles for unfocused tables
	// Basicaly, remove background colors when unselected.
	theme.UnfocusedPanel.BooksTableStyles.Header = theme.FocusedPanel.BooksTableStyles.Header
	theme.UnfocusedPanel.BooksTableStyles.Selected = theme.FocusedPanel.BooksTableStyles.Selected.
		Foreground(BrightBlue).
		UnsetBackground()

	return theme
}

type model struct {
	theme         Theme
	books         []Book
	filteredBooks []*Book
	navigation    navigation.Model
	table         table.Model
	searchText    textinput.Model
	layoutTree    *layout.Node
	panels        struct {
		books   textpanel.Model
		search  textpanel.Model
		summary textpanelwithscrollbar.Model
	}
	nodes struct {
		search  *layout.Node
		books   *layout.Node
		summary *layout.Node
	}
	width  int
	height int
}

// Theme display constants
const (
	borderWidth           = 1
	leftRightPaddingWidth = 1
	scrollBarWidth        = 2                                     // For example " █", note the space before the scroll bar cursor
	minTableItems         = 1                                     // Minimum number of data rows (excluding table's header rows)
	tableHeaderHeight     = 2                                     // Tables renders columns names in a 2 lines header
	minTableHeight        = minTableItems + tableHeaderHeight     // A full table height includes 2 line table header
	minContentHeight      = 4 + minTableHeight                    // For right panel, that is: + 2 lines for the table, minTableItems, 2 lines cursor indicator footer
	minViewportTextWidth  = 1                                     // Minimum width of the text (exclusing the scroll bars characters)
	minRightViewportWidth = minViewportTextWidth + scrollBarWidth // 1 character wide + scroll bar
	minRightWidth         = minRightViewportWidth +
		2*borderWidth +
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

// tree describes the layout of the panels.
func tree() *layout.Node {
	return layout.Row(layout.Policy{}, // root's own Size is ignored
		layout.ColWithName(LeftColumnName, layout.Policy{Fixed: 40}, // column 1: fixed 40 cols wide
			layout.Leaf(BooksPanelName, layout.Policy{Grow: 1, Min: minContentHeight}), // fills remaining height, minimum height 9
			layout.Leaf(SearchPanelName, layout.Policy{Fixed: 3}),                      // fixed height
		),
		layout.Leaf(SummaryPanelName, layout.Policy{Grow: 1}), // column 2: fills remaining width
	)
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

	searchText := textinput.New()
	searchText.Placeholder = "filter"
	searchText.CharLimit = 156
	searchText.Width = 40

	m := model{
		theme:      NewTheme(),
		books:      books,
		table:      t,
		searchText: searchText,
		layoutTree: tree(),
	}

	// Fill Table
	m.FillBooksTable("")

	// Now that Books table is filled, we know how wide it is.
	// Set left panels width based on this.
	tableWidth := getTableColumnsWidth(&m.table)
	tableWidth += 4 // +2 for padding (1 on each side), +2 borders
	m.layoutTree.Find(LeftColumnName).Policy.Fixed = tableWidth

	// Pre-find the leaf panels
	m.nodes.search = m.layoutTree.Find(SearchPanelName)
	m.nodes.books = m.layoutTree.Find(BooksPanelName)
	m.nodes.summary = m.layoutTree.Find(SummaryPanelName)

	// Set panel's and their titles
	m.panels.books = textpanel.New()
	m.panels.search = textpanel.New()
	m.panels.summary = textpanelwithscrollbar.New()

	m.panels.books.Title = "Books"
	m.panels.search.Title = "Search"
	m.panels.summary.Title = "Summary"

	// Setup navigation panels
	m.navigation.IDs = []int{int(BooksPanelId), int(SearchPanelId), int(SummaryPanelId)}
	m.navigation.SetFocusedComponentByValue(int(BooksPanelId))

	// Focus books by default
	activeComponent := UIComponent(m.navigation.CurrentFocusedComponent())
	m.FocusComponent(activeComponent)

	return m
}

// ActiveComponent return the active focused component in the main UI.
func (m *model) ActiveComponent() UIComponent {
	if m.table.Focused() {
		return BooksPanelId
	} else if m.panels.summary.Focused() {
		return SummaryPanelId
	} else if m.searchText.Focused() {
		return SearchPanelId
	}

	return UnknownPanelId
}

// FocusComponent focuses the given component and blur other components.
func (m *model) FocusComponent(c UIComponent) {
	switch c {
	case BooksPanelId:
		m.table.Focus()
		m.panels.summary.Blur()
		m.searchText.Blur()
	case SummaryPanelId:
		m.table.Blur()
		m.panels.summary.Focus()
		m.searchText.Blur()
	case SearchPanelId:
		m.table.Blur()
		m.panels.summary.Blur()
		m.searchText.Focus()
	default:
		m.table.Blur()
		m.panels.summary.Blur()
		m.searchText.Blur()
	}

	// Sync navigation pane
	m.navigation.SetFocusedComponentByValue(int(c))
}

// FocusNextComponent focuses the next component.
func (m *model) FocusNextComponent() {
	// Move navigation's focus to the next component
	newFocusedComponent := UIComponent(m.navigation.NextFocusedComponent())

	// Make it actually focused and blur others
	m.FocusComponent(newFocusedComponent)
}

// FocusPreviousComponent focuses the previous component.
func (m *model) FocusPreviousComponent() {
	// Move navigation's focus to the previous component
	newFocusedComponent := UIComponent(m.navigation.PreviousFocusedComponent())

	// Make it actually focused and blur others
	m.FocusComponent(newFocusedComponent)
}

// GetPanels returns the list of all panels in the model
func (m *model) GetPanels() []titledborderedpanels.TitledBorderedPanel {
	return []titledborderedpanels.TitledBorderedPanel{
		&m.panels.books,
		&m.panels.search,
		&m.panels.summary,
	}
}

// GetFocusedPanel returns the panels that contains the currently focused component
func (m *model) GetPanelsByFocusState() (focusedPanel titledborderedpanels.TitledBorderedPanel, unfocusedPanels []titledborderedpanels.TitledBorderedPanel) {
	// Get current component
	activeComponent := m.ActiveComponent()

	switch activeComponent {
	case BooksPanelId:
		focusedPanel = &m.panels.books
		unfocusedPanels = []titledborderedpanels.TitledBorderedPanel{
			&m.panels.search,
			&m.panels.summary,
		}
		return
	case SummaryPanelId:
		focusedPanel = &m.panels.summary
		unfocusedPanels = []titledborderedpanels.TitledBorderedPanel{
			&m.panels.books,
			&m.panels.search,
		}
		return
	case SearchPanelId:
		focusedPanel = &m.panels.search
		unfocusedPanels = []titledborderedpanels.TitledBorderedPanel{
			&m.panels.books,
			&m.panels.summary,
		}
		return
	default:
		return nil, m.GetPanels()
	}
}

// GetPanelFromId gets the matching panel given a panel id.
// Returns nil if the panel id is unknown.
func (m *model) GetPanelFromId(id UIComponent) titledborderedpanels.TitledBorderedPanel {
	switch id {
	case SearchPanelId:
		return &m.panels.search
	case BooksPanelId:
		return &m.panels.books
	case SummaryPanelId:
		return &m.panels.summary
	}
	return nil
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
		m.panels.summary.SetContent(bookPtr.Description)
	} else {
		// There is no book selected, clear the right panel content
		m.panels.summary.SetContent("")
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
		m.width = msg.Width //results in a summary panel that is 90x11
		m.height = msg.Height

		// Resolve panels size based on available space
		layout.Resolve(m.layoutTree, m.width, m.height-2) // 2 lines for the help string (the help string itself and a final \n)

		tableHeight := m.nodes.books.Dimension.H - 2*borderWidth - 2 // 2 lines cursor indicator footer
		m.table.SetHeight(tableHeight)

		// Copy resolved node sizes to their matching panels
		m.panels.books.SetSize(
			m.nodes.books.Dimension.W,
			m.nodes.books.Dimension.H,
		)
		m.panels.search.SetSize(
			m.nodes.search.Dimension.W,
			m.nodes.search.Dimension.H,
		)
		m.panels.summary.SetSize(
			m.nodes.summary.Dimension.W,
			m.nodes.summary.Dimension.H,
		)

		// The right viewport dimensions have changed.
		// Force updating the right viewport with new automatically wrapped content.
		m.onSelectedBookChanged()

		return m, tea.Batch(cmds...) // Message consumed
	case tea.KeyMsg:
		switch msg.String() {
		case "q":
			if activeComponent != SearchPanelId {
				// only allow q to quit when its not the search string that has focus
				return m, tea.Quit
			}
		case "esc", "ctrl+c":
			return m, tea.Quit

		case "shift+tab":
			m.FocusPreviousComponent()
			return m, tea.Batch(cmds...) // Message consumed
		case "tab":
			m.FocusNextComponent()
			return m, tea.Batch(cmds...) // Message consumed

		case "right", "left":
			// Quickly change from between left and right panels
			switch activeComponent {
			case BooksPanelId:
				m.FocusComponent(SummaryPanelId)
			case SummaryPanelId:
				m.FocusComponent(BooksPanelId)
			}
			return m, tea.Batch(cmds...) // Message consumed
		}
	}

	// The message was not consumed by previous code.
	// Delegate the msg to the active panel
	switch activeComponent {
	case BooksPanelId:
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
			m.panels.summary.GotoTop()
		}
	case SummaryPanelId:
		abstractModel, cmd := m.panels.summary.Update(msg)
		summaryModel, ok := abstractModel.(*textpanelwithscrollbar.Model)
		if ok {
			m.panels.summary = *summaryModel
		} else {
			panic("Interface update returned an unexpected underlying type")
		}
		cmds = append(cmds, cmd)
	case SearchPanelId:
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
	// Get all panels by focus state
	focusedPanel, unfocusedPanels := m.GetPanelsByFocusState()

	// handle focus panel
	focusedPanel.SetPanelStyle(m.theme.FocusedPanel.BorderStyle)
	focusedPanel.SetTitleStyle(m.theme.FocusedPanel.TitleStyle)

	// handle unfocused panels
	for _, p := range unfocusedPanels {
		p.SetPanelStyle(m.theme.UnfocusedPanel.BorderStyle)
		p.SetTitleStyle(m.theme.UnfocusedPanel.TitleStyle)
	}

	// Update the books table styles based on focused panel
	if &m.panels.books == focusedPanel {
		m.table.SetStyles(m.theme.FocusedPanel.BooksTableStyles)
	} else {
		m.table.SetStyles(m.theme.UnfocusedPanel.BooksTableStyles)
	}

	positionIndicatorText := fmt.Sprintf("[%d/%d]", m.table.Cursor()+1, len(m.table.Rows()))

	// Panel's content
	m.panels.search.SetContent(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF")).Render("Search: ") + " " + m.searchText.View())
	m.panels.books.SetContent(m.table.View() + "\n" + positionIndicatorText)
	// summary panel is already updated in onSelectedBookChanged()

	help := lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render(
		"Tab/←/→: Switch Active Panel  |  ↑/↓: Scroll  |  q: Quit",
	)

	// Join all panels
	leftColumn := lipgloss.JoinVertical(lipgloss.Left, m.panels.books.View(), m.panels.search.View())
	panels := lipgloss.JoinHorizontal(lipgloss.Top, leftColumn, m.panels.summary.View())
	body := panels + "\n" + help

	/*debug := true
	if debug {
		viewCount++
		debugging.DumpStringToFile(fmt.Sprintf("debug/leftColumn-%d.txt", viewCount), leftColumn)
		debugging.DumpStringToFile(fmt.Sprintf("debug/panels-%d.txt", viewCount), panels)
		debugging.DumpStringToFile(fmt.Sprintf("debug/body-%d.txt", viewCount), body)

		debugging.DumpRenderingWithoutStylesToFile("debug/m.panels.summary.View().txt", m.panels.summary.View())
		debugging.DumpRenderingWithoutStylesToFile("debug/m.panels.summaryPanel.View().txt", m.panels.summaryPanel.View())
	}*/

	return body
}

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	_, err := p.Run()
	if err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}
