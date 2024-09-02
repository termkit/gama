package error

import (
	"fmt"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	ts "github.com/termkit/gama/internal/terminal/style"
	"github.com/termkit/skeleton"
	"strings"
)

type ModelError struct {
	skeleton *skeleton.Skeleton
	// err is hold the error
	err error

	// errorMessage is hold the error message
	errorMessage string

	// message is hold the message, if there is no error
	message string

	// messageType is hold the message type
	messageType MessageType

	// spinner is hold the spinner
	spinner spinner.Model

	updateChan     chan UpdateSelf
	disableSpinner bool
}

type UpdateSelf struct {
	Message    string
	InProgress bool
}

type MessageType string

const (
	// MessageTypeDefault is the message type for default
	MessageTypeDefault MessageType = "default"

	// MessageTypeProgress is the message type for progress
	MessageTypeProgress MessageType = "progress"

	// MessageTypeSuccess is the message type for success
	MessageTypeSuccess MessageType = "success"
)

func SetupModelError(skeleton *skeleton.Skeleton) ModelError {
	s := spinner.New(spinner.WithSpinner(spinner.Dot))
	return ModelError{
		skeleton:       skeleton,
		spinner:        s,
		err:            nil,
		errorMessage:   "",
		updateChan:     make(chan UpdateSelf),
		disableSpinner: false,
	}
}

func (m *ModelError) Init() tea.Cmd {
	return tea.Batch(m.SelfUpdater())
}

func (m *ModelError) Update(msg tea.Msg) (*ModelError, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case spinner.TickMsg:
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)

		cmds = append(cmds, m.SelfUpdater())
	case UpdateSelf:
		if msg.InProgress {
			m.spinner, cmd = m.spinner.Update(m.spinner.Tick())
			cmds = append(cmds, cmd)
		}

		cmds = append(cmds, m.SelfUpdater())
	}

	return m, tea.Batch(cmds...)
}

func (m *ModelError) View() string {
	var windowStyle = lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder())
	width := m.skeleton.GetTerminalWidth() - 4
	doc := strings.Builder{}

	var s string
	if !m.disableSpinner {
		s = m.spinner.View()
	}

	if m.HaveError() {
		windowStyle = ts.WindowStyleError.Width(width)
		doc.WriteString(windowStyle.Render(m.viewError()))
		return lipgloss.JoinHorizontal(lipgloss.Top, doc.String())
	}

	switch m.messageType {
	case MessageTypeDefault:
		windowStyle = ts.WindowStyleDefault.Width(width)
		s = ""
	case MessageTypeProgress:
		windowStyle = ts.WindowStyleProgress.Width(width)
	case MessageTypeSuccess:
		windowStyle = ts.WindowStyleSuccess.Width(width)
		s = ""
	default:
		windowStyle = ts.WindowStyleDefault.Width(width)
	}

	doc.WriteString(windowStyle.Render(lipgloss.JoinHorizontal(lipgloss.Top, m.viewMessage(), " ", s)))
	return doc.String()
}

func (m *ModelError) SelfUpdater() tea.Cmd {
	return func() tea.Msg {
		return <-m.updateChan
	}
}

func (m *ModelError) EnableSpinner() {
	m.disableSpinner = false
	//m.updateChan <- UpdateSelf{Message: m.message, InProgress: true}
}

func (m *ModelError) DisableSpinner() {
	m.disableSpinner = true
	//m.updateChan <- UpdateSelf{Message: m.message, InProgress: false}
}

func (m *ModelError) SetError(err error) {
	m.err = err
}

func (m *ModelError) SetErrorMessage(message string) {
	m.errorMessage = message
	go func() {
		m.updateChan <- UpdateSelf{Message: message, InProgress: true}
	}()
}

func (m *ModelError) SetProgressMessage(message string) {
	m.messageType = MessageTypeProgress
	m.message = message
	go func() {
		m.updateChan <- UpdateSelf{Message: message, InProgress: true}
	}()
}

func (m *ModelError) SetSuccessMessage(message string) {
	m.messageType = MessageTypeSuccess
	m.message = message
	go func() {
		m.updateChan <- UpdateSelf{Message: message, InProgress: true}
	}()
}

func (m *ModelError) SetDefaultMessage(message string) {
	m.messageType = MessageTypeDefault
	m.message = message
	go func() {
		m.updateChan <- UpdateSelf{Message: message, InProgress: true}
	}()
}

func (m *ModelError) GetError() error {
	return m.err
}

func (m *ModelError) GetErrorMessage() string {
	return m.errorMessage
}

func (m *ModelError) GetMessage() string {
	return m.message
}

func (m *ModelError) ResetError() {
	m.err = nil
	m.errorMessage = ""
}

func (m *ModelError) ResetMessage() {
	m.message = ""
}

func (m *ModelError) Reset() {
	m.ResetError()
	m.ResetMessage()
}

func (m *ModelError) HaveError() bool {
	return m.err != nil
}

func (m *ModelError) viewError() string {
	doc := strings.Builder{}
	doc.WriteString(fmt.Sprintf("Error [%v]: %s", m.err, m.errorMessage))
	return doc.String()
}

func (m *ModelError) viewMessage() string {
	doc := strings.Builder{}
	doc.WriteString(m.message)
	return doc.String()
}
