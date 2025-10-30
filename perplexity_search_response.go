package perplexity

import (
	"encoding/json"
	"fmt"
	"reflect"
)

// SearchResponse represents a response from the Perplexity Search API.
type SearchResponse struct {
	Results []SearchResultItem `json:"results"`
}

// SearchResultItem represents a single search result from the Search API.
// This is distinct from SearchResult which is used in chat completion responses.
type SearchResultItem struct {
	Title   string             `json:"title"`
	URL     string             `json:"url"`
	Snippet *string            `json:"snippet,omitempty"`
	Date    *string            `json:"date,omitempty"`
	Score   *float64           `json:"score,omitempty"`
	Images  *[]SearchImageItem `json:"images,omitempty"`
}

// SearchImageItem represents an image in a search result.
type SearchImageItem struct {
	URL    string `json:"url"`
	Width  *int   `json:"width,omitempty"`
	Height *int   `json:"height,omitempty"`
}

// GetResults returns the search results.
func (r *SearchResponse) GetResults() []SearchResultItem {
	if r == nil {
		return []SearchResultItem{}
	}
	return r.Results
}

// GetResultCount returns the number of search results.
func (r *SearchResponse) GetResultCount() int {
	if r == nil {
		return 0
	}
	return len(r.Results)
}

// String returns a string representation of the SearchResponse.
func (r *SearchResponse) String() string {
	if r == nil {
		return ""
	}
	if reflect.DeepEqual(r, &SearchResponse{}) {
		return ""
	}
	b, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(b)
}

// String returns a string representation of the SearchResultItem.
func (item *SearchResultItem) String() string {
	if item == nil {
		return ""
	}

	result := item.Title
	if item.URL != "" {
		result += " (" + item.URL + ")"
	}
	if item.Date != nil {
		result += " - " + *item.Date
	}
	if item.Score != nil {
		result += fmt.Sprintf(" [score: %.2f]", *item.Score)
	}

	return result
}

// String returns a string representation of the SearchImageItem.
func (img *SearchImageItem) String() string {
	if img == nil {
		return ""
	}

	result := img.URL
	if img.Width != nil && img.Height != nil {
		result += fmt.Sprintf(" [%dx%d]", *img.Width, *img.Height)
	}

	return result
}
