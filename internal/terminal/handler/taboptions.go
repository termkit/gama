package handler

import (
	"fmt"
	"strings"
	"time"

	"github.com/termkit/skeleton"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type ModelTabOptions struct {
	skeleton *skeleton.Skeleton

	Style lipgloss.Style

	status         *ModelStatus
	previousStatus ModelStatus
	modelLock      bool

	optionStatus OptionStatus

	options       []string
	optionsAction []string

	optionsWithFunc map[int]func()

	timer int

	isTabSelected bool

	cursor int
}

type OptionStatus string

const (
	// StatusIdle is for when the options are ready to use
	StatusIdle OptionStatus = "Idle"

	// StatusWait is for when the options are not ready to use
	StatusWait OptionStatus = "Wait"

	// StatusNone is for when the options are not usable
	StatusNone OptionStatus = "None"
)

func (o OptionStatus) String() string {
	return string(o)
}

func NewOptions(sk *skeleton.Skeleton, modelStatus *ModelStatus) *ModelTabOptions {
	var b = lipgloss.RoundedBorder()
	b.Right = "├"
	b.Left = "┤"

	var OptionsStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("15")).
		Align(lipgloss.Center).Padding(0, 1, 0, 1).
		Border(b)

	var initialOptions = []string{
		StatusWait.String(),
	}
	var initialOptionsAction = []string{
		StatusWait.String(),
	}

	optionsWithFunc := make(map[int]func())
	optionsWithFunc[0] = func() {} // NO OPERATION

	return &ModelTabOptions{
		skeleton:        sk,
		Style:           OptionsStyle,
		options:         initialOptions,
		optionsAction:   initialOptionsAction,
		optionsWithFunc: optionsWithFunc,
		optionStatus:    StatusWait,
		status:          modelStatus,
	}
}

func (o *ModelTabOptions) Init() tea.Cmd {
	return nil
}

func (o *ModelTabOptions) Update(msg tea.Msg) (*ModelTabOptions, tea.Cmd) {
	var cmd tea.Cmd

	if o.optionStatus == StatusWait || o.optionStatus == StatusNone {
		return o, cmd
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "1":
			o.updateCursor(1)
		case "2":
			o.updateCursor(2)
		case "3":
			o.updateCursor(3)
		case "4":
			o.updateCursor(4)
		case "enter":
			o.executeOption()
		}
	}

	return o, cmd
}

func (o *ModelTabOptions) View() string {
	var style = o.Style.Foreground(lipgloss.Color("15"))

	var opts []string
	opts = append(opts, " ")

	for i, option := range o.optionsAction {
		switch o.optionStatus {
		case StatusWait:
			style = style.BorderForeground(lipgloss.Color("208"))
		case StatusNone:
			style = style.BorderForeground(lipgloss.Color("240"))
		default:
			isActive := i == o.cursor

			if isActive {
				style = style.BorderForeground(lipgloss.Color("150"))
			} else {
				style = style.BorderForeground(lipgloss.Color("240"))
			}
		}
		opts = append(opts, style.Render(option))
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, opts...)
}

func (o *ModelTabOptions) resetOptionsWithOriginal() {
	if o.isTabSelected {
		return
	}
	o.isTabSelected = true
	o.timer = 3
	for o.timer >= 0 {
		o.optionsAction[0] = fmt.Sprintf("> %ds", o.timer)
		time.Sleep(1 * time.Second)
		o.timer--
		o.skeleton.TriggerUpdate()
	}
	o.modelLock = false
	o.switchToPreviousError()
	o.optionsAction[0] = string(StatusIdle)
	o.cursor = 0
	o.isTabSelected = false
}

func (o *ModelTabOptions) updateCursor(cursor int) {
	if cursor < len(o.options) {
		o.cursor = cursor
		o.showAreYouSure()
		go o.resetOptionsWithOriginal()
	}
}

func (o *ModelTabOptions) SetStatus(status OptionStatus) {
	o.optionStatus = status
	o.options[0] = status.String()
	o.optionsAction[0] = status.String()
}

func (o *ModelTabOptions) AddOption(option string, action func()) {
	var optionWithNumber string
	var optionNumber = len(o.options)
	optionWithNumber = fmt.Sprintf("%d) %s", optionNumber, option)
	o.options = append(o.options, optionWithNumber)
	o.optionsAction = append(o.optionsAction, optionWithNumber)
	o.optionsWithFunc[optionNumber] = action
}

func (o *ModelTabOptions) getOptionMessage() string {
	option := o.options[o.cursor]
	option = strings.TrimPrefix(option, fmt.Sprintf("%d) ", o.cursor))
	return option
}

func (o *ModelTabOptions) showAreYouSure() {
	var yellowStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("4")).Blink(true)

	if !o.modelLock {
		o.previousStatus = *o.status
		o.modelLock = true
	}
	o.status.Reset()
	o.status.SetProgressMessage(fmt.Sprintf(
		"Are you sure you want to %s? %s",
		o.getOptionMessage(),
		yellowStyle.Render("[ Press Enter ]"),
	))

}

func (o *ModelTabOptions) switchToPreviousError() {
	if o.modelLock {
		return
	}
	*o.status = o.previousStatus
}

func (o *ModelTabOptions) executeOption() {
	go o.optionsWithFunc[o.cursor]()
	o.cursor = 0
	o.timer = -1
}
