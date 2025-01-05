package handler

import (
	"sync"

	"github.com/charmbracelet/lipgloss"
)

// Types
type SelectedRepository struct {
	RepositoryName string
	WorkflowName   string
	BranchName     string
}

// Constants
const (
	MinTerminalWidth  = 102
	MinTerminalHeight = 24

	DefaultTableHeight    = 13
	DefaultTerminalWidth  = 80
	DefaultTerminalHeight = 24

	ColorPrimary   = "#3b698f"
	ColorSecondary = "#ff0055"
	ColorError     = "#ff0000"
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
