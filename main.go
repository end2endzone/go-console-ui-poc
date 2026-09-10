package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Book struct {
	Title  string
	Author string
	Year   string
	Desc   string
}

type model struct {
	books  []Book
	cursor int
}

func initialModel() model {
	return model{
		books: []Book{
			{Title: "The Go Programming Language", Author: "Alan A. A. Donovan", Year: "2015", Desc: "An authoritative guide to writing clear, idiomatic Go program code."},
			{Title: "Concurrency in Go", Author: "Katherine Cox-Buday", Year: "2017", Desc: "Deep dive into concurrency patterns, primitives, and memory access in Go."},
			{Title: "Designing Data-Intensive Applications", Author: "Martin Kleppmann", Year: "2017", Desc: "A comprehensive guide to data system architectures, scalability, and reliability."},
			{Title: "Clean Code", Author: "Robert C. Martin", Year: "2008", Desc: "A handbook of agile software craftsmanship focused on writing readable, maintainable code."},
		},
		cursor: 0,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.books)-1 {
				m.cursor++
			}
		}
	}
	return m, nil
}

func (m model) View() string {
	// Column Styles
	leftStyle := lipgloss.NewStyle().
		Width(35).
		Height(10).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(0, 1)

	rightStyle := lipgloss.NewStyle().
		Width(45).
		Height(10).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")).
		Padding(0, 1)

	// Build Left Column (Book Titles)
	leftContent := "SELECT A BOOK:\n\n"
	for i, book := range m.books {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
			leftContent += fmt.Sprintf("[%s] %s\n", cursor, lipgloss.NewStyle().Bold(true).Render(book.Title))
		} else {
			leftContent += fmt.Sprintf(" %s  %s\n", cursor, book.Title)
		}
	}

	// Build Right Column (Selected Details)
	selected := m.books[m.cursor]
	rightContent := fmt.Sprintf(
		"DETAILS:\n\nTitle:  %s\nAuthor: %s\nYear:   %s\n\n%s",
		selected.Title,
		selected.Author,
		selected.Year,
		selected.Desc,
	)

	// Join both columns side-by-side horizontally
	view := lipgloss.JoinHorizontal(
		lipgloss.Top,
		leftStyle.Render(leftContent),
		rightStyle.Render(rightContent),
	)

	return view + "\n\nPress 'q' or 'Ctrl+C' to exit. Use ↑/↓ to navigate."
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}
