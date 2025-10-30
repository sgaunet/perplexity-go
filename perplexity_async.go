package perplexity

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

// Async API constants and types.
const (
	// AsyncEndpoint is the base endpoint for async operations.
	AsyncEndpoint = "https://api.perplexity.ai/async/chat/completions"

	// AsyncJobTTL is the time-to-live for async jobs (7 days).
	AsyncJobTTL = 7 * 24 * time.Hour
)

// AsyncJobStatus represents the status of an async job.
type AsyncJobStatus string

const (
	// StatusPending indicates the job is queued but not yet processing.
	StatusPending AsyncJobStatus = "pending"
	// StatusProcessing indicates the job is currently being processed.
	StatusProcessing AsyncJobStatus = "processing"
	// StatusCompleted indicates the job has completed successfully.
	StatusCompleted AsyncJobStatus = "completed"
	// StatusFailed indicates the job has failed.
	StatusFailed AsyncJobStatus = "failed"
	// StatusExpired indicates the job has expired (TTL exceeded).
	StatusExpired AsyncJobStatus = "expired"
)

// Error definitions for async operations.
var (
	// ErrAsyncModelRequired is returned when a non-sonar-deep-research model is used for async.
	ErrAsyncModelRequired = errors.New("async jobs require sonar-deep-research model")

	// ErrAsyncJobExpired is returned when an async job has expired.
	ErrAsyncJobExpired = errors.New("async job has expired")

	// ErrAsyncJobFailed is returned when an async job has failed.
	ErrAsyncJobFailed = errors.New("async job failed")

	// ErrAsyncJobNotFound is returned when an async job ID is not found.
	ErrAsyncJobNotFound = errors.New("async job not found")

	// ErrAsyncJobInvalidStatus is returned when an async job has an unexpected status.
	ErrAsyncJobInvalidStatus = errors.New("async job has invalid status")

	// ErrAsyncPollingTimeout is returned when polling times out.
	ErrAsyncPollingTimeout = errors.New("async job polling timeout")

	// ErrJobCompletedNoResult is returned when job is completed but no result is available.
	ErrJobCompletedNoResult = errors.New("job completed but no result available")

	// ErrJobStillProcessing is returned when job is still processing.
	ErrJobStillProcessing = errors.New("job is still processing")

	// ErrUnsupportedTimestampFormat is returned when timestamp format is unsupported.
	ErrUnsupportedTimestampFormat = errors.New("unsupported timestamp format")

	// ErrJobIDEmpty is returned when job ID is empty.
	ErrJobIDEmpty = errors.New("job ID cannot be empty")

	// ErrJobNotComplete is a sentinel error for when job is not yet complete.
	ErrJobNotComplete = errors.New("job not complete")
)

// AsyncJobRequest represents a request to create an async job.
// This is essentially a CompletionRequest but specifically for async operations.
type AsyncJobRequest struct {
	// Messages: The chat conversation messages.
	Messages []Message `json:"messages" validate:"omitempty,dive"`

	// MultimodalMessages: For requests containing images or mixed content.
	MultimodalMessages []MultimodalMessage `json:"-" validate:"omitempty,dive"`

	// Model: Must be "sonar-deep-research" for async operations.
	Model string `json:"model" validate:"required,eq=sonar-deep-research"`

	// ReasoningEffort: Controls computational effort for sonar-deep-research.
	// Options: "low", "medium", "high"
	ReasoningEffort string `json:"reasoning_effort,omitempty" validate:"omitempty,oneof=low medium high"`

	// MaxTokens: Maximum number of completion tokens.
	MaxTokens int `json:"max_tokens" validate:"omitempty,gt=0"`

	// Temperature: Randomness in response (0-2).
	Temperature float64 `json:"temperature" validate:"omitempty,gte=0,lt=2"`

	// TopP: Nucleus sampling threshold (0-1).
	TopP float64 `json:"top_p" validate:"omitempty,gt=0,lt=1"`

	// TopK: Top-k filtering (0-2048).
	TopK int `json:"top_k" validate:"omitempty,gte=0,lte=2048"`

	// PresencePenalty: Penalty for new tokens (-2.0 to 2.0).
	PresencePenalty float64 `json:"presence_penalty" validate:"omitempty,gte=-2,lte=2"`

	// FrequencyPenalty: Penalty for frequent tokens (-2.0 to 2.0).
	FrequencyPenalty float64 `json:"frequency_penalty" validate:"omitempty,gte=-2,lte=2"`

	// SearchDomainFilter: Domains to include or exclude from search.
	SearchDomainFilter []string `json:"search_domain_filter" validate:"omitempty,max=10,dive,required"`

	// SearchMode: Search mode ("academic" or "web").
	SearchMode string `json:"search_mode,omitempty" validate:"omitempty,oneof=academic web"`

	// SearchDomain: Specific domain type ("sec" for SEC documents).
	SearchDomain string `json:"search_domain,omitempty" validate:"omitempty,oneof=sec"`

	// ImageDomainFilter: Domains to include/exclude for image search.
	ImageDomainFilter []string `json:"image_domain_filter,omitempty" validate:"omitempty,max=10,dive,required"`

	// ImageFormatFilter: Image formats to include.
	ImageFormatFilter []string `json:"image_format_filter,omitempty" validate:"omitempty,max=10,dive,required"`

	// PublishedAfter: Filter content published after this date.
	PublishedAfter string `json:"published_after,omitempty" validate:"omitempty"`

	// PublishedBefore: Filter content published before this date.
	PublishedBefore string `json:"published_before,omitempty" validate:"omitempty"`

	// LastUpdatedAfterFilter: Filter content last updated after this date.
	LastUpdatedAfterFilter string `json:"last_updated_after_filter,omitempty" validate:"omitempty"`

	// LastUpdatedBeforeFilter: Filter content last updated before this date.
	LastUpdatedBeforeFilter string `json:"last_updated_before_filter,omitempty" validate:"omitempty"`

	// WebSearchOptions: Web search context and user location.
	WebSearchOptions *WebSearchOptions `json:"web_search_options,omitempty" validate:"omitempty"`

	// Stream: Always false for async requests (async doesn't support streaming).
	Stream bool `json:"stream"`
}

