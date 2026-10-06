package perplexity

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

// ResponseError is an error response object for the Perplexity API.
type ResponseError struct {
	ErrorData struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Code    int    `json:"code"`
	} `json:"error"`
	// Detail holds validation errors returned with HTTP 422 ({"detail": [...]}).
	Detail []ValidationErrorDetail `json:"detail,omitempty"`
	// StatusCode is the HTTP status code of the response, when known.
	StatusCode int `json:"-"`
}

// ValidationErrorDetail is a single entry of a 422 validation error response.
type ValidationErrorDetail struct {
	Loc  []any  `json:"loc"`
	Msg  string `json:"msg"`
	Type string `json:"type"`
}

// Error returns the error message of the ResponseError.
// When no message is available, it falls back to the validation details or the HTTP status code.
func (r *ResponseError) Error() string {
	if r == nil {
		return ""
	}
	if r.ErrorData.Message != "" {
		return r.ErrorData.Message
	}
	if len(r.Detail) > 0 {
		msgs := make([]string, 0, len(r.Detail))
		for _, d := range r.Detail {
			loc := make([]string, 0, len(d.Loc))
			for _, l := range d.Loc {
				loc = append(loc, fmt.Sprint(l))
			}
			msgs = append(msgs, fmt.Sprintf("%s: %s", strings.Join(loc, "."), d.Msg))
		}
		return "validation error: " + strings.Join(msgs, "; ")
	}
	if r.StatusCode != 0 {
		return fmt.Sprintf("unexpected status code %d", r.StatusCode)
	}
	return ""
}

// parseHTTPErrorResponse builds a *ResponseError from an HTTP error response, keeping the status code.
// It understands both {"error": {...}} and 422 {"detail": [...]} bodies, and falls back to the raw
// body as message when the body is not in a known JSON format.
func parseHTTPErrorResponse(statusCode int, data []byte) *ResponseError {
	errResp := &ResponseError{}
	if err := json.Unmarshal(data, errResp); err != nil {
		errResp = &ResponseError{}
		errResp.ErrorData.Message = strings.TrimSpace(string(data))
	}
	errResp.StatusCode = statusCode
	if errResp.ErrorData.Message == "" && len(errResp.Detail) == 0 && len(data) > 0 {
		errResp.ErrorData.Message = strings.TrimSpace(string(data))
	}
	return errResp
}

// ParseErrorMessage unmarshals a []byte representing an API error response
// and returns the error message.
func ParseErrorMessage(data []byte) error {
	var errResp ResponseError
	if err := json.Unmarshal(data, &errResp); err != nil {
		return fmt.Errorf("failed to unmarshal error response: %w", err)
	}
	return &errResp
}

// AsyncJobAPIError represents errors specific to async job operations.
type AsyncJobAPIError struct {
	StatusCode int
	Message    string
	JobID      string
	RequestID  string
}

// NewAsyncJobAPIError creates a new AsyncJobAPIError.
func NewAsyncJobAPIError(statusCode int, message, jobID, requestID string) *AsyncJobAPIError {
	return &AsyncJobAPIError{
		StatusCode: statusCode,
		Message:    message,
		JobID:      jobID,
		RequestID:  requestID,
	}
}

// Error implements the error interface for AsyncJobAPIError.
func (e *AsyncJobAPIError) Error() string {
	if e.JobID != "" {
		return fmt.Sprintf("async job API error (job: %s, status: %d): %s", e.JobID, e.StatusCode, e.Message)
	}
	return fmt.Sprintf("async job API error (status: %d): %s", e.StatusCode, e.Message)
}

// IsAsyncJobError checks if an error is an AsyncJobAPIError.
func IsAsyncJobError(err error) bool {
	var asyncErr *AsyncJobAPIError
	return errors.As(err, &asyncErr)
}

// GetAsyncJobError extracts AsyncJobAPIError from an error.
func GetAsyncJobError(err error) (*AsyncJobAPIError, bool) {
	var asyncErr *AsyncJobAPIError
	if errors.As(err, &asyncErr) {
		return asyncErr, true
	}
	return nil, false
}

// IsRetryableAsyncError determines if an async job error is retryable.
func IsRetryableAsyncError(err error) bool {
	var asyncErr *AsyncJobAPIError
	if !errors.As(err, &asyncErr) {
		return false
	}

	// Retryable status codes
	switch asyncErr.StatusCode {
	case http.StatusTooManyRequests, // 429
		http.StatusInternalServerError, // 500
		http.StatusBadGateway,          // 502
		http.StatusServiceUnavailable,  // 503
		http.StatusGatewayTimeout:      // 504
		return true
	default:
		return false
	}
}

// AsyncJobTimeoutError represents timeout errors for async operations.
type AsyncJobTimeoutError struct {
	JobID   string
	Timeout string
	Elapsed string
}

// Error implements the error interface for AsyncJobTimeoutError.
func (e *AsyncJobTimeoutError) Error() string {
	return fmt.Sprintf("async job %s timed out after %s (timeout: %s)", e.JobID, e.Elapsed, e.Timeout)
}

// Is checks if the error matches ErrAsyncPollingTimeout.
func (e *AsyncJobTimeoutError) Is(target error) bool {
	return errors.Is(target, ErrAsyncPollingTimeout)
}

// AsyncJobValidationError represents validation errors for async job requests.
type AsyncJobValidationError struct {
	Field   string
	Value   any
	Message string
}

// NewAsyncJobValidationError creates a new validation error.
func NewAsyncJobValidationError(field string, value any, message string) *AsyncJobValidationError {
	return &AsyncJobValidationError{
		Field:   field,
		Value:   value,
		Message: message,
	}
}

// Error implements the error interface for AsyncJobValidationError.
func (e *AsyncJobValidationError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("validation error for field '%s': %s (value: %v)", e.Field, e.Message, e.Value)
	}
	return "validation error: " + e.Message
}

// IsAsyncJobValidationError checks if an error is an AsyncJobValidationError.
func IsAsyncJobValidationError(err error) bool {
	var validationErr *AsyncJobValidationError
	return errors.As(err, &validationErr)
}

// ParseAsyncJobErrorResponse parses async job specific error responses.
func ParseAsyncJobErrorResponse(statusCode int, data []byte, jobID string) error {
	// Try to parse as standard error response first
	var errResp ResponseError
	if err := json.Unmarshal(data, &errResp); err == nil {
		return NewAsyncJobAPIError(statusCode, errResp.ErrorData.Message, jobID, "")
	}

	// Try to parse as async job specific error format
	var asyncErrResp struct {
		Error struct {
			Message   string `json:"message"`
			Type      string `json:"type"`
			Code      string `json:"code"`
			JobID     string `json:"job_id"`
			RequestID string `json:"request_id"`
		} `json:"error"`
	}

	if err := json.Unmarshal(data, &asyncErrResp); err == nil {
		return NewAsyncJobAPIError(
			statusCode,
			asyncErrResp.Error.Message,
			asyncErrResp.Error.JobID,
			asyncErrResp.Error.RequestID,
		)
	}

	// Fallback to generic error
	return NewAsyncJobAPIError(statusCode, string(data), jobID, "")
}
