package handler

import (
	"sync"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
)

// SelectedRepository is a struct that holds the selected repository, workflow, and branch
// It is a shared state between the different tabs
type SelectedRepository struct {
	RepositoryName string
	WorkflowName   string
	BranchName     string
}

// Constants
const (
	MinTerminalWidth  = 102
	MinTerminalHeight = 24
)

// Styles
var (
	WindowStyleOrange = lipgloss.NewStyle().BorderForeground(lipgloss.Color("#ffaf00")).Border(lipgloss.RoundedBorder())
	WindowStyleRed    = lipgloss.NewStyle().BorderForeground(lipgloss.Color("9")).Border(lipgloss.RoundedBorder())
	WindowStyleGreen  = lipgloss.NewStyle().BorderForeground(lipgloss.Color("10")).Border(lipgloss.RoundedBorder())
	WindowStyleGray   = lipgloss.NewStyle().BorderForeground(lipgloss.Color("240")).Border(lipgloss.RoundedBorder())
	WindowStyleWhite  = lipgloss.NewStyle().BorderForeground(lipgloss.Color("255")).Border(lipgloss.RoundedBorder())

	WindowStyleHelp     = WindowStyleGray.Margin(0, 0, 0, 0).Padding(0, 2, 0, 2)
	WindowStyleError    = WindowStyleRed.Margin(0, 0, 0, 0).Padding(0, 2, 0, 2)
	WindowStyleProgress = WindowStyleOrange.Margin(0, 0, 0, 0).Padding(0, 2, 0, 2)
	WindowStyleSuccess  = WindowStyleGreen.Margin(0, 0, 0, 0).Padding(0, 2, 0, 2)
	WindowStyleDefault  = WindowStyleWhite.Margin(0, 0, 0, 0).Padding(0, 2, 0, 2)
)

// Constructor
var (
	onceSelectedRepository sync.Once
	selectedRepository     *SelectedRepository
)

func NewSelectedRepository() *SelectedRepository {
	onceSelectedRepository.Do(func() {
		selectedRepository = &SelectedRepository{}
	})
	return selectedRepository
}

// Default table styles

func defaultTableStyles() table.Styles {
	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	return s
}

func defaultTableKeyMap() table.KeyMap {
	// We use "up" and "down" for both line up and down, we do not use "k" and "j" to prevent conflict with text input
	return table.KeyMap{
		LineUp:     key.NewBinding(key.WithKeys("up"), key.WithHelp("↑", "up")),
		LineDown:   key.NewBinding(key.WithKeys("down"), key.WithHelp("↓", "down")),
		PageUp:     key.NewBinding(key.WithKeys("pgup"), key.WithHelp("pgup", "page up")),
		PageDown:   key.NewBinding(key.WithKeys("pgdown", " "), key.WithHelp("pgdn", "page down")),
		GotoTop:    key.NewBinding(key.WithKeys("home"), key.WithHelp("home", "go to start")),
		GotoBottom: key.NewBinding(key.WithKeys("end"), key.WithHelp("end", "go to end")),
	}
}
