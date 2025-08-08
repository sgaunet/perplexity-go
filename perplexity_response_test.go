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

func TestImageString(t *testing.T) {
	t.Run("nil image returns empty string", func(t *testing.T) {
		var img *perplexity.Image
		assert.Equal(t, "", img.String())
	})

	t.Run("image with URL only", func(t *testing.T) {
		img := &perplexity.Image{
			ImageURL: "https://example.com/image.png",
		}
		assert.Equal(t, "https://example.com/image.png", img.String())
	})

	t.Run("image with URL and origin URL", func(t *testing.T) {
		img := &perplexity.Image{
			ImageURL:  "https://example.com/image.png",
			OriginURL: "https://example.com/source",
		}
		assert.Equal(t, "https://example.com/image.png (from: https://example.com/source)", img.String())
	})

	t.Run("image with URL, origin URL, and dimensions", func(t *testing.T) {
		img := &perplexity.Image{
			ImageURL:  "https://example.com/image.png",
			OriginURL: "https://example.com/source",
			Width:     2016,
			Height:    1512,
		}
		assert.Equal(t, "https://example.com/image.png (from: https://example.com/source) [2016x1512]", img.String())
	})

	t.Run("image with URL and dimensions only", func(t *testing.T) {
		img := &perplexity.Image{
			ImageURL: "https://example.com/image.png",
			Width:    800,
			Height:   600,
		}
		assert.Equal(t, "https://example.com/image.png [800x600]", img.String())
	})
}

func TestGetImages(t *testing.T) {
	t.Run("empty response returns empty images", func(t *testing.T) {
		content := perplexity.CompletionResponse{}
		assert.Equal(t, content.GetImages(), []perplexity.Image{})
	})

	t.Run("nil images returns empty images", func(t *testing.T) {
		content := perplexity.CompletionResponse{
			Images: nil,
		}
		assert.Equal(t, content.GetImages(), []perplexity.Image{})
	})

	t.Run("case with real images", func(t *testing.T) {
		images := []perplexity.Image{
			{
				ImageURL:  "https://content-management-files.canva.com/cdn-cgi/image/f=auto,q=70/b94ec02b-ed6a-47ce-80fc-7eb2d5679e90/ai-face-generator_promo-showcase_012x.png",
				OriginURL: "https://www.canva.com/ai-face-generator/",
				Height:    1512,
				Width:     2016,
			},
			{
				ImageURL: "https://example.com/another-image.jpg",
				Width:    800,
				Height:   600,
			},
		}
		content := perplexity.CompletionResponse{
			Images: &images,
		}
		assert.Equal(t, content.GetImages(), images)
	})
}

func TestGetRelatedQuestions(t *testing.T) {
	t.Run("empty response returns empty related questions", func(t *testing.T) {
		content := perplexity.CompletionResponse{}
		assert.Equal(t, content.GetRelatedQuestions(), []string{})
	})

	t.Run("nil related questions returns empty related questions", func(t *testing.T) {
		content := perplexity.CompletionResponse{
			RelatedQuestions: nil,
		}
		assert.Equal(t, content.GetRelatedQuestions(), []string{})
	})

	t.Run("case with real related questions", func(t *testing.T) {
		relatedQuestions := []string{
			"What are the best free resources for learning to code",
			"How do I choose the right programming language to start with",
			"What are some beginner-friendly coding projects to start with",
			"How can I stay motivated while learning to code",
			"What are the most common mistakes beginners make when learning to code",
		}
		content := perplexity.CompletionResponse{
			RelatedQuestions: &relatedQuestions,
		}
		assert.Equal(t, content.GetRelatedQuestions(), relatedQuestions)
	})
}

