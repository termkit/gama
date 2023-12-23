package error

import (
	"fmt"
	"strings"
)

type ModelError struct {
	// err is hold the error
	err error

	// errorMessage is hold the error message
	errorMessage string

	// message is hold the message, if there is no error
	message string
}

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

func (m *ModelError) SetMessage(message string) {
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

func (m *ModelError) IsError() bool {
	return m.err != nil
}

func (m *ModelError) ViewError() string {
	doc := strings.Builder{}
	doc.WriteString(fmt.Sprintf("Error [%v]: %s", m.err, m.errorMessage))
	return doc.String()
}

func (m *ModelError) ViewMessage() string {
	doc := strings.Builder{}
	doc.WriteString(fmt.Sprintf("Operation: %s", m.message))
	return doc.String()
}

func (m *ModelError) View() string {
	doc := strings.Builder{}
	if m.IsError() {
		doc.WriteString(m.ViewError())
	} else {
		doc.WriteString(m.ViewMessage())
	}
	return doc.String()
}
