package perplexity

import (
	"errors"
	"time"
)

// AsyncJobRequestOption is a functional option for configuring AsyncJobRequest.
type AsyncJobRequestOption func(*AsyncJobRequest)

// DefaultAsyncJobRequest creates a new AsyncJobRequest with default values.
func DefaultAsyncJobRequest() *AsyncJobRequest {
	return &AsyncJobRequest{
		Model:              ModelSonarDeepResearch, // Must be sonar-deep-research for async
		MaxTokens:          DefaultMaxTokens,       // 4000
		Temperature:        DefaultTemperature,     // 0.2
		TopP:               DefaultTopP,            // 0.9
		TopK:               DefaultTopK,            // 0
		PresencePenalty:    DefaultPresencePenalty, // 0.0
		FrequencyPenalty:   DefaultFrequencyPenalty, // 1.0
		SearchMode:         DefaultSearchMode,      // "web"
		SearchDomainFilter: nil,                    // null in JSON
		Stream:             false,                  // async doesn't support streaming
	}
}

// NewAsyncJobRequest creates a new AsyncJobRequest with functional options.
func NewAsyncJobRequest(opts ...AsyncJobRequestOption) *AsyncJobRequest {
	req := DefaultAsyncJobRequest()

	for _, opt := range opts {
		opt(req)
	}

	return req
}

// WithAsyncMessages sets the messages for the async request.
func WithAsyncMessages(messages []Message) AsyncJobRequestOption {
	return func(req *AsyncJobRequest) {
		req.Messages = messages
		req.MultimodalMessages = nil // Clear multimodal messages
	}
}

// WithAsyncMultimodalMessages sets the multimodal messages for the async request.
func WithAsyncMultimodalMessages(messages []MultimodalMessage) AsyncJobRequestOption {
	return func(req *AsyncJobRequest) {
		req.MultimodalMessages = messages
		req.Messages = nil // Clear regular messages
	}
}

// WithAsyncMessagesObject sets messages from a Messages object.
func WithAsyncMessagesObject(messages Messages) AsyncJobRequestOption {
	return func(req *AsyncJobRequest) {
		if messages.IsMultimodal() {
			req.MultimodalMessages = messages.GetMultimodalMessages()
			req.Messages = nil
		} else {
			req.Messages = messages.GetMessages()
			req.MultimodalMessages = nil
		}
	}
}

// WithAsyncReasoningEffort sets the reasoning effort level.
func WithAsyncReasoningEffort(effort string) AsyncJobRequestOption {
	return func(req *AsyncJobRequest) {
		req.ReasoningEffort = effort
	}
}

// WithAsyncMaxTokens sets the maximum number of tokens.
func WithAsyncMaxTokens(maxTokens int) AsyncJobRequestOption {
	return func(req *AsyncJobRequest) {
		req.MaxTokens = maxTokens
	}
}

// WithAsyncTemperature sets the temperature parameter.
func WithAsyncTemperature(temperature float64) AsyncJobRequestOption {
	return func(req *AsyncJobRequest) {
		req.Temperature = temperature
	}
}

// WithAsyncTopP sets the top-p parameter.
func WithAsyncTopP(topP float64) AsyncJobRequestOption {
	return func(req *AsyncJobRequest) {
		req.TopP = topP
	}
}

// WithAsyncTopK sets the top-k parameter.
func WithAsyncTopK(topK int) AsyncJobRequestOption {
	return func(req *AsyncJobRequest) {
		req.TopK = topK
	}
}

// WithAsyncPresencePenalty sets the presence penalty.
func WithAsyncPresencePenalty(penalty float64) AsyncJobRequestOption {
	return func(req *AsyncJobRequest) {
		req.PresencePenalty = penalty
	}
}

// WithAsyncFrequencyPenalty sets the frequency penalty.
func WithAsyncFrequencyPenalty(penalty float64) AsyncJobRequestOption {
	return func(req *AsyncJobRequest) {
		req.FrequencyPenalty = penalty
	}
}

// WithAsyncSearchDomainFilter sets the search domain filter.
func WithAsyncSearchDomainFilter(domains []string) AsyncJobRequestOption {
	return func(req *AsyncJobRequest) {
		req.SearchDomainFilter = domains
	}
}

// WithAsyncSearchMode sets the search mode.
func WithAsyncSearchMode(mode string) AsyncJobRequestOption {
	return func(req *AsyncJobRequest) {
		req.SearchMode = mode
	}
}

// WithAsyncSearchDomain sets the search domain.
func WithAsyncSearchDomain(domain string) AsyncJobRequestOption {
	return func(req *AsyncJobRequest) {
		req.SearchDomain = domain
	}
}

// WithAsyncImageDomainFilter sets the image domain filter.
func WithAsyncImageDomainFilter(domains []string) AsyncJobRequestOption {
	return func(req *AsyncJobRequest) {
		req.ImageDomainFilter = domains
	}
}

// WithAsyncImageFormatFilter sets the image format filter.
func WithAsyncImageFormatFilter(formats []string) AsyncJobRequestOption {
	return func(req *AsyncJobRequest) {
		req.ImageFormatFilter = formats
	}
}

// WithAsyncPublishedAfter sets the published after filter.
func WithAsyncPublishedAfter(date string) AsyncJobRequestOption {
	return func(req *AsyncJobRequest) {
		req.PublishedAfter = date
	}
}

