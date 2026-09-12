package main

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Book struct {
	Title       string
	Author      string
	Year        int
	Description string
}

// Force model to always implements interface tea.Model
var _ tea.Model = (*model)(nil)

type model struct {
	allBooks     []Book // These are the Books raw data. This slice is never rendered in the UI.
	filtered     []Book // These are the Books that are rendered. Even when we do not filter, we copy allBooks to filtered. See `applyFilter()` for details.
	cursor       int    // Selected Book index in `filtered` list.
	scrollOffset int    // For list scrolling subwindow within filtered when too many books can not be rendered into the ui.
	searchQuery  string //
	searching    bool   // Searching mode. When disabled, show "text to explain how to trigger the search mode". When enabled, show the actual text filter.
	selectedBook *Book  // Selected Book when user presses ENTER
	view         View
}

const (
	topBorder            = 1 //
	header               = 1 // "Books"
	searchBar            = 1 //
	topSpacer            = 1 //
	booksCount           = 0 // Book titles
	bottomSpacer         = 1 //
	cursorIndexIndicator = 1 //
	cursorWidth          = 2 // 2 characters "> "
	bottomBorder         = 1 //
	helpTextHeight       = 1 // help text is 1 line. It must not ends with a \n
)

func initialModel() model {
	books := []Book{
		{"The Go Programming Language", "Alan A. A. Donovan & Brian W. Kernighan", 2015, "The authoritative resource for learning Go."},
		{"Concurrency in Go", "Katherine Cox-Buday", 2017, "Tools and techniques for developers looking to master concurrent code."},
		{"Learning Go", "Jon Bodner", 2021, "An idiomatic guide to real-world Go programming."},
		{"Go in Action", "William Kennedy", 2015, "An introduction to Go focusing on practical application development."},
		{"Head First Go", "Jay McGavren", 2019, "A brain-friendly guide to learning Go."},
		{"100 Go Mistakes and How to Avoid Them", "Tevfik Kazi", 2022, "Avoid common pitfalls and write cleaner Go code."},
		{"Designing Data-Intensive Applications", "Martin Kleppmann", 2017, "Deep dive into distributed systems architectures."},
		{"Clean Code", "Robert C. Martin", 2008, "A handbook of agile software craftsmanship."},
		{"The Pragmatic Programmer", "David Thomas & Andrew Hunt", 2019, "Your journey to mastery in software development."},
		{"Refactoring", "Martin Fowler", 2018, "Improving the design of existing code."},
		{"Domain-Driven Design", "Eric Evans", 2003, "Tackling complexity in the heart of software."},
		{"Building Microservices", "Sam Newman", 2021, "Designing fine-grained systems."},
	}

	m := model{
		allBooks: books,
		view: View{
			Panels: []*Panel{
				&Panel{},
				&Panel{},
			},
		},
	}

	// Set initial view size
	m.view.Size = Size{
		Width:  80,
		Height: 24,
	}

	// Since the view size has changed, recompute panels dimensions
	m.view.SplitPanelsVerticalyByRatio([]float32{0.4, 0.6})

	// Set constant settings for all panels
	for _, panel := range m.view.Panels {
		panel.Padding = Bounds{0, 1, 0, 1}
		panel.MinimumHeight = 6
	}

	m.view.Panels[0].MinimumWidth = 22
	m.view.Panels[1].MinimumWidth = 25

	m.applyFilter()

	return m
}

// Filters the list based on search query matching title or author
func (m *model) applyFilter() {
	if m.searchQuery == "" {
		m.filtered = make([]Book, len(m.allBooks))
		copy(m.filtered, m.allBooks)
	} else {
		m.filtered = nil
		q := strings.ToLower(m.searchQuery)
		for _, b := range m.allBooks {
			if strings.Contains(strings.ToLower(b.Title), q) || strings.Contains(strings.ToLower(b.Author), q) {
				m.filtered = append(m.filtered, b)
			}
		}
	}

	// Reset cursor/scroll bounds if filtered list shrunk
	if m.cursor >= len(m.filtered) {
		m.cursor = len(m.filtered) - 1
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
	m.scrollOffset = 0
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.view.Size.Width = msg.Width
		m.view.Size.Height = msg.Height

		// Remove 1 line to render helpText below the panels
		m.view.Size.Height -= 1
		if m.view.Size.Height < 0 {
			m.view.Size.Height = 0
		}

		// Since the view size has changed, recompute panels dimensions
		m.view.SplitPanelsVerticalyByRatio([]float32{0.4, 0.6})

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			if !m.searching {
				return m, tea.Quit
			}

		case "enter":
			if len(m.filtered) > 0 {
				m.selectedBook = &m.filtered[m.cursor]
			}
			return m, tea.Quit

		case "/", "tab":
			m.searching = true

		case "esc":
			m.searching = false
			m.searchQuery = ""
			m.applyFilter()

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if m.cursor < len(m.filtered)-1 {
				m.cursor++
			}

		case "backspace":
			if m.searching && len(m.searchQuery) > 0 {
				m.searchQuery = m.searchQuery[:len(m.searchQuery)-1]
				m.applyFilter()
			}

		default:
			// Capture typed search characters
			if m.searching && len(msg.String()) == 1 {
				m.searchQuery += msg.String()
				m.applyFilter()
			}
		}

		// Adjust dynamic vertical scroll window (offset tracking)
		visibleItems := m.getVisibleListHeight()
		if m.cursor < m.scrollOffset {
			m.scrollOffset = m.cursor
		} else if m.cursor >= m.scrollOffset+visibleItems {
			m.scrollOffset = m.cursor - visibleItems + 1
		}
	}
	return m, nil
}

