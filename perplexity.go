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

// AsyncJobEndpoint is the base endpoint for async operations.
const AsyncJobEndpoint = "https://api.perplexity.ai/async/chat/completions"

// SearchEndpoint is the endpoint for the Perplexity Search API.
const SearchEndpoint = "https://api.perplexity.ai/search"

// DefaultTimeout is the default timeout for the HTTP client.
const DefaultTimeout = 30 * time.Second

// DefaultModel is the default model for the Perplexity API.
const DefaultModel = "sonar"

// ModelSonarDeepResearch is the sonar-deep-research model that supports reasoning_effort parameter.
const ModelSonarDeepResearch = "sonar-deep-research"

// Error definitions.
var (
	// ErrNilRequest is returned when a nil request is provided.
	ErrNilRequest = errors.New("request must not be nil")
	// ErrUnauthorized is returned when the API key is invalid or missing.
	ErrUnauthorized = errors.New("unauthorized: check your API key")
	// ErrNilResponseChannel is returned when a nil response channel is provided.
	ErrNilResponseChannel = errors.New("response channel must not be nil")
	// ErrNilWaitGroup is returned when a nil wait group is provided.
	ErrNilWaitGroup = errors.New("wait group must not be nil")
)

// Client is a client for the Perplexity API.
type Client struct {
	endpoint       string
	asyncEndpoint  string
	searchEndpoint string
	apiKey         string
	httpClient     *http.Client
}

// NewClient creates a new Perplexity API client.
// The apiKey is the API key to use for authentication.
// The default model is llama-3-sonar-small-32k-online.
func NewClient(apiKey string) *Client {
	s := &Client{
		apiKey:         apiKey,
		endpoint:       DefaultEndpoint,
		asyncEndpoint:  AsyncJobEndpoint,
		searchEndpoint: SearchEndpoint,
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

// SetAsyncEndpoint sets the async API endpoint.
func (s *Client) SetAsyncEndpoint(endpoint string) {
	s.asyncEndpoint = endpoint
}

// SetSearchEndpoint sets the Search API endpoint.
func (s *Client) SetSearchEndpoint(endpoint string) {
	s.searchEndpoint = endpoint
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
		return nil, ErrNilRequest
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
			return nil, ErrUnauthorized
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
	if err := json.Unmarshal(body, r); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response body: %w - body response=%s", err, string(body))
	}
	return r, nil
}

// StreamCompletion sends a completion request to the Perplexity API using Server-Sent Events.
// It blocks until the stream ends or an error occurs, writing each event to responseChannel,
// and closes responseChannel before returning.
//
// Typical usage: run this in its own goroutine and consume the channel from the caller:
//
//	ch := make(chan perplexity.CompletionResponse)
//	errCh := make(chan error, 1)
//	go func() { errCh <- client.StreamCompletion(req, ch) }()
//	for event := range ch { ... }
//	if err := <-errCh; err != nil { ... }
func (s *Client) StreamCompletion(req *CompletionRequest, responseChannel chan<- CompletionResponse) error {
	return s.StreamCompletionWithContext(context.Background(), req, responseChannel)
}

// StreamCompletionWithContext is like StreamCompletion but accepts a context for cancellation.
// Cancelling ctx stops the stream and returns ctx.Err().
func (s *Client) StreamCompletionWithContext(
	ctx context.Context,
	req *CompletionRequest,
	responseChannel chan<- CompletionResponse,
) error {
	if responseChannel == nil {
		return ErrNilResponseChannel
	}
	if req == nil {
		return ErrNilRequest
	}
	defer close(responseChannel)
	return s.streamSSE(ctx, req, responseChannel)
}

// SendSSEHTTPRequest sends a completion request to the Perplexity API using Server-Sent Events.
// The caller must call wg.Add(1) before invoking; this function calls wg.Done() on return and
// closes responseChannel when the request is done.
//
// Deprecated: Use StreamCompletion, which does not require a WaitGroup and applies backpressure
// instead of silently dropping events when the consumer is not ready.
func (s *Client) SendSSEHTTPRequest(wg *sync.WaitGroup, req *CompletionRequest, responseChannel chan<- CompletionResponse) error {
	return s.SendSSEHTTPRequestWithContext(context.Background(), wg, req, responseChannel)
}

// SendSSEHTTPRequestWithContext is like SendSSEHTTPRequest but accepts a context for cancellation.
//
// Deprecated: Use StreamCompletionWithContext, which does not require a WaitGroup and applies
// backpressure instead of silently dropping events when the consumer is not ready.
func (s *Client) SendSSEHTTPRequestWithContext(ctx context.Context, wg *sync.WaitGroup, req *CompletionRequest, responseChannel chan<- CompletionResponse) error {
	if responseChannel == nil {
		return ErrNilResponseChannel
	}
	if wg == nil {
		return ErrNilWaitGroup
	}
	if req == nil {
		return ErrNilRequest
	}
	defer close(responseChannel)
	defer wg.Done()
	return s.streamSSE(ctx, req, responseChannel)
}

// CreateAsyncJob creates a new async job for long-running Sonar Deep Research tasks.
func (s *Client) CreateAsyncJob(req *AsyncJobRequest) (*AsyncJobResponse, error) {
	return s.CreateAsyncJobWithContext(context.Background(), req)
}

// CreateAsyncJobWithContext creates a new async job with the given context.
func (s *Client) CreateAsyncJobWithContext(ctx context.Context, req *AsyncJobRequest) (*AsyncJobResponse, error) {
	if err := s.validateAsyncJobRequest(req); err != nil {
		return nil, err
	}

	httpReq, err := s.prepareAsyncJobRequest(ctx, req)
	if err != nil {
		return nil, err
	}

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	return s.handleAsyncJobResponse(resp)
}

// GetAsyncJob retrieves the status and result of an async job by ID.
func (s *Client) GetAsyncJob(jobID string) (*AsyncJobResponse, error) {
	return s.GetAsyncJobWithContext(context.Background(), jobID)
}

// GetAsyncJobWithContext retrieves the status and result of an async job by ID with the given context.
func (s *Client) GetAsyncJobWithContext(ctx context.Context, jobID string) (*AsyncJobResponse, error) {
	if jobID == "" {
		return nil, ErrJobIDEmpty
	}

	url := fmt.Sprintf("%s/%s", s.asyncEndpoint, jobID)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+s.apiKey)

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Check return status code
	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrAsyncJobNotFound
	}
	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusUnauthorized {
			return nil, ErrUnauthorized
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

	var result AsyncJobResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response body: %w - body response=%s", err, string(body))
	}

	return &result, nil
}