// MarshalJSON implements custom JSON marshaling for AsyncJobRequest.
// The async API requires the request to be wrapped in a "request" field.
func (r *AsyncJobRequest) MarshalJSON() ([]byte, error) {
	// Create the request wrapper structure
	type asyncRequestWrapper struct {
		Request interface{} `json:"request"`
	}

	// If no multimodal messages, use standard marshaling for the inner request
	if len(r.MultimodalMessages) == 0 {
		type alias AsyncJobRequest
		wrapper := asyncRequestWrapper{
			Request: (*alias)(r),
		}
		data, err := json.Marshal(wrapper)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal async request: %w", err)
		}
		return data, nil
	}

	// Create a temporary struct that replaces Messages with MultimodalMessages
	type tempRequest struct {
		Messages                []MultimodalMessage `json:"messages"`
		Model                   string              `json:"model"`
		ReasoningEffort         string              `json:"reasoning_effort,omitempty"`
		MaxTokens               int                 `json:"max_tokens,omitempty"`
		Temperature             float64             `json:"temperature,omitempty"`
		TopP                    float64             `json:"top_p,omitempty"`
		TopK                    int                 `json:"top_k,omitempty"`
		PresencePenalty         float64             `json:"presence_penalty,omitempty"`
		FrequencyPenalty        float64             `json:"frequency_penalty,omitempty"`
		SearchDomainFilter      []string            `json:"search_domain_filter,omitempty"`
		SearchMode              string              `json:"search_mode,omitempty"`
		SearchDomain            string              `json:"search_domain,omitempty"`
		ImageDomainFilter       []string            `json:"image_domain_filter,omitempty"`
		ImageFormatFilter       []string            `json:"image_format_filter,omitempty"`
		PublishedAfter          string              `json:"published_after,omitempty"`
		PublishedBefore         string              `json:"published_before,omitempty"`
		LastUpdatedAfterFilter  string              `json:"last_updated_after_filter,omitempty"`
		LastUpdatedBeforeFilter string              `json:"last_updated_before_filter,omitempty"`
		WebSearchOptions        *WebSearchOptions   `json:"web_search_options,omitempty"`
		Stream                  bool                `json:"stream"`
	}

	temp := tempRequest{
		Messages:                r.MultimodalMessages,
		Model:                   r.Model,
		ReasoningEffort:         r.ReasoningEffort,
		MaxTokens:               r.MaxTokens,
		Temperature:             r.Temperature,
		TopP:                    r.TopP,
		TopK:                    r.TopK,
		PresencePenalty:         r.PresencePenalty,
		FrequencyPenalty:        r.FrequencyPenalty,
		SearchDomainFilter:      r.SearchDomainFilter,
		SearchMode:              r.SearchMode,
		SearchDomain:            r.SearchDomain,
		ImageDomainFilter:       r.ImageDomainFilter,
		ImageFormatFilter:       r.ImageFormatFilter,
		PublishedAfter:          r.PublishedAfter,
		PublishedBefore:         r.PublishedBefore,
		LastUpdatedAfterFilter:  r.LastUpdatedAfterFilter,
		LastUpdatedBeforeFilter: r.LastUpdatedBeforeFilter,
		WebSearchOptions:        r.WebSearchOptions,
		Stream:                  r.Stream,
	}

	wrapper := asyncRequestWrapper{
		Request: temp,
	}
	data, err := json.Marshal(wrapper)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal async request: %w", err)
	}
	return data, nil
}

