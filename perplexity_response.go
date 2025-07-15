package perplexity

import (
	"encoding/json"
	"fmt"
	"reflect"
)

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
	ID            string          `json:"id"`
	Model         string          `json:"model"`
	Created       int             `json:"created"`
	Usage         Usage           `json:"usage"`
	Object        string          `json:"object"`
	Choices       []Choice        `json:"choices"`
	SearchResults *[]SearchResult `json:"search_results,omitempty"`
	Images        *[]Image        `json:"images,omitempty"`
	// Deprecated: Use SearchResults instead for better structured data with titles, URLs, and metadata.
	//go:deprecated
	Citations *[]string `json:"citations,omitempty"`
}

// SearchResult represents a single search result in the Perplexity API response.
type SearchResult struct {
	Title       string  `json:"title"`
	URL         string  `json:"url"`
	Date        *string `json:"date,omitempty"`
	LastUpdated *string `json:"last_updated,omitempty"`
}

// Image represents a single image in the Perplexity API response.
type Image struct {
	ImageURL  string `json:"image_url"`
	OriginURL string `json:"origin_url"`
	Height    int    `json:"height"`
	Width     int    `json:"width"`
}

// String returns a string representation of the SearchResult.
func (sr *SearchResult) String() string {
	if sr == nil {
		return ""
	}

	result := sr.Title
	if sr.URL != "" {
		result += " (" + sr.URL + ")"
	}
	if sr.Date != nil {
		result += " - " + *sr.Date
	}
	if sr.LastUpdated != nil {
		result += " (updated: " + *sr.LastUpdated + ")"
	}

	return result
}

// String returns a string representation of the Image.
func (img *Image) String() string {
	if img == nil {
		return ""
	}

	result := img.ImageURL
	if img.OriginURL != "" {
		result += " (from: " + img.OriginURL + ")"
	}
	if img.Width > 0 && img.Height > 0 {
		result += fmt.Sprintf(" [%dx%d]", img.Width, img.Height)
	}

	return result
}

// String returns a string representation of the CompletionResponse.
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

// GetLastContent returns the last content of the completion response.
func (r *CompletionResponse) GetLastContent() string {
	if len(r.Choices) == 0 {
		return ""
	}
	return r.Choices[len(r.Choices)-1].Message.Content
}

// GetCitations returns the citations of the completion response.
// Deprecated: Use GetSearchResults instead for better structured data with titles, URLs, and metadata.
//
//go:deprecated
func (r *CompletionResponse) GetCitations() []string {
	if r.Citations == nil {
		return []string{}
	}
	return *r.Citations
}

// GetSearchResults returns the search results of the completion response.
func (r *CompletionResponse) GetSearchResults() []SearchResult {
	if r.SearchResults == nil {
		return []SearchResult{}
	}
	return *r.SearchResults
}

// GetImages returns the images of the completion response.
func (r *CompletionResponse) GetImages() []Image {
	if r.Images == nil {
		return []Image{}
	}
	return *r.Images
}
