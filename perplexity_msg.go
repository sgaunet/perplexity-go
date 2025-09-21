package perplexity

import "errors"

// ContentType represents the type of content in a multimodal message.
type ContentType string

const (
	// ContentTypeText represents text content in a multimodal message.
	ContentTypeText ContentType = "text"
	// ContentTypeImageURL represents image URL content in a multimodal message.
	ContentTypeImageURL ContentType = "image_url"
)

// Content represents either text or image content in a multimodal message.
type Content struct {
	Type     ContentType `json:"type" validate:"required,oneof=text image_url"`
	Text     *string     `json:"text,omitempty" validate:"required_if=Type text"`
	ImageURL *ImageURL   `json:"image_url,omitempty" validate:"required_if=Type image_url"`
}

// ImageURL represents image content with URL or base64 data URI.
type ImageURL struct {
	URL string `json:"url" validate:"required"`
}

// MultimodalMessage supports both text and image content according to Perplexity API specification.
type MultimodalMessage struct {
	Role    string    `json:"role" validate:"required,oneof=system user assistant"`
	Content []Content `json:"content" validate:"required,min=1,dive"`
}

// NewTextContent creates a text content object for multimodal messages.
func NewTextContent(text string) Content {
	return Content{
		Type: ContentTypeText,
		Text: &text,
	}
}

// NewImageURLContent creates an image URL content object for multimodal messages.
func NewImageURLContent(url string) Content {
	return Content{
		Type: ContentTypeImageURL,
		ImageURL: &ImageURL{
			URL: url,
		},
	}
}

// NewImageFileContent creates image content from a file path by encoding it to base64 data URI.
func NewImageFileContent(filepath string) (Content, error) {
	processor := NewImageProcessor()
	dataURI, err := processor.EncodeImageFromFile(filepath)
	if err != nil {
		return Content{}, err
	}

	return NewImageURLContent(dataURI), nil
}

// Error definitions.
var (
	// ErrPrevMessageShouldBeAssistant is returned when the previous message is not from the assistant.
	ErrPrevMessageShouldBeAssistant = errors.New("previous message should be an assistant message")
	// ErrFirstMessageShouldBeUser is returned when the first message is not from the user.
	ErrFirstMessageShouldBeUser = errors.New("first message should be a user message")
	// ErrPrevMessageShouldBeUser is returned when the previous message is not from the user.
	ErrPrevMessageShouldBeUser = errors.New("previous message should be a user message")
	// ErrMultimodalContentEmpty is returned when multimodal message content is empty.
	ErrMultimodalContentEmpty = errors.New("multimodal message content cannot be empty")
	// ErrMultimodalInvalidContentType is returned when multimodal content has invalid type.
	ErrMultimodalInvalidContentType = errors.New("invalid content type in multimodal message")
)

// Message is a message object for the Perplexity API.
type Message struct {
	Role    string `json:"role" validate:"required,oneof=system user assistant"`
	Content string `json:"content"`
}

// Messages is an object that contains a list of messages for the Perplexity API.
type Messages struct {
	systemMessage       string
	messages            []Message            // A list of messages comprising the conversation so far.
	multimodalMessages  []MultimodalMessage  // A list of multimodal messages (when using images).
	useMultimodal       bool                 // Flag to indicate if multimodal messages are being used.
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
			return ErrPrevMessageShouldBeAssistant
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
		return ErrFirstMessageShouldBeUser
	}
	// Previous message should be a user message.
	if m.messages[len(m.messages)-1].Role != "user" {
		return ErrPrevMessageShouldBeUser
	}
	m.messages = append(m.messages, Message{
		Role:    "assistant",
		Content: content,
	})
	return nil
}

// GetMessages returns all messages including the system message (if any) as a slice of Message.
// The system message is always included as the first message when present.
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

// AddMultimodalUserMessage adds a user message with mixed content (text + images).
func (m *Messages) AddMultimodalUserMessage(contents []Content) error {
	if len(contents) == 0 {
		return ErrMultimodalContentEmpty
	}

	// Validate content types
	for _, content := range contents {
		if content.Type != ContentTypeText && content.Type != ContentTypeImageURL {
			return ErrMultimodalInvalidContentType
		}
	}

	if len(m.multimodalMessages) > 0 {
		// Previous message should be an assistant message.
		if m.multimodalMessages[len(m.multimodalMessages)-1].Role != "assistant" {
			return ErrPrevMessageShouldBeAssistant
		}
	}

	m.multimodalMessages = append(m.multimodalMessages, MultimodalMessage{
		Role:    "user",
		Content: contents,
	})
	m.useMultimodal = true
	return nil
}

// AddUserMessageWithImage adds a user message with text and a single image URL.
func (m *Messages) AddUserMessageWithImage(text, imageURL string) error {
	// Validate image URL
	processor := NewImageProcessor()
	if err := processor.ValidateImageURL(imageURL); err != nil {
		return err
	}

	contents := []Content{
		NewTextContent(text),
		NewImageURLContent(imageURL),
	}

	return m.AddMultimodalUserMessage(contents)
}

// AddUserMessageWithImageFile adds a user message with text and image from file path.
func (m *Messages) AddUserMessageWithImageFile(text, filepath string) error {
	imageContent, err := NewImageFileContent(filepath)
	if err != nil {
		return err
	}

	contents := []Content{
		NewTextContent(text),
		imageContent,
	}

	return m.AddMultimodalUserMessage(contents)
}

// AddMultimodalAgentMessage adds an assistant message with mixed content.
func (m *Messages) AddMultimodalAgentMessage(contents []Content) error {
	if len(contents) == 0 {
		return ErrMultimodalContentEmpty
	}

	if len(m.multimodalMessages) == 0 {
		// First message should be a user message.
		return ErrFirstMessageShouldBeUser
	}

	// Previous message should be a user message.
	if m.multimodalMessages[len(m.multimodalMessages)-1].Role != "user" {
		return ErrPrevMessageShouldBeUser
	}

	m.multimodalMessages = append(m.multimodalMessages, MultimodalMessage{
		Role:    "assistant",
		Content: contents,
	})
	m.useMultimodal = true
	return nil
}

// GetMultimodalMessages returns all messages in multimodal format including the system message.
func (m *Messages) GetMultimodalMessages() []MultimodalMessage {
	var result []MultimodalMessage

	// Add system message as first message if present
	if m.systemMessage != "" {
		result = append(result, MultimodalMessage{
			Role: "system",
			Content: []Content{
				NewTextContent(m.systemMessage),
			},
		})
	}

	// Add multimodal messages
	result = append(result, m.multimodalMessages...)
	return result
}

// IsMultimodal returns true if the Messages object contains multimodal messages.
func (m *Messages) IsMultimodal() bool {
	return m.useMultimodal
}

// HasImages returns true if any of the multimodal messages contain images.
func (m *Messages) HasImages() bool {
	if !m.useMultimodal {
		return false
	}

	for _, msg := range m.multimodalMessages {
		for _, content := range msg.Content {
			if content.Type == ContentTypeImageURL {
				return true
			}
		}
	}
	return false
}
