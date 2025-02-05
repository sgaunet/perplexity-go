package perplexity

import (
	"bytes"
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

// DefautTimeout is the default timeout for the HTTP client.
const DefautTimeout = 10 * time.Second

// DefaultModel is the default model for the Perplexity API.
const DefaultModel = "sonar"
const ProModel = "sonar-pro"

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
			Timeout: DefautTimeout,
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
	// if req.Stream {
	// 	return s.sendSSEHTTPRequest(req)
	// }
	return s.SendHTTPRequest(req)
}

// sendHTTPRequest sends a completion request to the Perplexity API. (basic http request, not SSE)
func (s *Client) SendHTTPRequest(req *CompletionRequest) (*CompletionResponse, error) {
	r := &CompletionResponse{}
	if req == nil {
		return nil, fmt.Errorf("request must not be nil")
	}
	requestBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}
	httpReq, err := http.NewRequest("POST", s.endpoint, bytes.NewBuffer(requestBody))
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

// SendSSEHTTPRequest sends a completion request to the Perplexity API using Server-Sent Events.
// It writes each response (event) on the channel responseChannel
// The channel will be closed when the request is done.
func (s *Client) SendSSEHTTPRequest(wg *sync.WaitGroup, req *CompletionRequest, responseChannel chan<- CompletionResponse) error {
	if responseChannel == nil {
		return fmt.Errorf("responseChannel must not be nil")
	}
	defer close(responseChannel)
	defer wg.Done()
	if req == nil {
		return fmt.Errorf("request must not be nil")
	}
	requestBody, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %w", err)
	}
	httpReq, err := http.NewRequest("POST", s.endpoint, bytes.NewBuffer(requestBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+s.apiKey)
	httpReq.Header.Set("Cache-Control", "no-cache")
	httpReq.Header.Set("Accept", "text/event-stream")
	httpReq.Header.Set("Connection", "keep-alive")

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// lastMessage is used to store the last message in case of truncation
	// received events may be truncated
	var lastMessage []byte
	for {
		var tmpData []byte
		data := make([]byte, defaultSizeSSEResponse)
		_, errBody := resp.Body.Read(data)
		if errBody != io.EOF && err != nil {
			return fmt.Errorf("failed to read response body: %w", err)
		}
		// // Trim 'data: ' from the beginning of the message
		// if len(data) < 6 {
		// 	return fmt.Errorf("invalid message: %s", string(data))
		// }

		// split tmpData by '\r\n\r\n'
		splittedData := bytes.Split(data, []byte("\r\n\r\n"))
	loop:
		for _, d := range splittedData {
			// Check if the last message has been truncated
			// if not, we can directly use the data
			if len(lastMessage) == 0 {
				tmpData = d[6:]
			}
			// if the last message has been truncated, we need to concatenate the last message with the next one
			if len(lastMessage) > 0 {
				tmpData = append(lastMessage, d[6:]...)
				lastMessage = nil
			}
			// trim nil bytes
			tmpData = bytes.Trim(tmpData, "\x00")
			if len(tmpData) == 0 {
				break loop
			}
			var r CompletionResponse
			err = json.Unmarshal(tmpData, &r)
			if err != nil {
				// we ignore the error because the last message has been truncated
				// we need to concatenate the last message with the next one
				lastMessage = tmpData
				break loop
			}
			// Append the response to the full response
			responseChannel <- r
		}
		// Check if it's the end of the stream
		if errors.Is(errBody, io.EOF) {
			break
		}
	}
	// Check return status code
	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusUnauthorized {
			return fmt.Errorf("unauthorized: check your API key")
		}
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	return nil
}
