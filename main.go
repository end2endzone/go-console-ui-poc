package main

import (
	"fmt"
	"os"

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
	books    []Book
	cursor   int
	selected int
}

func initialModel() model {
	return model{
		books: []Book{
			{
				Title:       "The Go Programming Language",
				Author:      "Alan A. A. Donovan & Brian W. Kernighan",
				Year:        2015,
				Description: "The authoritative resource for learning the Go programming language.",
			},
			{
				Title:       "Concurrency in Go",
				Author:      "Katherine Cox-Buday",
				Year:        2017,
				Description: "Tools and techniques for developers looking to master concurrent code.",
			},
			{
				Title:       "Learning Go",
				Author:      "Jon Bodner",
				Year:        2021,
				Description: "An idiomatic guide to real-world Go programming.",
			},
			{
				Title:       "Go in Action",
				Author:      "William Kennedy",
				Year:        2015,
				Description: "An introduction to Go focusing on practical application development.",
			},
		},
		cursor:   0,
		selected: 0,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
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
	// Styles
	leftStyle := lipgloss.NewStyle().
		Width(30).
		Height(12).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")).
		Padding(1)

	rightStyle := lipgloss.NewStyle().
		Width(45).
		Height(12).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("205")).
		Padding(1)

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		MarginBottom(1)

	cursorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("205")).
		Bold(true)

	// Left Column: Book List
	leftContent := lipgloss.NewStyle().Bold(true).Render("Books") + "\n\n"
	for i, book := range m.books {
		cursor := "  "
		item := book.Title

		if m.cursor == i {
			cursor = "> "
			item = cursorStyle.Render(book.Title)
		}
		leftContent += fmt.Sprintf("%s%s\n", cursor, item)
	}

	// Right Column: Book Details
	selectedBook := m.books[m.cursor]
	rightContent := titleStyle.Render(selectedBook.Title) + "\n" +
		fmt.Sprintf("Author: %s\n", selectedBook.Author) +
		fmt.Sprintf("Year:   %d\n\n", selectedBook.Year) +
		fmt.Sprintf("Details: %s\n", selectedBook.Description)

	// Join both columns horizontally
	columns := lipgloss.JoinHorizontal(
		lipgloss.Top,
		leftStyle.Render(leftContent),
		rightStyle.Render(rightContent),
	)

	helpText := "\n Use ↑/↓ or j/k to navigate • Press 'q' to exit"
	return columns + helpText + "\n"
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}
