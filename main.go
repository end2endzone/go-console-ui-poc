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

type model struct {
	allBooks     []Book
	filtered     []Book
	cursor       int // Selected index in `filtered`
	scrollOffset int // For list scrolling window
	searchQuery  string
	searching    bool
	selectedBook *Book
	width        int
	height       int
}

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
		width:    80,
		height:   24,
	}
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
		m.width = msg.Width
		m.height = msg.Height

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

func (m model) getVisibleListHeight() int {
	// Total available height minus borders, margin, header, search bar, and help line
	h := m.height - 8
	if h < 3 {
		return 3
	}
	return h
}

// Truncates text with trailing ellipsis to fit within column bounds
func truncateText(text string, maxLen int) string {
	if maxLen <= 3 {
		return "..."
	}
	runes := []rune(text)
	if len(runes) > maxLen {
		return string(runes[:maxLen-3]) + "..."
	}
	return text
}

func (m model) View() string {
	helpHeight := 2
	availableHeight := m.height - helpHeight - 2
	if availableHeight < 6 {
		availableHeight = 6
	}

	// Dynamic column calculation
	leftWidth := int(float64(m.width) * 0.38)
	rightWidth := m.width - leftWidth - 4

	if leftWidth < 22 {
		leftWidth = 22
	}
	if rightWidth < 25 {
		rightWidth = 25
	}

	// Styles
	leftStyle := lipgloss.NewStyle().
		Width(leftWidth).
		Height(availableHeight).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")).
		Padding(0, 1)

	rightStyle := lipgloss.NewStyle().
		Width(rightWidth).
		Height(availableHeight).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("205")).
		Padding(0, 1)

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

	// Build Left Column Content
	searchPrompt := "Search: " + m.searchQuery
	if m.searching {
		searchPrompt += "█" // Fake cursor line
	} else if m.searchQuery == "" {
		searchPrompt = "(Press '/' to search)"
	}

	leftContent := lipgloss.NewStyle().Bold(true).Render("Books") + "\n" +
		searchPromptStyle.Render(searchPrompt) + "\n\n"

	visibleCount := m.getVisibleListHeight()
	maxTitleLen := leftWidth - 5 // Account for padding & indicator

	if len(m.filtered) == 0 {
		leftContent += lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(" No results found")
	} else {
		endIdx := m.scrollOffset + visibleCount
		if endIdx > len(m.filtered) {
			endIdx = len(m.filtered)
		}

		for i := m.scrollOffset; i < endIdx; i++ {
			book := m.filtered[i]
			cursor := "  "
			displayTitle := truncateText(book.Title, maxTitleLen)

			if m.cursor == i {
				cursor = "> "
				displayTitle = cursorStyle.Render(displayTitle)
			}
			leftContent += fmt.Sprintf("%s%s\n", cursor, displayTitle)
		}

		// Scroll indicator hint if more items exist
		if len(m.filtered) > visibleCount {
			leftContent += fmt.Sprintf("\n%s", lipgloss.NewStyle().Foreground(lipgloss.Color("241")).
				Render(fmt.Sprintf(" [%d/%d]", m.cursor+1, len(m.filtered))))
		}
	}

	// Build Right Column Content
	var rightContent string
	if len(m.filtered) > 0 && m.cursor < len(m.filtered) {
		selected := m.filtered[m.cursor]
		descStyle := lipgloss.NewStyle().Width(rightWidth - 4)

		rightContent = titleStyle.Render(selected.Title) + "\n" +
			fmt.Sprintf("Author: %s\n", selected.Author) +
			fmt.Sprintf("Year:   %d\n\n", selected.Year) +
			descStyle.Render(fmt.Sprintf("Details:\n%s", selected.Description))
	} else {
		rightContent = lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("No book selected.")
	}

	// Join both columns
	columns := lipgloss.JoinHorizontal(
		lipgloss.Top,
		leftStyle.Render(leftContent),
		rightStyle.Render(rightContent),
	)

	helpText := "\n ↑/↓: Navigate • Enter: Select • /: Search • Esc: Clear • q: Quit"
	return columns + helpText + "\n"
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
