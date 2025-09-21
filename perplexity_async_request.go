package perplexity

import (
	"errors"
	"time"
)

// Constants for async request validation.
const (
	// MaxDomainFilterCount is the maximum number of domains allowed in filters.
	MaxDomainFilterCount = 10
	// MaxImageFormatFilterCount is the maximum number of image formats allowed.
	MaxImageFormatFilterCount = 10
)

// Static validation errors.
var (
	// ErrRequestNil is returned when request is nil.
	ErrRequestNil = errors.New("request cannot be nil")
	// ErrMessagesRequired is returned when no messages are provided.
	ErrMessagesRequired = errors.New("at least one message is required")
	// ErrBothMessageTypes is returned when both regular and multimodal messages are provided.
	ErrBothMessageTypes = errors.New("cannot specify both messages and multimodal messages")
	// ErrInvalidReasoningEffort is returned when reasoning effort is invalid.
	ErrInvalidReasoningEffort = errors.New("reasoning effort must be 'low', 'medium', or 'high'")
	// ErrInvalidTemperature is returned when temperature is out of range.
	ErrInvalidTemperature = errors.New("temperature must be between 0 and 2")
	// ErrInvalidTopP is returned when top_p is out of range.
	ErrInvalidTopP = errors.New("top_p must be between 0 and 1")
	// ErrInvalidTopK is returned when top_k is out of range.
	ErrInvalidTopK = errors.New("top_k must be between 0 and 2048")
	// ErrInvalidPresencePenalty is returned when presence penalty is out of range.
	ErrInvalidPresencePenalty = errors.New("presence penalty must be between -2.0 and 2.0")
	// ErrInvalidFrequencyPenalty is returned when frequency penalty is out of range.
	ErrInvalidFrequencyPenalty = errors.New("frequency penalty must be between -2.0 and 2.0")
	// ErrTooManySearchDomains is returned when too many search domains are provided.
	ErrTooManySearchDomains = errors.New("search domain filter cannot exceed 10 domains")
	// ErrInvalidSearchMode is returned when search mode is invalid.
	ErrInvalidSearchMode = errors.New("search mode must be 'academic' or 'web'")
	// ErrInvalidSearchDomain is returned when search domain is invalid.
	ErrInvalidSearchDomain = errors.New("search domain must be 'sec'")
	// ErrTooManyImageDomains is returned when too many image domains are provided.
	ErrTooManyImageDomains = errors.New("image domain filter cannot exceed 10 domains")
	// ErrTooManyImageFormats is returned when too many image formats are provided.
	ErrTooManyImageFormats = errors.New("image format filter cannot exceed 10 formats")
	// ErrInvalidPublishedAfter is returned when published_after date is invalid.
	ErrInvalidPublishedAfter = errors.New("invalid published_after date format")
	// ErrInvalidPublishedBefore is returned when published_before date is invalid.
	ErrInvalidPublishedBefore = errors.New("invalid published_before date format")
	// ErrInvalidLastUpdatedAfter is returned when last_updated_after_filter date is invalid.
	ErrInvalidLastUpdatedAfter = errors.New("invalid last_updated_after_filter date format")
	// ErrInvalidLastUpdatedBefore is returned when last_updated_before_filter date is invalid.
	ErrInvalidLastUpdatedBefore = errors.New("invalid last_updated_before_filter date format")
	// ErrUnsupportedDateFormat is returned when date format is unsupported.
	ErrUnsupportedDateFormat = errors.New("unsupported date format")
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
	if err := v.validateBasicFields(req); err != nil {
		return err
	}
	if err := v.validateParameters(req); err != nil {
		return err
	}
	if err := v.validateSearchOptions(req); err != nil {
		return err
	}
	if err := v.validateImageOptions(req); err != nil {
		return err
	}
	return v.validateDateFields(req)
}

// validateBasicFields validates the basic required fields.
func (v *AsyncJobRequestValidator) validateBasicFields(req *AsyncJobRequest) error {
	if req == nil {
		return ErrRequestNil
	}

	// Validate model
	if req.Model != ModelSonarDeepResearch {
		return ErrAsyncModelRequired
	}

	// Validate messages
	if len(req.Messages) == 0 && len(req.MultimodalMessages) == 0 {
		return ErrMessagesRequired
	}

	// Cannot have both regular and multimodal messages
	if len(req.Messages) > 0 && len(req.MultimodalMessages) > 0 {
		return ErrBothMessageTypes
	}

	return nil
}

