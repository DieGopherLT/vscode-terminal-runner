package messages

// MessageType represents the type of message to be displayed.
type MessageType int

const (
	Error MessageType = iota
	Success
	Warning
	Info
)

// String returns the string representation of the MessageType.
func (t MessageType) String() string {
	switch t {
	case Error:
		return "error"
	case Success:
		return "success"
	case Warning:
		return "warning"
	case Info:
		return "info"
	default:
		return "unknown"
	}
}

// Message represents a single message with its type and content.
type Message struct {
	Type    MessageType
	Content string
}