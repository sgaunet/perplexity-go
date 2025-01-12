package perplexity

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// DefaultEndpoint is the default endpoint for the Perplexity API.
const DefaultEndpoint = "https://api.perplexity.ai/chat/completions"

// DefautTimeout is the default timeout for the HTTP client.
const DefautTimeout = 10 * time.Second

// Llama31SonarSmall128kOnline is the default model for the Perplexity API.
const Llama31SonarSmall128kOnline = "llama-3.1-sonar-small-128k-online"
const Llama31SonarLarge128kOnline = "llama-3.1-sonar-large-128k-online"
const Llama31SonarHuge128kOnline = "llama-3.1-sonar-huge-128k-online"

// Client is a client for the Perplexity API.
type Client struct {
	endpoint    string
	apiKey      string
	httpClient  *http.Client
	httpTimeout time.Duration
}

// NewClient creates a new Perplexity API client.
// The apiKey is the API key to use for authentication.
// The default model is llama-3-sonar-small-32k-online.
func NewClient(apiKey string) *Client {
	s := &Client{
		apiKey:      apiKey,
		endpoint:    DefaultEndpoint,
		httpClient:  &http.Client{},
		httpTimeout: DefautTimeout,
	}
	return s
}

// SetEndpoint sets the API endpoint.
func (s *Client) SetEndpoint(endpoint string) {
	s.endpoint = endpoint
}

// SetHTTPClient sets the HTTP client.
func (s *Client) SetHTTPClient(httpClient *http.Client) {
	s.httpClient = httpClient
}

// SetHTTPTimeout sets the HTTP timeout.
func (s *Client) SetHTTPTimeout(timeout time.Duration) {
	s.httpTimeout = timeout
}

// GetHTTPTimeout sets the HTTP timeout.
func (s *Client) GetHTTPTimeout() time.Duration {
	return s.httpTimeout
}

// SendCompletionRequest sends a completion request to the Perplexity API.
func (s *Client) SendCompletionRequest(req *CompletionRequest) (*CompletionResponse, error) {
	r := &CompletionResponse{}
	if req == nil {
		return nil, fmt.Errorf("request must not be nil")
	}
	requestBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(s.httpTimeout))
	defer cancel()
	httpReq, err := http.NewRequestWithContext(ctx, "POST", s.endpoint, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+s.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")
	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Check return status code
	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusUnauthorized {
			return nil, fmt.Errorf("unauthorized: check your API key")
		}
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}
	err = json.Unmarshal(body, r)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response body: %w - body response=%s", err, string(body))
	}
	return r, err
}
