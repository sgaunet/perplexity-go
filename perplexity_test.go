package perplexity_test

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sgaunet/perplexity-go"
	"github.com/stretchr/testify/assert"
)

func TestGetCompletion(t *testing.T) {
	apiKey := "apikey"
	t.Run("handles wrong response format", func(t *testing.T) {
		ts := httptest.NewTLSServer(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Add("Content-Type", "application/json")
				fmt.Fprintln(w, "not json")
			}))
		defer ts.Close()

		client := ts.Client()
		r := perplexity.NewClient(apiKey)
		r.SetHTTPClient(client)
		r.SetEndpoint(ts.URL)

		res, err := r.CreateCompletion([]perplexity.Message{
			{
				Role:    "user",
				Content: "What's the capital of France?",
			},
		})
		assert.NotNil(t, err)
		assert.Nil(t, res)
	})
	t.Run("send payload successfully", func(t *testing.T) {
		ts := httptest.NewTLSServer(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Check the request
				assert.Equal(t, r.Method, "POST")
				assert.Equal(t, r.Header.Get("Authorization"), "Bearer "+apiKey)
				assert.Equal(t, r.Header.Get("Content-Type"), "application/json")
				defer r.Body.Close()
				b, err := io.ReadAll(r.Body)
				assert.Nil(t, err)
				assert.Equal(t, string(b), `{"messages":[{"role":"user","content":"What's the capital of France?"}],"model":"llama-3-sonar-small-32k-online"}`)
				w.Header().Add("Content-Type", "application/json")
				fmt.Fprintln(w, "{}")
			}))
		defer ts.Close()

		client := ts.Client()
		r := perplexity.NewClient(apiKey)
		r.SetHTTPClient(client)
		r.SetEndpoint(ts.URL)

		res, err := r.CreateCompletion([]perplexity.Message{
			{
				Role:    "user",
				Content: "What's the capital of France?",
			},
		})
		assert.Nil(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, res, &perplexity.CompletionResponse{})
	})

	t.Run("return error if no message to send to the API", func(t *testing.T) {
		r := perplexity.NewClient(apiKey)
		res, err := r.CreateCompletion([]perplexity.Message{})
		assert.NotNil(t, err)
		assert.Nil(t, res)
	})
}

func TestSetModels(t *testing.T) {
	r := perplexity.NewClient("apikey")
	t.Run("set model llama-3-sonar-small-32k-chat", func(t *testing.T) {
		r.SetModuleLlama3SonarSmall32kChat()
		assert.Equal(t, r.GetModel(), "llama-3-sonar-small-32k-chat")
	})
	t.Run("set model llama-3-sonar-small-32k-online", func(t *testing.T) {
		r.SetModuleLlama3SonarSmall32kOnline()
		assert.Equal(t, r.GetModel(), "llama-3-sonar-small-32k-online")
	})
	t.Run("set model llama-3-sonar-large-32k-chat", func(t *testing.T) {
		r.SetModuleLlama3SonarLarge32kChat()
		assert.Equal(t, r.GetModel(), "llama-3-sonar-large-32k-chat")
	})
	t.Run("set model llama-3-sonar-large-32k-online", func(t *testing.T) {
		r.SetModuleLlama3SonarLarge32kOnline()
		assert.Equal(t, r.GetModel(), "llama-3-sonar-large-32k-online")
	})
	t.Run("set model llama-3-8b-instruct", func(t *testing.T) {
		r.SetModuleLlama38bInstruct()
		assert.Equal(t, r.GetModel(), "llama-3-8b-instruct")
	})
	t.Run("set model llama-3-70b-instruct", func(t *testing.T) {
		r.SetModuleLlama370bInstruct()
		assert.Equal(t, r.GetModel(), "llama-3-70b-instruct")
	})
	t.Run("set model mixtral-8x7b-instruct", func(t *testing.T) {
		r.SetModuleMixtral8x7bInstruct()
		assert.Equal(t, r.GetModel(), "mixtral-8x7b-instruct")
	})
}

func TestGetLastContent(t *testing.T) {
	t.Run("empty response retuens nothing", func(t *testing.T) {
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