// ListAsyncJobs retrieves a list of async jobs with optional pagination.
func (s *Client) ListAsyncJobs(limit, offset int) (*AsyncJobListResponse, error) {
	return s.ListAsyncJobsWithContext(context.Background(), limit, offset)
}

// ListAsyncJobsWithContext retrieves a list of async jobs with the given context and optional pagination.
func (s *Client) ListAsyncJobsWithContext(ctx context.Context, limit, offset int) (*AsyncJobListResponse, error) {
	if limit <= 0 {
		limit = 20 // Default limit
	}
	if offset < 0 {
		offset = 0
	}

	url := fmt.Sprintf("%s?limit=%d&offset=%d", s.asyncEndpoint, limit, offset)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+s.apiKey)

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Check return status code
	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusUnauthorized {
			return nil, ErrUnauthorized
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

	var result AsyncJobListResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response body: %w - body response=%s", err, string(body))
	}

	return &result, nil
}

// SendSearchRequest sends a search request to the Perplexity Search API.
// The Search API provides direct access to Perplexity's real-time web index
// without the generative LLM layer, returning raw ranked search results.
func (s *Client) SendSearchRequest(req *SearchRequest) (*SearchResponse, error) {
	return s.SendSearchRequestWithContext(context.Background(), req)
}

// SendSearchRequestWithContext sends a search request to the Perplexity Search API with the given context.
func (s *Client) SendSearchRequestWithContext(ctx context.Context, req *SearchRequest) (*SearchResponse, error) {
	if req == nil {
		return nil, ErrNilRequest
	}

	requestBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, s.searchEndpoint, bytes.NewBuffer(requestBody))
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
			return nil, ErrUnauthorized
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

	var result SearchResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response body: %w - body response=%s", err, string(body))
	}

	return &result, nil
}

// streamSSE performs the actual SSE round-trip, decoding events and sending them on
// responseChannel. It does NOT close responseChannel; callers own the channel lifecycle.
// The send is blocking (with ctx.Done() as an escape hatch), so events are never dropped.
func (s *Client) streamSSE(ctx context.Context, req *CompletionRequest, responseChannel chan<- CompletionResponse) error { //nolint:gocognit,cyclop
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
			return ErrUnauthorized
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

		// Check if this is a data line and remove the "data: " prefix
		data, found := bytes.CutPrefix(line, []byte("data: "))
		if !found {
			continue
		}
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

		// Blocking send: respects ctx cancellation, never drops events.
		select {
		case responseChannel <- r:
		case <-ctx.Done():
			return fmt.Errorf("stream cancelled: %w", ctx.Err())
		}
	}

	return nil
}

// validateAsyncJobRequest validates the async job request.
func (s *Client) validateAsyncJobRequest(req *AsyncJobRequest) error {
	if req == nil {
		return ErrNilRequest
	}
	if req.Model != ModelSonarDeepResearch {
		return ErrAsyncModelRequired
	}
	return nil
}

// prepareAsyncJobRequest prepares the HTTP request for async job creation.
func (s *Client) prepareAsyncJobRequest(ctx context.Context, req *AsyncJobRequest) (*http.Request, error) {
	requestBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, s.asyncEndpoint, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+s.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")
	return httpReq, nil
}

// handleAsyncJobResponse handles the HTTP response for async job creation.
func (s *Client) handleAsyncJobResponse(resp *http.Response) (*AsyncJobResponse, error) {
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return s.handleAsyncJobErrorResponse(resp)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var result AsyncJobResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response body: %w - body response=%s", err, string(body))
	}

	return &result, nil
}

// handleAsyncJobErrorResponse handles error responses for async job creation.
func (s *Client) handleAsyncJobErrorResponse(resp *http.Response) (*AsyncJobResponse, error) {
	if resp.StatusCode == http.StatusUnauthorized {
		return nil, ErrUnauthorized
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("unexpected status code (%d) and cannot read response: %w", resp.StatusCode, err)
	}
	return nil, ParseErrorMessage(body)
}
