package taboptions

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	hdlerror "github.com/termkit/gama/internal/terminal/handler/status"
)

type Options struct {
	Style lipgloss.Style

	modelError         *hdlerror.ModelStatus
	previousModelError hdlerror.ModelStatus
	modelLock          bool

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

func NewOptions(modelError *hdlerror.ModelStatus) *Options {
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
	optionsWithFunc[0] = func() {} // NO OPERATION

	return &Options{
		Style:           OptionsStyle,
		options:         initialOptions,
		optionsAction:   initialOptionsAction,
		optionsWithFunc: optionsWithFunc,
		status:          OptionWait,
		modelError:      modelError,
	}
}

func (o *Options) Init() tea.Cmd {
	return nil
}

func (o *Options) Update(msg tea.Msg) (*Options, tea.Cmd) {
	var cmd tea.Cmd

	if o.status == OptionWait || o.status == OptionNone {
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

func (o *Options) View() string {
	var style = o.Style.Foreground(lipgloss.Color("15"))

	var opts []string
	opts = append(opts, " ")

	for i, option := range o.optionsAction {
		switch o.status {
		case OptionWait:
			style = style.BorderForeground(lipgloss.Color("208"))
		case OptionNone:
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
	o.modelLock = false
	o.switchToPreviousError()
	o.optionsAction[0] = string(OptionIdle)
	o.cursor = 0
	o.isTabSelected = false
}

func (o *Options) updateCursor(cursor int) {
	if cursor < len(o.options) {
		o.cursor = cursor
		o.showAreYouSure()
		go o.resetOptionsWithOriginal()
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

func (o *Options) getOptionMessage() string {
	option := o.options[o.cursor]
	option = strings.TrimPrefix(option, fmt.Sprintf("%d) ", o.cursor))
	return option
}

func (o *Options) showAreYouSure() {
	if !o.modelLock {
		o.previousModelError = *o.modelError
		o.modelLock = true
	}
	o.modelError.Reset()
	o.modelError.SetProgressMessage(fmt.Sprintf("Are you sure you want to %s?", o.getOptionMessage()))
}

func (o *Options) switchToPreviousError() {
	if o.modelLock {
		return
	}
	*o.modelError = o.previousModelError
}

func (o *Options) executeOption() {
	go o.optionsWithFunc[o.cursor]()
	o.cursor = 0
	o.timer = -1
}
