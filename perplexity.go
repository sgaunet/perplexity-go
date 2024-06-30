package perplexity

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"time"
)

const DefaultEndpoint = "https://api.perplexity.ai/chat/completions"

// Message is a message object for the Perplexity API.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// CompletionRequest is a request object for the Perplexity API.
type CompletionRequest struct {
	Messages []Message `json:"messages"`
	Model    string    `json:"model"`
}

// Usage is a usage object for the Perplexity API.
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// Choice is a choice object for the Perplexity API.
type Choice struct {
	Index        int     `json:"index"`
	FinishReason string  `json:"finish_reason"`
	Message      Message `json:"message"`
	Delta        Message `json:"delta"`
}

// CompletionResponse is a response object for the Perplexity API.
type CompletionResponse struct {
	ID      string   `json:"id"`
	Model   string   `json:"model"`
	Created int      `json:"created"`
	Usage   Usage    `json:"usage"`
	Object  string   `json:"object"`
	Choices []Choice `json:"choices"`
}

// Client is a client for the Perplexity API.
type Client struct {
	endpoint   string
	apiKey     string
	model      string
	httpClient *http.Client
}

// NewClient creates a new Perplexity API client.
// The apiKey is the API key to use for authentication.
// The default model is llama-3-sonar-small-32k-online.
func NewClient(apiKey string) *Client {
	s := &Client{
		apiKey:     apiKey,
		endpoint:   DefaultEndpoint,
		httpClient: &http.Client{},
	}
	s.SetModuleLlama3SonarSmall32kOnline()
	return s
}

// setModel sets the model to use for completions.
func (s *Client) setModel(model string) {
	s.model = model
}

// GetModel returns the model currently in use.
func (s *Client) GetModel() string {
	return s.model
}

// SetModuleLlama3SonarSmall32kChat sets the model to llama-3-sonar-small-32k-chat.
func (s *Client) SetModuleLlama3SonarSmall32kChat() {
	s.setModel("llama-3-sonar-small-32k-chat")
}

// SetModuleLlama3SonarSmall32kOnline sets the model to llama-3-sonar-small-32k-online.
func (s *Client) SetModuleLlama3SonarSmall32kOnline() {
	s.setModel("llama-3-sonar-small-32k-online")
}

// SetModuleLlama3SonarLarge32kChat sets the model to llama-3-sonar-large-32k-chat.
func (s *Client) SetModuleLlama3SonarLarge32kChat() {
	s.setModel("llama-3-sonar-large-32k-chat")
}

// SetModuleLlama3SonarLarge32kOnline sets the model to llama-3-sonar-large-32k-online.
func (s *Client) SetModuleLlama3SonarLarge32kOnline() {
	s.setModel("llama-3-sonar-large-32k-online")
}

// SetModuleLlama3SonarLarge32kInstruct sets the model to llama-3-sonar-large-32k-instruct.
func (s *Client) SetModuleLlama38bInstruct() {
	s.setModel("llama-3-8b-instruct")
}

// SetModuleLlama370bInstruct sets the model to llama-3-70b-instruct.
func (s *Client) SetModuleLlama370bInstruct() {
	s.setModel("llama-3-70b-instruct")
}

// SetModuleMixtral8x7bInstruct sets the model to mixtral-8x7b-instruct.
func (s *Client) SetModuleMixtral8x7bInstruct() {
	s.setModel("mixtral-8x7b-instruct")
}

// SetEndpoint sets the API endpoint.
func (s *Client) SetEndpoint(endpoint string) {
	s.endpoint = endpoint
}

// SetHTTPClient sets the HTTP client.
func (s *Client) SetHTTPClient(httpClient *http.Client) {
	s.httpClient = httpClient
}

// CreateCompletion sends simple text to the Perplexity API and retrieve the response.
func (s *Client) CreateCompletion(messages []Message) (*CompletionResponse, error) {
	r := &CompletionResponse{}
	if len(messages) == 0 {
		return nil, fmt.Errorf("messages must not be empty")
	}
	reqBody := CompletionRequest{
		Messages: messages,
		Model:    s.model,
	}
	requestBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(5*time.Second))
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "POST", s.endpoint, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	req.Header.Set("Content-Type", "application/json")
	resp, err := s.httpClient.Do(req)
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

func (r *CompletionResponse) String() string {
	if r == nil {
		return ""
	}
	if reflect.DeepEqual(r, &CompletionResponse{}) {
		return ""
	}
	b, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(b)
}

func (r *CompletionResponse) GetLastContent() string {
	if len(r.Choices) == 0 {
		return ""
	}
	return r.Choices[len(r.Choices)-1].Message.Content
}