// IsMultimodal returns true if the request contains multimodal messages.
func (r *AsyncJobRequest) IsMultimodal() bool {
	return len(r.MultimodalMessages) > 0
}

// HasImages returns true if any of the messages contain images.
func (r *AsyncJobRequest) HasImages() bool {
	if !r.IsMultimodal() {
		return false
	}

	for _, msg := range r.MultimodalMessages {
		for _, content := range msg.Content {
			if content.Type == ContentTypeImageURL {
				return true
			}
		}
	}
	return false
}

// AsyncJobResponse represents the response from an async job operation.
type AsyncJobResponse struct {
	// ID: Unique identifier for the async job.
	ID string `json:"id"`

	// Status: Current status of the job.
	Status AsyncJobStatus `json:"status"`

	// CreatedAt: When the job was created.
	CreatedAt time.Time `json:"created_at"`

	// CompletedAt: When the job completed (if applicable).
	CompletedAt *time.Time `json:"completed_at,omitempty"`

	// ExpiresAt: When the job and its results will expire.
	ExpiresAt time.Time `json:"expires_at"`

	// Result: The completion response (only present when status is "completed").
	Result *CompletionResponse `json:"result,omitempty"`

	// Error: Error message (only present when status is "failed").
	Error *string `json:"error,omitempty"`

	// Progress: Optional progress information during processing.
	Progress *AsyncJobProgress `json:"progress,omitempty"`
}

// UnmarshalJSON implements custom JSON unmarshaling for AsyncJobResponse.
// The Perplexity async API returns Unix timestamps as integers, but tests may use RFC3339 strings.
func (r *AsyncJobResponse) UnmarshalJSON(data []byte) error {
	raw, temp, err := r.parseBasicFields(data)
	if err != nil {
		return err
	}

	r.assignBasicFields(&temp)
	return r.parseTimestampFields(raw)
}

// parseTimestamp parses a timestamp from interface{} (either Unix int or RFC3339 string).
func parseTimestamp(value interface{}) (time.Time, error) {
	switch v := value.(type) {
	case float64:
		// Unix timestamp as number
		return time.Unix(int64(v), 0), nil
	case string:
		// RFC3339 string (from tests)
		parsed, err := time.Parse(time.RFC3339, v)
		if err != nil {
			return time.Time{}, fmt.Errorf("failed to parse timestamp: %w", err)
		}
		return parsed, nil
	default:
		return time.Time{}, ErrUnsupportedTimestampFormat
	}
}

// AsyncJobProgress represents progress information for a running async job.
type AsyncJobProgress struct {
	// Stage: Current processing stage.
	Stage string `json:"stage,omitempty"`

	// Percentage: Completion percentage (0-100).
	Percentage *int `json:"percentage,omitempty"`

	// EstimatedTimeRemaining: Estimated time until completion.
	EstimatedTimeRemaining *time.Duration `json:"estimated_time_remaining,omitempty"`
}

// AsyncJobListResponse represents a list of async jobs.
type AsyncJobListResponse struct {
	// Jobs: List of async jobs.
	Jobs []AsyncJobResponse `json:"jobs"`

	// Total: Total number of jobs available.
	Total int `json:"total"`

	// Limit: Maximum number of jobs returned in this response.
	Limit int `json:"limit"`

	// Offset: Offset used for pagination.
	Offset int `json:"offset"`

	// HasMore: Whether there are more jobs available.
	HasMore bool `json:"has_more"`
}

// IsCompleted returns true if the job has completed (successfully or with failure).
func (r *AsyncJobResponse) IsCompleted() bool {
	return r.Status == StatusCompleted || r.Status == StatusFailed
}

// IsExpired returns true if the job has expired.
func (r *AsyncJobResponse) IsExpired() bool {
	return r.Status == StatusExpired || time.Now().After(r.ExpiresAt)
}

