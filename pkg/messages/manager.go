package messages

import (
	"strings"

	"github.com/DieGopherLT/vscode-terminal-runner/pkg/styles"
	"github.com/samber/lo"
)

// MessageManager handles the collection and display of messages.
type MessageManager struct {
	showErrors   bool
	showMessages bool
	errors       []string
	messages     []Message
}

// NewManager creates a new MessageManager instance.
func NewManager() *MessageManager {
	return &MessageManager{
		showErrors:   false,
		showMessages: false,
		errors:       make([]string, 0),
		messages:     make([]Message, 0),
	}
}

// AddError adds an error message to be displayed.
func (m *MessageManager) AddError(message string) {
	m.errors = append(m.errors, message)
	m.showErrors = true
}

// AddSuccess adds a success message to be displayed.
func (m *MessageManager) AddSuccess(message string) {
	m.messages = append(m.messages, Message{
		Type:    Success,
		Content: message,
	})
	m.showMessages = true
}

// AddWarning adds a warning message to be displayed.
func (m *MessageManager) AddWarning(message string) {
	m.messages = append(m.messages, Message{
		Type:    Warning,
		Content: message,
	})
	m.showMessages = true
}

// AddInfo adds an info message to be displayed.
func (m *MessageManager) AddInfo(message string) {
	m.messages = append(m.messages, Message{
		Type:    Info,
		Content: message,
	})
	m.showMessages = true
}

// Clear removes all messages and resets display flags.
func (m *MessageManager) Clear() {
	m.errors = make([]string, 0)
	m.messages = make([]Message, 0)
	m.showErrors = false
	m.showMessages = false
}

// HasMessages returns true if there are any messages to display.
func (m *MessageManager) HasMessages() bool {
	return m.showErrors || m.showMessages
}

// HasErrors returns true if there are error messages to display.
func (m *MessageManager) HasErrors() bool {
	return m.showErrors && len(m.errors) > 0
}

// HasNonErrorMessages returns true if there are non-error messages to display.
func (m *MessageManager) HasNonErrorMessages() bool {
	return m.showMessages && len(m.messages) > 0
}

// GetErrors returns all error messages.
func (m *MessageManager) GetErrors() []string {
	return m.errors
}

// GetMessages returns all non-error messages.
func (m *MessageManager) GetMessages() []Message {
	return m.messages
}

// Render returns the styled string representation of all messages.
// This method will be implemented after we create the message styles.
func (m *MessageManager) Render() string {
	if !m.HasMessages() {
		return ""
	}

	var sections []string

	// Render error messages
	if m.HasErrors() {
		errorMessages := lo.Map(m.errors, func(err string, _ int) string {
			return renderMessage(Error, err)
		})
		sections = append(sections, strings.Join(errorMessages, "\n"))
	}

	// Render other messages
	if m.HasNonErrorMessages() {
		otherMessages := lo.Map(m.messages, func(msg Message, _ int) string {
			return renderMessage(msg.Type, msg.Content)
		})
		sections = append(sections, strings.Join(otherMessages, "\n"))
	}

	return strings.Join(sections, "\n")
}

// renderMessage returns a styled message based on its type using proper styles.
func renderMessage(msgType MessageType, content string) string {
	switch msgType {
	case Error:
		return styles.RenderErrorMessage(content)
	case Success:
		return styles.RenderSuccessMessage(content)
	case Warning:
		return styles.RenderWarningMessage(content)
	case Info:
		return styles.RenderInfoMessage(content)
	default:
		return content
	}
}