package types

import (
	"github.com/charmbracelet/lipgloss"
)

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
