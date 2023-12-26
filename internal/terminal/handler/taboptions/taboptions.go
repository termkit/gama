package taboptions

// TODO : set Idle to Wait if options isn't ready to use

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Options struct {
	Style lipgloss.Style

	status OptionStatus

	options       []string
	optionsAction []string

	optionsWithFunc map[int]func()

	timer int

	isTabSelected bool

	cursor int
}

type OptionStatus string

const (
	Idle OptionStatus = "Idle"
	Wait OptionStatus = "Wait"
)

func (o OptionStatus) String() string {
	return string(o)
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
		Wait.String(),
	}
	var initalOptionsAction = []string{
		Wait.String(),
	}

	optionsWithFunc := make(map[int]func())
	optionsWithFunc[0] = func() {}

	return &Options{
		Style:           OptionsStyle,
		options:         initalOptions,
		optionsAction:   initalOptionsAction,
		optionsWithFunc: optionsWithFunc,
		status:          Wait,
	}
}

func (o *Options) SetStatus(status OptionStatus) {
	o.status = status
	o.options[0] = status.String()
	o.optionsAction[0] = status.String()
}

func (o *Options) AddOption(option string, action func()) {
	var optionWithNumber string
	var optionNumber = len(o.options)
	optionWithNumber = fmt.Sprintf("%d) %s", optionNumber, option)
	o.options = append(o.options, optionWithNumber)
	o.optionsAction = append(o.optionsAction, optionWithNumber)
	o.optionsWithFunc[optionNumber] = action
}

func (o *Options) executeOption() {
	go o.optionsWithFunc[o.cursor]()
	o.cursor = 0
	o.timer = -1
}

func (o *Options) Init() tea.Cmd {
	return nil
}

func (o *Options) resetOptionsWithOriginal() {
	if o.isTabSelected {
		return
	}
	o.isTabSelected = true
	o.timer = 3
	for o.timer >= 0 {
		o.optionsAction[0] = fmt.Sprintf("> %ds", o.timer)
		time.Sleep(1 * time.Second)
		o.timer--
	}
	o.optionsAction[0] = string(Idle)
	o.cursor = 0
	o.isTabSelected = false
}

func (o *Options) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	if o.status == Wait {
		return o, cmd
	}

	// TODO: If option less than 4, still count down the timer
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "1":
			o.cursor = 1
			go o.resetOptionsWithOriginal()
		case "2":
			o.cursor = 2
			go o.resetOptionsWithOriginal()
		case "3":
			o.cursor = 3
			go o.resetOptionsWithOriginal()
		case "4":
			o.cursor = 4
			go o.resetOptionsWithOriginal()
		case "enter":
			o.executeOption()
		}
	}

	return o, cmd
}

func (o *Options) View() string {
	var opts []string
	for i, option := range o.optionsAction {
		var style lipgloss.Style
		isActive := i == o.cursor
		if o.status == Wait {
			// orange
			style = o.Style.Copy().
				Foreground(lipgloss.Color("15")).
				BorderForeground(lipgloss.Color("208"))
		} else {
			if isActive {
				style = o.Style.Copy().
					Foreground(lipgloss.Color("15")).
					BorderForeground(lipgloss.Color("150"))
			} else {
				style = o.Style.Copy().
					Foreground(lipgloss.Color("15")).
					BorderForeground(lipgloss.Color("240"))
			}
		}
		opts = append(opts, style.Render(option))
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, opts...)
}