// validateParameters validates the generation parameters.
func (v *AsyncJobRequestValidator) validateParameters(req *AsyncJobRequest) error {
	if err := v.validateReasoningEffort(req); err != nil {
		return err
	}
	if err := v.validateGenerationParams(req); err != nil {
		return err
	}
	return v.validatePenaltyParams(req)
}

// validateReasoningEffort validates the reasoning effort parameter.
func (v *AsyncJobRequestValidator) validateReasoningEffort(req *AsyncJobRequest) error {
	if req.ReasoningEffort != "" {
		validEfforts := map[string]bool{
			"low":    true,
			"medium": true,
			"high":   true,
		}
		if !validEfforts[req.ReasoningEffort] {
			return ErrInvalidReasoningEffort
		}
	}
	return nil
}

// validateGenerationParams validates temperature, top-p, and top-k parameters.
func (v *AsyncJobRequestValidator) validateGenerationParams(req *AsyncJobRequest) error {
	// Validate temperature
	if req.Temperature < 0 || req.Temperature >= 2 {
		return ErrInvalidTemperature
	}

	// Validate top-p
	if req.TopP != 0 && (req.TopP <= 0 || req.TopP >= 1) {
		return ErrInvalidTopP
	}

	// Validate top-k
	if req.TopK < 0 || req.TopK > 2048 {
		return ErrInvalidTopK
	}

	return nil
}

// validatePenaltyParams validates presence and frequency penalty parameters.
func (v *AsyncJobRequestValidator) validatePenaltyParams(req *AsyncJobRequest) error {
	// Validate presence penalty
	if req.PresencePenalty < -2.0 || req.PresencePenalty > 2.0 {
		return ErrInvalidPresencePenalty
	}

	// Validate frequency penalty
	if req.FrequencyPenalty < -2.0 || req.FrequencyPenalty > 2.0 {
		return ErrInvalidFrequencyPenalty
	}

	return nil
}

// validateSearchOptions validates search-related options.
func (v *AsyncJobRequestValidator) validateSearchOptions(req *AsyncJobRequest) error {
	// Validate search domain filter
	if len(req.SearchDomainFilter) > MaxDomainFilterCount {
		return ErrTooManySearchDomains
	}

	// Validate search mode
	if req.SearchMode != "" {
		validModes := map[string]bool{
			"academic": true,
			"web":      true,
		}
		if !validModes[req.SearchMode] {
			return ErrInvalidSearchMode
		}
	}

	// Validate search domain
	if req.SearchDomain != "" {
		validDomains := map[string]bool{
			"sec": true,
		}
		if !validDomains[req.SearchDomain] {
			return ErrInvalidSearchDomain
		}
	}

	return nil
}

// validateImageOptions validates image-related options.
func (v *AsyncJobRequestValidator) validateImageOptions(req *AsyncJobRequest) error {
	// Validate image domain filter
	if len(req.ImageDomainFilter) > MaxDomainFilterCount {
		return ErrTooManyImageDomains
	}

	// Validate image format filter
	if len(req.ImageFormatFilter) > MaxImageFormatFilterCount {
		return ErrTooManyImageFormats
	}

	return nil
}

// validateDateFields validates date-related fields.
func (v *AsyncJobRequestValidator) validateDateFields(req *AsyncJobRequest) error {
	// Validate date formats (basic validation)
	if req.PublishedAfter != "" {
		if err := v.validateDateString(req.PublishedAfter); err != nil {
			return ErrInvalidPublishedAfter
		}
	}

	if req.PublishedBefore != "" {
		if err := v.validateDateString(req.PublishedBefore); err != nil {
			return ErrInvalidPublishedBefore
		}
	}

	if req.LastUpdatedAfterFilter != "" {
		if err := v.validateDateString(req.LastUpdatedAfterFilter); err != nil {
			return ErrInvalidLastUpdatedAfter
		}
	}

	if req.LastUpdatedBeforeFilter != "" {
		if err := v.validateDateString(req.LastUpdatedBeforeFilter); err != nil {
			return ErrInvalidLastUpdatedBefore
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

	return ErrUnsupportedDateFormat
}

// ValidateAsyncJobRequest is a convenience function for validating async job requests.
func ValidateAsyncJobRequest(req *AsyncJobRequest) error {
	validator := NewAsyncJobRequestValidator()
	return validator.Validate(req)
}