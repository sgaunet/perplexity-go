package perplexity_test

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/sgaunet/perplexity-go/v2"
	"github.com/stretchr/testify/assert"
)

const apiKey = "apikey"

func TestGetCompletion(t *testing.T) {
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

		req := perplexity.NewCompletionRequest(perplexity.WithMessages([]perplexity.Message{
			{
				Role:    "user",
				Content: "What's the capital of France?",
			},
		}))
		res, err := r.SendCompletionRequest(req)
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
				assert.Equal(t, string(b), `{"messages":[{"role":"user","content":"What's the capital of France?"}],"model":"llama-3.1-sonar-small-128k-online","max_tokens":0,"temperature":0.2,"top_p":0.9,"search_domain_filter":null,"return_images":false,"return_related_questions":false,"search_recency_filter":"","top_k":0,"stream":false,"presence_penalty":0,"frequency_penalty":1}`)
				w.Header().Add("Content-Type", "application/json")
				fmt.Fprintln(w, "{}")
			}))
		defer ts.Close()

		client := ts.Client()
		r := perplexity.NewClient(apiKey)
		r.SetHTTPClient(client)
		r.SetEndpoint(ts.URL)

		req := perplexity.NewCompletionRequest(perplexity.WithMessages([]perplexity.Message{
			{
				Role:    "user",
				Content: "What's the capital of France?",
			},
		}))
		res, err := r.SendCompletionRequest(req)
		assert.Nil(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, res, &perplexity.CompletionResponse{})
	})

	t.Run("return error if no message to send to the API", func(t *testing.T) {
		r := perplexity.NewClient(apiKey)
		req := perplexity.NewCompletionRequest()
		res, err := r.SendCompletionRequest(req)
		assert.NotNil(t, err)
		assert.Nil(t, res)
	})
}

func TestHTTPTimeout(t *testing.T) {
	t.Run("Check default timeout", func(t *testing.T) {
		r := perplexity.NewClient(apiKey)
		assert.Equal(t, perplexity.DefautTimeout, r.GetHTTPTimeout())
		r.SetHTTPTimeout(1 * time.Second)
		assert.Equal(t, 1*time.Second, r.GetHTTPTimeout())
	})

	t.Run("Check that request take the timeout in account", func(t *testing.T) {
		ts := httptest.NewTLSServer(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				time.Sleep(2 * time.Second)
				w.Header().Add("Content-Type", "application/json")
				fmt.Fprintln(w, "{}")
			}))
		defer ts.Close()

		client := ts.Client()
		r := perplexity.NewClient(apiKey)
		r.SetHTTPClient(client)
		r.SetEndpoint(ts.URL)
		r.SetHTTPTimeout(300 * time.Millisecond)

		startTime := time.Now()
		req := perplexity.NewCompletionRequest(perplexity.WithMessages([]perplexity.Message{
			{
				Role:    "user",
				Content: "What's the capital of France?",
			},
		}))
		res, err := r.SendCompletionRequest(req)
		assert.LessOrEqual(t, time.Since(startTime).Nanoseconds(), int64(350_000_000)) // 350ms
		fmt.Println(time.Since(startTime).Nanoseconds())
		assert.NotNil(t, err) // timeout
		assert.Nil(t, res)
	})
}