// getVisibleListHeight returns the number of books that must be displayed in the left panel.
// This number can not be smaller than 3 and it's maximum value is limited based on the total height of the rendering area.
func (m model) getVisibleListHeight() int {

	// Content of left panel:
	// ```
	// Header
	// Search Bar
	//
	// List Content
	//
	// Cursor indicator [5/12]
	// ```
	//
	// Which is 5 constant lines + how many lines of the list we want to show

	contentSize := m.view.Panels[0].GetRenderSize()
	h := contentSize.Height - 5
	if h < 1 {
		return 1
	}
	return h
}

func (m model) View() string {
	return m.ViewOfficial()
}

func (m model) ViewOfficial() string {
	leftPanel := m.view.Panels[0]
	rightPanel := m.view.Panels[1]

	panelsContentHeight := leftPanel.GetRenderSize().Height

	// Styles
	leftStyle := leftPanel.Style().
		BorderForeground(lipgloss.Color("63"))

	rightStyle := rightPanel.Style().
		BorderForeground(lipgloss.Color("205"))

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		MarginBottom(1)

	cursorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("205")).
		Bold(true)

	searchPromptStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("214")).
		Italic(true)

	noResultFoundStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240"))

	// Build Left Column Content
	leftPanel.Content = ""

	// Render left header
	leftPanel.Content += lipgloss.NewStyle().Bold(true).Render("Books") + "\n"

	// Render the search prompt or the searched text based on current search mode.
	searchPrompt := "Search: " + m.searchQuery
	if m.searching {
		// When in search mode, show a fake cursor `█` to help guide the focus to this section.
		// This will tell the user that pressing letters is now possible.
		searchPrompt += searchPromptStyle.Foreground(lipgloss.Color("#00FF00")).Render("█")
	} else if m.searchQuery == "" {
		searchPrompt = "(Press '/' to search)"
	}
	leftPanel.Content += searchPromptStyle.Render(searchPrompt) + "\n\n"

	visibleCount := m.getVisibleListHeight()

	if len(m.filtered) == 0 {
		leftPanel.Content += noResultFoundStyle.Render("No results found")
	} else {
		endIdx := m.scrollOffset + visibleCount
		if endIdx > len(m.filtered) {
			endIdx = len(m.filtered)
		}

		// Render each filtered books.
		// We skip some books (scrollOffset) if the number filtered books exceed how many book can fit in the left column.
		// We then only render a subsection (a subwindow) of the filtered books.
		for i := m.scrollOffset; i < endIdx; i++ {
			book := m.filtered[i]

			contentSize := leftPanel.GetRenderSize()
			displayTitleMaxLen := contentSize.Width - cursorWidth // Account for indicator
			displayTitle := truncateTextWidth(book.Title, displayTitleMaxLen)

			// If this Books is the selected book...
			cursor := "  "
			if m.cursor == i {
				// Set the selected cursor & colorize with a style this book's title
				cursor = "> "
				displayTitle = cursorStyle.Render(displayTitle)
			}

			// Render the Book
			leftPanel.Content += cursor + displayTitle + "\n"
		}

		// Scroll indicator hint if more items exist
		if len(m.filtered) > visibleCount {
			scrollIndicatorStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("241"))
			text := fmt.Sprintf(" [%d/%d]", m.cursor+1, len(m.filtered))
			leftPanel.Content += "\n" + scrollIndicatorStyle.Render(text)
		}
	}

	// Build Right Column Content
	rightPanel.Content = ""
	if len(m.filtered) > 0 && m.cursor < len(m.filtered) {
		selected := m.filtered[m.cursor]
		descStyle := lipgloss.NewStyle().Width(rightPanel.GetRenderSize().Width)

		rightPanel.Content = titleStyle.Render(selected.Title) + "\n" +
			fmt.Sprintf("Author: %s\n", selected.Author) +
			fmt.Sprintf("Year:   %d\n\n", selected.Year) +
			descStyle.Render(fmt.Sprintf("Details:\n%s", selected.Description))
	} else {
		rightPanel.Content = lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("No book selected.")
	}

	// Truncate content if required
	leftPanel.Content = truncateTextHeight(leftPanel.Content, panelsContentHeight)
	rightPanel.Content = truncateTextHeight(rightPanel.Content, panelsContentHeight)

	// Join both columns
	columns := lipgloss.JoinHorizontal(
		lipgloss.Top,
		leftStyle.Render(leftPanel.Content),
		rightStyle.Render(rightPanel.Content),
	)

	helpText := "\n ↑/↓: Navigate  |  Enter: Select  |  /: Search  |  Esc: Clear  |  q: Quit"
	return columns + helpText
}

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	m, err := p.Run()
	if err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}

	// Print final user selection on stdout after exiting alt screen
	finalModel := m.(model)
	if finalModel.selectedBook != nil {
		fmt.Printf("\nSelected Book: \"%s\" by %s (%d)\n",
			finalModel.selectedBook.Title,
			finalModel.selectedBook.Author,
			finalModel.selectedBook.Year,
		)
	} else {
		fmt.Println("\nNo book was selected.")
	}
}