// GetResult returns the completion result if available, or an error.
func (r *AsyncJobResponse) GetResult() (*CompletionResponse, error) {
	switch r.Status {
	case StatusCompleted:
		if r.Result == nil {
			return nil, ErrJobCompletedNoResult
		}
		return r.Result, nil
	case StatusFailed:
		errorMsg := "job failed"
		if r.Error != nil {
			errorMsg = *r.Error
		}
		return nil, &AsyncJobError{
			JobID:   r.ID,
			Status:  r.Status,
			Message: errorMsg,
		}
	case StatusExpired:
		return nil, &AsyncJobError{
			JobID:   r.ID,
			Status:  r.Status,
			Message: "job has expired",
		}
	case StatusPending, StatusProcessing:
		return nil, ErrJobStillProcessing
	default:
		return nil, &AsyncJobError{
			JobID:   r.ID,
			Status:  r.Status,
			Message: "unknown job status",
		}
	}
}

// tempResponse is a temporary struct for parsing JSON data.
type tempResponse struct {
	ID       string              `json:"id"`
	Status   AsyncJobStatus      `json:"status"`
	Result   *CompletionResponse `json:"result"`
	Error    *string             `json:"error"`
	Progress *AsyncJobProgress   `json:"progress"`
}

// parseBasicFields parses the basic non-timestamp fields.
func (r *AsyncJobResponse) parseBasicFields(data []byte) (map[string]interface{}, tempResponse, error) {
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, tempResponse{}, fmt.Errorf("failed to unmarshal raw data: %w", err)
	}

	var temp tempResponse
	if err := json.Unmarshal(data, &temp); err != nil {
		return nil, tempResponse{}, fmt.Errorf("failed to unmarshal temp response: %w", err)
	}

	return raw, temp, nil
}

// assignBasicFields assigns the basic fields to the struct.
func (r *AsyncJobResponse) assignBasicFields(temp *tempResponse) {
	r.ID = temp.ID
	r.Status = temp.Status
	r.Result = temp.Result
	r.Error = temp.Error
	r.Progress = temp.Progress
}

// parseTimestampFields parses all timestamp fields.
func (r *AsyncJobResponse) parseTimestampFields(raw map[string]interface{}) error {
	if err := r.parseCreatedAt(raw); err != nil {
		return err
	}
	if err := r.parseCompletedAt(raw); err != nil {
		return err
	}
	return r.parseExpiresAt(raw)
}

// parseCreatedAt parses the created_at timestamp.
func (r *AsyncJobResponse) parseCreatedAt(raw map[string]interface{}) error {
	if createdAt, exists := raw["created_at"]; exists {
		parsed, err := parseTimestamp(createdAt)
		if err != nil {
			return fmt.Errorf("failed to parse created_at: %w", err)
		}
		r.CreatedAt = parsed
	}
	return nil
}

// parseCompletedAt parses the completed_at timestamp (optional).
func (r *AsyncJobResponse) parseCompletedAt(raw map[string]interface{}) error {
	if completedAt, exists := raw["completed_at"]; exists && completedAt != nil {
		parsed, err := parseTimestamp(completedAt)
		if err != nil {
			return fmt.Errorf("failed to parse completed_at: %w", err)
		}
		r.CompletedAt = &parsed
	}
	return nil
}

// parseExpiresAt parses the expires_at timestamp.
func (r *AsyncJobResponse) parseExpiresAt(raw map[string]interface{}) error {
	if expiresAt, exists := raw["expires_at"]; exists {
		parsed, err := parseTimestamp(expiresAt)
		if err != nil {
			return fmt.Errorf("failed to parse expires_at: %w", err)
		}
		r.ExpiresAt = parsed
	}
	return nil
}

// AsyncJobError represents an error related to async job operations.
type AsyncJobError struct {
	JobID   string
	Status  AsyncJobStatus
	Message string
}

// Error implements the error interface.
func (e *AsyncJobError) Error() string {
	return fmt.Sprintf("async job %s (%s): %s", e.JobID, e.Status, e.Message)
}

// Is checks if the error matches a target error.
func (e *AsyncJobError) Is(target error) bool {
	switch e.Status {
	case StatusFailed:
		return errors.Is(target, ErrAsyncJobFailed)
	case StatusExpired:
		return errors.Is(target, ErrAsyncJobExpired)
	case StatusPending, StatusProcessing, StatusCompleted:
		return errors.Is(target, ErrAsyncJobInvalidStatus)
	default:
		return errors.Is(target, ErrAsyncJobInvalidStatus)
	}
}
