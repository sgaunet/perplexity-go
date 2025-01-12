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
