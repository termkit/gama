package style

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	WindowStyleOrange = lipgloss.NewStyle().BorderForeground(lipgloss.Color("#ffaf00")).Border(lipgloss.RoundedBorder())
	WindowStyleRed    = lipgloss.NewStyle().BorderForeground(lipgloss.Color("9")).Border(lipgloss.RoundedBorder())
	WindowStyleGreen  = lipgloss.NewStyle().BorderForeground(lipgloss.Color("10")).Border(lipgloss.RoundedBorder())
	WindowStyleGray   = lipgloss.NewStyle().BorderForeground(lipgloss.Color("240")).Border(lipgloss.NormalBorder())
	WindowStyleWhite  = lipgloss.NewStyle().BorderForeground(lipgloss.Color("255")).Border(lipgloss.NormalBorder())

	WindowStyleHelp     = WindowStyleGray.Margin(0, 0, 0, 1).Padding(0, 2, 0, 2).Border(lipgloss.RoundedBorder())
	WindowStyleError    = WindowStyleRed.Margin(0, 0, 0, 1).Padding(0, 2, 0, 2).Border(lipgloss.RoundedBorder())
	WindowStyleProgress = WindowStyleOrange.Margin(0, 0, 0, 1).Padding(0, 2, 0, 2).Border(lipgloss.RoundedBorder())
	WindowStyleSuccess  = WindowStyleGreen.Margin(0, 0, 0, 1).Padding(0, 2, 0, 2).Border(lipgloss.RoundedBorder())
	WindowStyleDefault  = WindowStyleWhite.Margin(0, 0, 0, 1).Padding(0, 2, 0, 2).Border(lipgloss.RoundedBorder())
)

var (
	TitleStyleActive = func() lipgloss.Style {
		b := lipgloss.DoubleBorder()
		b.Right = "├"
		b.Left = "┤"
		return lipgloss.NewStyle().BorderStyle(b).Padding(0, 2).BorderForeground(lipgloss.Color("205"))
	}()

	TitleStyleInactive = func() lipgloss.Style {
		b := lipgloss.RoundedBorder()
		b.Right = "├"
		b.Left = "┤"
		return lipgloss.NewStyle().BorderStyle(b).Padding(0, 2).BorderForeground(lipgloss.Color("255"))
	}()

	TitleStyleDisabled = func() lipgloss.Style {
		b := lipgloss.RoundedBorder()
		b.Right = "├"
		b.Left = "┤"
		return lipgloss.NewStyle().BorderStyle(b).Padding(0, 2).BorderForeground(lipgloss.Color("240")).Foreground(lipgloss.Color("240"))
	}()
)