// WithAsyncPublishedBefore sets the published before filter.
func WithAsyncPublishedBefore(date string) AsyncJobRequestOption {
	return func(req *AsyncJobRequest) {
		req.PublishedBefore = date
	}
}

// WithAsyncLastUpdatedAfter sets the last updated after filter.
func WithAsyncLastUpdatedAfter(date string) AsyncJobRequestOption {
	return func(req *AsyncJobRequest) {
		req.LastUpdatedAfterFilter = date
	}
}

// WithAsyncLastUpdatedBefore sets the last updated before filter.
func WithAsyncLastUpdatedBefore(date string) AsyncJobRequestOption {
	return func(req *AsyncJobRequest) {
		req.LastUpdatedBeforeFilter = date
	}
}

// WithAsyncWebSearchOptions sets the web search options.
func WithAsyncWebSearchOptions(options *WebSearchOptions) AsyncJobRequestOption {
	return func(req *AsyncJobRequest) {
		req.WebSearchOptions = options
	}
}

// AsyncJobRequestValidator provides validation for AsyncJobRequest.
type AsyncJobRequestValidator struct{}

// NewAsyncJobRequestValidator creates a new request validator.
func NewAsyncJobRequestValidator() *AsyncJobRequestValidator {
	return &AsyncJobRequestValidator{}
}

// Validate validates an AsyncJobRequest.
func (v *AsyncJobRequestValidator) Validate(req *AsyncJobRequest) error {
	if req == nil {
		return errors.New("request cannot be nil")
	}

	// Validate model
	if req.Model != ModelSonarDeepResearch {
		return ErrAsyncModelRequired
	}

	// Validate messages
	if len(req.Messages) == 0 && len(req.MultimodalMessages) == 0 {
		return errors.New("at least one message is required")
	}

	// Cannot have both regular and multimodal messages
	if len(req.Messages) > 0 && len(req.MultimodalMessages) > 0 {
		return errors.New("cannot specify both messages and multimodal messages")
	}

	// Validate reasoning effort
	if req.ReasoningEffort != "" {
		validEfforts := map[string]bool{
			"low":    true,
			"medium": true,
			"high":   true,
		}
		if !validEfforts[req.ReasoningEffort] {
			return errors.New("reasoning effort must be 'low', 'medium', or 'high'")
		}
	}

	// Validate temperature
	if req.Temperature < 0 || req.Temperature >= 2 {
		return errors.New("temperature must be between 0 and 2")
	}

	// Validate top-p
	if req.TopP != 0 && (req.TopP <= 0 || req.TopP >= 1) {
		return errors.New("top_p must be between 0 and 1")
	}

	// Validate top-k
	if req.TopK < 0 || req.TopK > 2048 {
		return errors.New("top_k must be between 0 and 2048")
	}

	// Validate presence penalty
	if req.PresencePenalty < -2.0 || req.PresencePenalty > 2.0 {
		return errors.New("presence penalty must be between -2.0 and 2.0")
	}

	// Validate frequency penalty
	if req.FrequencyPenalty < -2.0 || req.FrequencyPenalty > 2.0 {
		return errors.New("frequency penalty must be between -2.0 and 2.0")
	}

	// Validate search domain filter
	if len(req.SearchDomainFilter) > 10 {
		return errors.New("search domain filter cannot exceed 10 domains")
	}

	// Validate search mode
	if req.SearchMode != "" {
		validModes := map[string]bool{
			"academic": true,
			"web":      true,
		}
		if !validModes[req.SearchMode] {
			return errors.New("search mode must be 'academic' or 'web'")
		}
	}

	// Validate search domain
	if req.SearchDomain != "" {
		validDomains := map[string]bool{
			"sec": true,
		}
		if !validDomains[req.SearchDomain] {
			return errors.New("search domain must be 'sec'")
		}
	}

	// Validate image domain filter
	if len(req.ImageDomainFilter) > 10 {
		return errors.New("image domain filter cannot exceed 10 domains")
	}

	// Validate image format filter
	if len(req.ImageFormatFilter) > 10 {
		return errors.New("image format filter cannot exceed 10 formats")
	}

	// Validate date formats (basic validation)
	if req.PublishedAfter != "" {
		if err := v.validateDateString(req.PublishedAfter); err != nil {
			return errors.New("invalid published_after date format")
		}
	}

	if req.PublishedBefore != "" {
		if err := v.validateDateString(req.PublishedBefore); err != nil {
			return errors.New("invalid published_before date format")
		}
	}

	if req.LastUpdatedAfterFilter != "" {
		if err := v.validateDateString(req.LastUpdatedAfterFilter); err != nil {
			return errors.New("invalid last_updated_after_filter date format")
		}
	}

	if req.LastUpdatedBeforeFilter != "" {
		if err := v.validateDateString(req.LastUpdatedBeforeFilter); err != nil {
			return errors.New("invalid last_updated_before_filter date format")
		}
	}

	return nil
}

// validateDateString performs basic date string validation.
func (v *AsyncJobRequestValidator) validateDateString(dateStr string) error {
	// Try to parse common date formats
	formats := []string{
		"2006-01-02",
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05-07:00",
		time.RFC3339,
	}

	for _, format := range formats {
		if _, err := time.Parse(format, dateStr); err == nil {
			return nil
		}
	}

	return errors.New("unsupported date format")
}

// ValidateAsyncJobRequest is a convenience function for validating async job requests.
func ValidateAsyncJobRequest(req *AsyncJobRequest) error {
	validator := NewAsyncJobRequestValidator()
	return validator.Validate(req)
}