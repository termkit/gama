package handler

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/termkit/skeleton"
)

type ModelStatus struct {
	skeleton *skeleton.Skeleton

	// err is hold the error
	err error

	// errorMessage is hold the error message
	errorMessage string

	// message is hold the message, if there is no error
	message string

	// messageType is hold the message type
	messageType MessageType
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

func SetupModelStatus(skeleton *skeleton.Skeleton) *ModelStatus {
	return &ModelStatus{
		skeleton:     skeleton,
		err:          nil,
		errorMessage: "",
	}
}

func (m *ModelStatus) View() string {
	var windowStyle = lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder())
	width := m.skeleton.GetTerminalWidth() - 4
	doc := strings.Builder{}

	if m.HaveError() {
		windowStyle = WindowStyleError.Width(width)
		doc.WriteString(windowStyle.Render(m.viewError()))
		return lipgloss.JoinHorizontal(lipgloss.Top, doc.String())
	}

	switch m.messageType {
	case MessageTypeDefault:
		windowStyle = WindowStyleDefault.Width(width)
	case MessageTypeProgress:
		windowStyle = WindowStyleProgress.Width(width)
	case MessageTypeSuccess:
		windowStyle = WindowStyleSuccess.Width(width)
	default:
		windowStyle = WindowStyleDefault.Width(width)
	}

	doc.WriteString(windowStyle.Render(m.viewMessage()))
	return doc.String()
}

func (m *ModelStatus) SetError(err error) {
	m.err = err
}

func (m *ModelStatus) SetErrorMessage(message string) {
	m.errorMessage = message
}

func (m *ModelStatus) SetProgressMessage(message string) {
	m.messageType = MessageTypeProgress
	m.message = message
}

func (m *ModelStatus) SetSuccessMessage(message string) {
	m.messageType = MessageTypeSuccess
	m.message = message
}

func (m *ModelStatus) SetDefaultMessage(message string) {
	m.messageType = MessageTypeDefault
	m.message = message
}

func (m *ModelStatus) GetError() error {
	return m.err
}

func (m *ModelStatus) GetErrorMessage() string {
	return m.errorMessage
}

func (m *ModelStatus) GetMessage() string {
	return m.message
}

func (m *ModelStatus) ResetError() {
	m.err = nil
	m.errorMessage = ""
}

func (m *ModelStatus) ResetMessage() {
	m.message = ""
}

func (m *ModelStatus) Reset() {
	m.ResetError()
	m.ResetMessage()
}

func (m *ModelStatus) HaveError() bool {
	return m.err != nil
}

func (m *ModelStatus) viewError() string {
	doc := strings.Builder{}
	doc.WriteString(fmt.Sprintf("Error [%v]: %s", m.err, m.errorMessage))
	return doc.String()
}

func (m *ModelStatus) viewMessage() string {
	doc := strings.Builder{}
	doc.WriteString(m.message)
	return doc.String()
}
