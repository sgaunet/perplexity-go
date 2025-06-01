package perplexity

import (
	"encoding/json"
	"fmt"
)

// ResponseError is an error response object for the Perplexity API.
type ResponseError struct {
	ErrorData struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Code    int    `json:"code"`
	} `json:"error"`
}

// Error returns the error message of the ResponseError.
func (r *ResponseError) Error() string {
	if r == nil {
		return ""
	}
	return r.ErrorData.Message
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
