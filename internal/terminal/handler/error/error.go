package error

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	hdltypes "github.com/termkit/gama/internal/terminal/handler/types"
	ts "github.com/termkit/gama/internal/terminal/style"
)

type ModelError struct {
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

func SetupModelError() ModelError {
	return ModelError{
		err:          nil,
		errorMessage: "",
	}
}

func (m *ModelError) SetError(err error) {
	m.err = err
}

func (m *ModelError) SetErrorMessage(errorMessage string) {
	m.errorMessage = errorMessage
}

func (m *ModelError) SetProgressMessage(message string) {
	m.messageType = MessageTypeProgress
	m.message = message
}

func (m *ModelError) SetSuccessMessage(message string) {
	m.messageType = MessageTypeSuccess
	m.message = message
}

func (m *ModelError) SetDefaultMessage(message string) {
	m.messageType = MessageTypeDefault
	m.message = message
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

func (m *ModelError) ViewError() string {
	doc := strings.Builder{}
	doc.WriteString(fmt.Sprintf("Error [%v]: %s", m.err, m.errorMessage))
	return doc.String()
}

func (m *ModelError) ViewMessage() string {
	doc := strings.Builder{}
	doc.WriteString(m.message)
	return doc.String()
}

func (m *ModelError) View() string {
	var windowStyle = lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder())

	doc := strings.Builder{}
	if m.HaveError() {
		windowStyle = ts.WindowStyleError.Width(*hdltypes.ScreenWidth)
		doc.WriteString(windowStyle.Render(m.ViewError()))
		return doc.String()
	}

	switch m.messageType {
	case MessageTypeDefault:
		windowStyle = ts.WindowStyleDefault.Width(*hdltypes.ScreenWidth)
	case MessageTypeProgress:
		windowStyle = ts.WindowStyleProgress.Width(*hdltypes.ScreenWidth)
	case MessageTypeSuccess:
		windowStyle = ts.WindowStyleSuccess.Width(*hdltypes.ScreenWidth)
	default:
		windowStyle = ts.WindowStyleDefault.Width(*hdltypes.ScreenWidth)
	}

	doc.WriteString(windowStyle.Render(m.ViewMessage()))

	return doc.String()
}
