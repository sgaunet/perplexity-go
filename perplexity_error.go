package perplexity

import (
	"encoding/json"
	"fmt"
)

// ErrorResponse is an error response object for the Perplexity API.
type ErrorResponse struct {
	ErrorData struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Code    int    `json:"code"`
	} `json:"error"`
}

// Error returns the error message of the ErrorResponse.
func (r *ErrorResponse) Error() string {
	if r == nil {
		return ""
	}
	return r.ErrorData.Message
}

// ParseErrorMessage unmarshals a []byte representing an API error response
// and returns the error message.
func ParseErrorMessage(data []byte) error {
	var errResp ErrorResponse
	if err := json.Unmarshal(data, &errResp); err != nil {
		return fmt.Errorf("failed to unmarshal error response: %w", err)
	}
	return &errResp
}
