// Package perplexity provides a Go client for interacting with the Perplexity AI API.
// It supports both synchronous and streaming (SSE) completion requests.
package perplexity

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

// DefaultEndpoint is the default endpoint for the Perplexity API.
const DefaultEndpoint = "https://api.perplexity.ai/chat/completions"

// DefaultTimeout is the default timeout for the HTTP client.
const DefaultTimeout = 30 * time.Second

// DefaultModel is the default model for the Perplexity API.
const DefaultModel = "sonar"

const defaultSizeSSEResponse = 64000

// Client is a client for the Perplexity API.
type Client struct {
	endpoint   string
	apiKey     string
	httpClient *http.Client
}

// NewClient creates a new Perplexity API client.
// The apiKey is the API key to use for authentication.
// The default model is llama-3-sonar-small-32k-online.
func NewClient(apiKey string) *Client {
	s := &Client{
		apiKey:   apiKey,
		endpoint: DefaultEndpoint,
		httpClient: &http.Client{
			Timeout: DefaultTimeout,
		},
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
	s.httpClient.Timeout = timeout
}

// GetHTTPTimeout sets the HTTP timeout.
func (s *Client) GetHTTPTimeout() time.Duration {
	return s.httpClient.Timeout
}

// SendCompletionRequest sends a completion request to the Perplexity API.
func (s *Client) SendCompletionRequest(req *CompletionRequest) (*CompletionResponse, error) {
	return s.SendCompletionRequestWithContext(context.Background(), req)
}

// SendCompletionRequestWithContext sends a completion request to the Perplexity API with the given context.
func (s *Client) SendCompletionRequestWithContext(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error) {
	r := &CompletionResponse{}
	if req == nil {
		return nil, fmt.Errorf("request must not be nil")
	}
	requestBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, s.endpoint, bytes.NewBuffer(requestBody))
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
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("unexpected status code (%d) and cannot read response: %w", resp.StatusCode, err)
		}
		return nil, ParseErrorMessage(body)
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

// SendSSEHTTPRequest sends a completion request to the Perplexity API using Server-Sent Events.
// It writes each response (event) on the channel responseChannel
// The channel will be closed when the request is done.
func (s *Client) SendSSEHTTPRequest(wg *sync.WaitGroup, req *CompletionRequest, responseChannel chan<- CompletionResponse) error {
	return s.SendSSEHTTPRequestWithContext(context.Background(), wg, req, responseChannel)
}

// SendSSEHTTPRequestWithContext sends a completion request to the Perplexity API using Server-Sent Events with the given context.
// It writes each response (event) on the provided responseChannel.
// The channel will be closed when the request is done.
func (s *Client) SendSSEHTTPRequestWithContext(ctx context.Context, wg *sync.WaitGroup, req *CompletionRequest, responseChannel chan<- CompletionResponse) error {
	if responseChannel == nil {
		return fmt.Errorf("responseChannel must not be nil")
	}
	if wg == nil {
		return fmt.Errorf("wg must not be nil")
	}
	if req == nil {
		return fmt.Errorf("request must not be nil")
	}

	defer close(responseChannel)
	defer wg.Done()

	requestBody, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, s.endpoint, bytes.NewBuffer(requestBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+s.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "text/event-stream")
	httpReq.Header.Set("Cache-Control", "no-cache")
	httpReq.Header.Set("Connection", "keep-alive")

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusUnauthorized {
			return fmt.Errorf("unauthorized: check your API key")
		}
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("unexpected status code (%d) and cannot read response: %w", resp.StatusCode, err)
		}
		return ParseErrorMessage(body)
	}

	// Create a buffered reader for the response body
	reader := bufio.NewReader(resp.Body)
	var buffer bytes.Buffer

	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return fmt.Errorf("error reading response: %w", err)
		}

		// Skip empty lines and comments
		line = bytes.TrimSpace(line)
		if len(line) == 0 || bytes.HasPrefix(line, []byte(":")) {
			continue
		}

		// Check if this is a data line
		if bytes.HasPrefix(line, []byte("data: ")) {
			// Remove the "data: " prefix
			data := bytes.TrimPrefix(line, []byte("data: "))
			data = bytes.TrimSpace(data)

			// Check for [DONE] message
			if bytes.Equal(data, []byte("[DONE]")) {
				break
			}

			// Try to parse the JSON data
			var r CompletionResponse
			if err := json.Unmarshal(data, &r); err != nil {
				// If we have a buffer, try to append to it
				if buffer.Len() > 0 {
					buffer.Write(data)
					if err := json.Unmarshal(buffer.Bytes(), &r); err != nil {
						// If we still can't parse, continue collecting
						continue
					}
					buffer.Reset()
				} else {
					// Start buffering incomplete JSON
					buffer.Write(data)
					continue
				}
			}

			// Send the response to the channel
			select {
			case responseChannel <- r:
			default:
				// Don't block if the channel is full
			}
		}
	}

	return nil
}
