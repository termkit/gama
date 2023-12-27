package taboptions

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
	// OptionIdle is for when the options are ready to use
	OptionIdle OptionStatus = "Idle"

	// OptionWait is for when the options are not ready to use
	OptionWait OptionStatus = "Wait"

	// OptionNone is for when the options are not usable
	OptionNone OptionStatus = "None"
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

	var initialOptions = []string{
		OptionWait.String(),
	}
	var initialOptionsAction = []string{
		OptionWait.String(),
	}

	optionsWithFunc := make(map[int]func())
	optionsWithFunc[0] = func() {}

	return &Options{
		Style:           OptionsStyle,
		options:         initialOptions,
		optionsAction:   initialOptionsAction,
		optionsWithFunc: optionsWithFunc,
		status:          OptionWait,
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
	o.optionsAction[0] = string(OptionIdle)
	o.cursor = 0
	o.isTabSelected = false
}

func (o *Options) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	if o.status == OptionWait || o.status == OptionNone {
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
		if o.status == OptionWait {
			// orange
			style = o.Style.Copy().
				Foreground(lipgloss.Color("15")).
				BorderForeground(lipgloss.Color("208"))
		} else if o.status == OptionNone {
			// gray
			style = o.Style.Copy().
				Foreground(lipgloss.Color("15")).
				BorderForeground(lipgloss.Color("240"))
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
