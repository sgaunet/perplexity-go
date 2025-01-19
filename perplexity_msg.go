package perplexity

import "fmt"

// Message is a message object for the Perplexity API.
type Message struct {
	Role    string `json:"role" validate:"required,oneof=system user assistant"`
	Content string `json:"content"`
}

// Messages is an object that contains a list of messages for the Perplexity API.
type Messages struct {
	systemMessage string
	messages      []Message // A list of messages comprising the conversation so far.
}

// NewMessages returns a new Messages object.
func NewMessages(opts ...MessagesOption) Messages {
	m := Messages{}
	for _, opt := range opts {
		opt(&m)
	}
	return m
}

// MessagesOption is an option for the NewMessages function.
type MessagesOption func(*Messages)

// WithSystemMessage sets the system message for the Messages object.
func WithSystemMessage(content string) MessagesOption {
	return func(m *Messages) {
		m.systemMessage = content
	}
}

// AddUserMessage adds a user message to the Messages object.
func (m *Messages) AddUserMessage(content string) error {
	if len(m.messages) > 0 {
		// Previous message should be an assistant message.
		if m.messages[len(m.messages)-1].Role != "assistant" {
			return fmt.Errorf("previous message should be an assistant message")
		}
	}
	m.messages = append(m.messages, Message{
		Role:    "user",
		Content: content,
	})
	return nil
}

// AddAgentMessage adds an assistant message to the Messages object.
func (m *Messages) AddAgentMessage(content string) error {
	if len(m.messages) == 0 {
		// First message should be a user message.
		return fmt.Errorf("first message should be a user message")
	}
	// Previous message should be a user message.
	if m.messages[len(m.messages)-1].Role != "user" {
		return fmt.Errorf("previous message should be a user message")
	}
	m.messages = append(m.messages, Message{
		Role:    "assistant",
		Content: content,
	})
	return nil
}

func (m *Messages) GetMessages() []Message {
	var result []Message
	// system message is added in the first position
	if m.systemMessage != "" {
		result = append(result, Message{
			Role:    "system",
			Content: m.systemMessage,
		})
	}
	// user and assistant messages are added in the following positions
	result = append(result, m.messages...)
	return result
}

// GetSystemMessage returns the system message.
func (m *Messages) GetSystemMessage() string {
	return m.systemMessage
}
