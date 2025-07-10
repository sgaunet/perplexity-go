package perplexity_test

import (
	"testing"

	"github.com/sgaunet/perplexity-go/v2"
	"github.com/stretchr/testify/assert"
)

func TestGetLastContent(t *testing.T) {
	t.Run("empty response returns nothing", func(t *testing.T) {
		content := perplexity.CompletionResponse{}
		assert.Equal(t, content.GetLastContent(), "")
	})
	t.Run("returns the content if there is only one message", func(t *testing.T) {
		content := perplexity.CompletionResponse{
			Choices: []perplexity.Choice{
				{
					Message: perplexity.Message{
						Role:    "assistant",
						Content: "hello",
					},
				},
			},
		}
		assert.Equal(t, content.GetLastContent(), "hello")
	})
	t.Run("returns the last content of message if there is multiples messages", func(t *testing.T) {
		content := perplexity.CompletionResponse{
			Choices: []perplexity.Choice{
				{
					Message: perplexity.Message{
						Role:    "assistant",
						Content: "hello",
					},
				},
				{
					Message: perplexity.Message{
						Role:    "assistant",
						Content: "hello2",
					},
				},
			},
		}
		assert.Equal(t, content.GetLastContent(), "hello2")
	})
}

func TestString(t *testing.T) {
	t.Run("empty response retuns empty string", func(t *testing.T) {
		content := perplexity.CompletionResponse{}
		assert.Equal(t, content.String(), "")
	})
	t.Run("nil pointer retuns empty string", func(t *testing.T) {
		var content *perplexity.CompletionResponse
		assert.Equal(t, content.String(), "")
	})
	t.Run("case with a real object", func(t *testing.T) {
		content := perplexity.CompletionResponse{
			ID:      "id",
			Model:   "model",
			Created: 1,
			Usage: perplexity.Usage{
				TotalTokens:      1,
				PromptTokens:     1,
				CompletionTokens: 1,
			},
			Object: "object",
			Choices: []perplexity.Choice{
				{
					Message: perplexity.Message{
						Role:    "assistant",
						Content: "hello",
					},
				},
			},
		}
		assert.Equal(t, "{\n  \"id\": \"id\",\n  \"model\": \"model\",\n  \"created\": 1,\n  \"usage\": {\n    \"prompt_tokens\": 1,\n    \"completion_tokens\": 1,\n    \"total_tokens\": 1\n  },\n  \"object\": \"object\",\n  \"choices\": [\n    {\n      \"index\": 0,\n      \"finish_reason\": \"\",\n      \"message\": {\n        \"role\": \"assistant\",\n        \"content\": \"hello\"\n      },\n      \"delta\": {\n        \"role\": \"\",\n        \"content\": \"\"\n      }\n    }\n  ]\n}", content.String())
	})
}

func TestGetCitations(t *testing.T) {
	t.Run("empty response returns empty citations", func(t *testing.T) {
		content := perplexity.CompletionResponse{}
		assert.Equal(t, content.GetCitations(), []string{})
	})

	t.Run("nil citations returns empty citations", func(t *testing.T) {
		content := perplexity.CompletionResponse{
			Citations: nil,
		}
		assert.Equal(t, content.GetCitations(), []string{})
	})

	t.Run("case with a real citations", func(t *testing.T) {
		content := perplexity.CompletionResponse{
			Citations: &[]string{"citation1", "citation2"},
		}
		assert.Equal(t, content.GetCitations(), []string{"citation1", "citation2"})
	})
}

func TestGetSearchResults(t *testing.T) {
	t.Run("empty response returns empty search results", func(t *testing.T) {
		content := perplexity.CompletionResponse{}
		assert.Equal(t, content.GetSearchResults(), []perplexity.SearchResult{})
	})

	t.Run("nil search results returns empty search results", func(t *testing.T) {
		content := perplexity.CompletionResponse{
			SearchResults: nil,
		}
		assert.Equal(t, content.GetSearchResults(), []perplexity.SearchResult{})
	})

	t.Run("case with real search results", func(t *testing.T) {
		searchResults := []perplexity.SearchResult{
			{
				Title: "Test Article 1",
				URL:   "https://example.com/article1",
				Date:  stringPtr("2024-01-01"),
			},
			{
				Title:       "Test Article 2",
				URL:         "https://example.com/article2",
				LastUpdated: stringPtr("2024-01-02"),
			},
		}
		content := perplexity.CompletionResponse{
			SearchResults: &searchResults,
		}
		assert.Equal(t, content.GetSearchResults(), searchResults)
	})
}

// stringPtr returns a pointer to a string value
func stringPtr(s string) *string {
	return &s
}

func TestSearchResultString(t *testing.T) {
	t.Run("nil search result returns empty string", func(t *testing.T) {
		var sr *perplexity.SearchResult
		assert.Equal(t, "", sr.String())
	})

	t.Run("search result with title only", func(t *testing.T) {
		sr := &perplexity.SearchResult{
			Title: "Test Article",
		}
		assert.Equal(t, "Test Article", sr.String())
	})

	t.Run("search result with title and URL", func(t *testing.T) {
		sr := &perplexity.SearchResult{
			Title: "Test Article",
			URL:   "https://example.com/article",
		}
		assert.Equal(t, "Test Article (https://example.com/article)", sr.String())
	})

	t.Run("search result with title, URL, and date", func(t *testing.T) {
		sr := &perplexity.SearchResult{
			Title: "Test Article",
			URL:   "https://example.com/article",
			Date:  stringPtr("2024-01-01"),
		}
		assert.Equal(t, "Test Article (https://example.com/article) - 2024-01-01", sr.String())
	})

	t.Run("search result with title, URL, and last updated", func(t *testing.T) {
		sr := &perplexity.SearchResult{
			Title:       "Test Article",
			URL:         "https://example.com/article",
			LastUpdated: stringPtr("2024-01-02"),
		}
		assert.Equal(t, "Test Article (https://example.com/article) (updated: 2024-01-02)", sr.String())
	})

	t.Run("search result with all fields", func(t *testing.T) {
		sr := &perplexity.SearchResult{
			Title:       "Test Article",
			URL:         "https://example.com/article",
			Date:        stringPtr("2024-01-01"),
			LastUpdated: stringPtr("2024-01-02"),
		}
		assert.Equal(t, "Test Article (https://example.com/article) - 2024-01-01 (updated: 2024-01-02)", sr.String())
	})
}
