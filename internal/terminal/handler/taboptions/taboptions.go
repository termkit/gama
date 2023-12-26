package taboptions

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Options struct {
	Style lipgloss.Style

	options       []string
	optionsAction []string

	isTabSelected bool

	cursor int
}

func NewOptions() *Options {
	var b = lipgloss.RoundedBorder()
	b.Right = "├"
	b.Left = "┤"

	var OptionsStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("15")).
		Align(lipgloss.Center).Padding(0, 1, 0, 1).
		Border(b)

	var initalOptions = []string{
		"Idle",
	}
	var initalOptionsAction = []string{
		"Idle",
	}

	return &Options{
		Style:         OptionsStyle,
		options:       initalOptions,
		optionsAction: initalOptionsAction,
	}
}

func (o *Options) AddOption(option string) {
	var optionWithNumber string
	var optionNumber = len(o.options)
	optionWithNumber = fmt.Sprintf("%d) %s", optionNumber, option)
	o.options = append(o.options, optionWithNumber)
	o.optionsAction = append(o.optionsAction, optionWithNumber)
}

func (o *Options) Init() tea.Cmd {
	return nil
}

func (o *Options) resetOptionsWithOriginal() {
	if o.isTabSelected {
		return
	}
	o.isTabSelected = true
	//time.Sleep(2 * time.Second)
	//copy(o.optionsAction, o.options)
	//o.cursor = 0
	//o.isTabSelected = false
	for i := 3; i >= 0; i-- {
		o.optionsAction[0] = fmt.Sprintf("> %ds", i)
		time.Sleep(1 * time.Second)
	}
	o.optionsAction[0] = "Idle"
	o.cursor = 0
	o.isTabSelected = false
}

func (o *Options) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "1":
			o.cursor = 1
			//o.isTabSelected = true
			go o.resetOptionsWithOriginal()
		case "2":
			o.cursor = 2
			//o.isTabSelected = true
			go o.resetOptionsWithOriginal()
		case "3":
			o.cursor = 3
			//o.isTabSelected = true
			go o.resetOptionsWithOriginal()
		case "4":
			o.cursor = 4
			//o.isTabSelected = true
			go o.resetOptionsWithOriginal()
		case "enter":
			o.cursor = 0 // debug
		}

	}

	return o, cmd
}

func (o *Options) View() string {
	var opts []string
	for i, option := range o.optionsAction {
		var style lipgloss.Style
		isActive := i == o.cursor
		if isActive {
			style = o.Style.Copy().
				Foreground(lipgloss.Color("15")).
				BorderForeground(lipgloss.Color("150"))
		} else {
			style = o.Style.Copy().
				Foreground(lipgloss.Color("15")).
				BorderForeground(lipgloss.Color("240"))
		}
		opts = append(opts, style.Render(option))
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, opts...)
}