func TestGetPostThinkingContent(t *testing.T) {
	t.Run("empty response returns empty string", func(t *testing.T) {
		response := perplexity.CompletionResponse{}
		content := response.GetPostThinkingContent()
		assert.Equal(t, "", content)
	})

	t.Run("no choices returns empty string", func(t *testing.T) {
		response := perplexity.CompletionResponse{
			Choices: []perplexity.Choice{},
		}
		content := response.GetPostThinkingContent()
		assert.Equal(t, "", content)
	})

	t.Run("content without thinking section returns original content", func(t *testing.T) {
		originalContent := "This is a regular response without thinking."
		response := perplexity.CompletionResponse{
			Choices: []perplexity.Choice{
				{
					Message: perplexity.Message{
						Content: originalContent,
					},
				},
			},
		}
		content := response.GetPostThinkingContent()
		assert.Equal(t, originalContent, content)
	})

	t.Run("content with thinking section returns post-thinking content", func(t *testing.T) {
		fullContent := "<think>I need to analyze this question carefully. Let me think about the best approach...</think>\n\nThis is the final answer after thinking."
		expectedContent := "This is the final answer after thinking."
		response := perplexity.CompletionResponse{
			Choices: []perplexity.Choice{
				{
					Message: perplexity.Message{
						Content: fullContent,
					},
				},
			},
		}
		content := response.GetPostThinkingContent()
		assert.Equal(t, expectedContent, content)
	})

	t.Run("content with thinking section and extra whitespace", func(t *testing.T) {
		fullContent := "<think>Let me think about this...</think>   \n\n  Final answer with extra whitespace.  \n\n"
		expectedContent := "Final answer with extra whitespace."
		response := perplexity.CompletionResponse{
			Choices: []perplexity.Choice{
				{
					Message: perplexity.Message{
						Content: fullContent,
					},
				},
			},
		}
		content := response.GetPostThinkingContent()
		assert.Equal(t, expectedContent, content)
	})

	t.Run("content with multiple thinking sections returns content after last one", func(t *testing.T) {
		fullContent := "<think>First thought</think>Some content<think>Second thought</think>\n\nFinal answer."
		expectedContent := "Final answer."
		response := perplexity.CompletionResponse{
			Choices: []perplexity.Choice{
				{
					Message: perplexity.Message{
						Content: fullContent,
					},
				},
			},
		}
		content := response.GetPostThinkingContent()
		assert.Equal(t, expectedContent, content)
	})

	t.Run("content with only opening think tag returns original content", func(t *testing.T) {
		fullContent := "<think>This thinking section never closes..."
		response := perplexity.CompletionResponse{
			Choices: []perplexity.Choice{
				{
					Message: perplexity.Message{
						Content: fullContent,
					},
				},
			},
		}
		content := response.GetPostThinkingContent()
		assert.Equal(t, fullContent, content)
	})

	t.Run("content with nested or complex thinking sections", func(t *testing.T) {
		fullContent := "<think>I need to think about this <nested>tag</nested> carefully.</think>\n\n{\"answer\": \"This is JSON formatted response\"}"
		expectedContent := "{\"answer\": \"This is JSON formatted response\"}"
		response := perplexity.CompletionResponse{
			Choices: []perplexity.Choice{
				{
					Message: perplexity.Message{
						Content: fullContent,
					},
				},
			},
		}
		content := response.GetPostThinkingContent()
		assert.Equal(t, expectedContent, content)
	})

	t.Run("content with empty post-thinking section", func(t *testing.T) {
		fullContent := "<think>I'm thinking...</think>"
		expectedContent := ""
		response := perplexity.CompletionResponse{
			Choices: []perplexity.Choice{
				{
					Message: perplexity.Message{
						Content: fullContent,
					},
				},
			},
		}
		content := response.GetPostThinkingContent()
		assert.Equal(t, expectedContent, content)
	})

	t.Run("multiple choices returns content from last choice", func(t *testing.T) {
		response := perplexity.CompletionResponse{
			Choices: []perplexity.Choice{
				{
					Message: perplexity.Message{
						Content: "<think>First choice thinking</think>First choice answer",
					},
				},
				{
					Message: perplexity.Message{
						Content: "<think>Second choice thinking</think>Second choice answer",
					},
				},
			},
		}
		content := response.GetPostThinkingContent()
		assert.Equal(t, "Second choice answer", content)
	})

	t.Run("real world example with reasoning model response", func(t *testing.T) {
		fullContent := `<think>
The user is asking about the capital of France. This is a straightforward geography question. The capital of France is Paris. I should provide a clear, accurate answer.
</think>

The capital of France is Paris. Paris is not only the political capital but also the largest city in France, serving as the country's economic, cultural, and administrative center.`
		expectedContent := "The capital of France is Paris. Paris is not only the political capital but also the largest city in France, serving as the country's economic, cultural, and administrative center."
		response := perplexity.CompletionResponse{
			Choices: []perplexity.Choice{
				{
					Message: perplexity.Message{
						Content: fullContent,
					},
				},
			},
		}
		content := response.GetPostThinkingContent()
		assert.Equal(t, expectedContent, content)
	})
}
