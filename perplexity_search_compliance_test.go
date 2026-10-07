package perplexity

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSearchAPIEndpointCompliance verifies that the Search API endpoint configuration
// matches the official Perplexity API specification.
func TestSearchAPIEndpointCompliance(t *testing.T) {
	t.Run("correct endpoint URL", func(t *testing.T) {
		// Specification: POST https://api.perplexity.ai/search
		expectedEndpoint := "https://api.perplexity.ai/search"
		assert.Equal(t, expectedEndpoint, SearchEndpoint, "SearchEndpoint must match API specification")
	})

	t.Run("client initializes with correct endpoint", func(t *testing.T) {
		client := NewClient("test-api-key")
		assert.NotNil(t, client)
		assert.Equal(t, SearchEndpoint, client.searchEndpoint)
	})

	t.Run("endpoint can be customized", func(t *testing.T) {
		client := NewClient("test-api-key")
		customEndpoint := "https://custom.api.endpoint/search"
		client.SetSearchEndpoint(customEndpoint)
		assert.Equal(t, customEndpoint, client.searchEndpoint)
	})
}

// TestSearchRequestStructureCompliance verifies the SearchRequest structure
// matches the API specification exactly.
func TestSearchRequestStructureCompliance(t *testing.T) {
	t.Run("request structure has all required fields", func(t *testing.T) {
		// API Spec: query (required), max_results, return_images, return_snippets, country, search_domain_filter
		req := &SearchRequest{
			Query:              "test query",
			MaxResults:         intPtr(10),
			ReturnImages:       boolPtr(true),
			ReturnSnippets:     boolPtr(true),
			Country:            strPtr("US"),
			SearchDomainFilter: &[]string{"example.com"},
		}

		// Verify all fields are present
		require.NotNil(t, req)
		assert.NotNil(t, req.Query)
		assert.NotNil(t, req.MaxResults)
		assert.NotNil(t, req.ReturnImages)
		assert.NotNil(t, req.ReturnSnippets)
		assert.NotNil(t, req.Country)
		assert.NotNil(t, req.SearchDomainFilter)
	})

	t.Run("query field accepts string", func(t *testing.T) {
		// Spec: query can be a string
		req := NewSearchRequest("test query")
		assert.IsType(t, "", req.Query)
		assert.Equal(t, "test query", req.Query)
	})

	t.Run("query field accepts array of strings", func(t *testing.T) {
		// Spec: query can be an array of strings for multi-query
		queries := []string{"query1", "query2", "query3"}
		req := NewSearchRequest(queries)
		assert.IsType(t, []string{}, req.Query)
		assert.Equal(t, queries, req.Query)
	})

	t.Run("JSON marshaling uses correct field names", func(t *testing.T) {
		// Verify JSON field names match API spec exactly
		req := NewSearchRequest(
			"test",
			WithSearchMaxResults(5),
			WithSearchReturnImages(true),
			WithSearchReturnSnippets(false),
			WithSearchCountry("US"),
			WithSearchDomains([]string{"example.com"}),
		)

		data, err := json.Marshal(req)
		require.NoError(t, err)

		var unmarshaled map[string]interface{}
		err = json.Unmarshal(data, &unmarshaled)
		require.NoError(t, err)

		// Verify exact field names from spec
		assert.Contains(t, unmarshaled, "query")
		assert.Contains(t, unmarshaled, "max_results")
		assert.Contains(t, unmarshaled, "return_images")
		assert.Contains(t, unmarshaled, "return_snippets")
		assert.Contains(t, unmarshaled, "country")
		assert.Contains(t, unmarshaled, "search_domain_filter")
	})

	t.Run("optional fields are omitted when not set", func(t *testing.T) {
		req := NewSearchRequest("test")

		data, err := json.Marshal(req)
		require.NoError(t, err)

		var unmarshaled map[string]interface{}
		err = json.Unmarshal(data, &unmarshaled)
		require.NoError(t, err)

		// Only query should be present
		assert.Contains(t, unmarshaled, "query")
		assert.NotContains(t, unmarshaled, "max_results")
		assert.NotContains(t, unmarshaled, "return_images")
		assert.NotContains(t, unmarshaled, "return_snippets")
		assert.NotContains(t, unmarshaled, "country")
		assert.NotContains(t, unmarshaled, "search_domain_filter")
		assert.Len(t, unmarshaled, 1)
	})

	t.Run("all current spec fields use correct JSON names", func(t *testing.T) {
		day := time.Date(2025, 3, 1, 0, 0, 0, 0, time.UTC)
		req := NewSearchRequest(
			"test",
			WithSearchType(SearchTypeFast),
			WithSearchMaxResults(5),
			WithSearchMaxTokens(10000),
			WithSearchMaxTokensPerPage(2048),
			WithSearchCountry("US"),
			WithSearchDomains([]string{"example.com"}),
			WithSearchLanguageFilter([]string{"en"}),
			WithSearchPublishedAfter(day),
			WithSearchPublishedBefore(day),
			WithSearchUpdatedAfter(day),
			WithSearchUpdatedBefore(day),
			WithSearchDisplayServerTime(true),
		)

		data, err := json.Marshal(req)
		require.NoError(t, err)

		var m map[string]any
		require.NoError(t, json.Unmarshal(data, &m))

		assert.Equal(t, "fast", m["search_type"])
		assert.InDelta(t, 10000, m["max_tokens"], 0)
		assert.InDelta(t, 2048, m["max_tokens_per_page"], 0)
		assert.Equal(t, []any{"en"}, m["search_language_filter"])
		assert.Equal(t, "3/1/2025", m["search_after_date_filter"])
		assert.Equal(t, "3/1/2025", m["search_before_date_filter"])
		assert.Equal(t, "3/1/2025", m["last_updated_after_filter"])
		assert.Equal(t, "3/1/2025", m["last_updated_before_filter"])
		assert.Equal(t, true, m["display_server_time"])

		recency, err := json.Marshal(NewSearchRequest("test", WithSearchRecency(SearchRecencyWeek)))
		require.NoError(t, err)
		assert.Contains(t, string(recency), `"search_recency_filter":"week"`)
	})
}

// TestSearchResponseStructureCompliance verifies the SearchResponse structure
// matches the API specification exactly.
func TestSearchResponseStructureCompliance(t *testing.T) {
	t.Run("response structure matches spec", func(t *testing.T) {
		// Spec: { "results": [ { "title", "url", "snippet", "date", "score", "images" } ] }
		resp := &SearchResponse{
			Results: []SearchResultItem{
				{
					Title:   "Test Result",
					URL:     "https://example.com",
					Snippet: strPtr("Test snippet"),
					Date:    strPtr("2024-01-01"),
					Score:   floatPtr(0.95),
					Images: &[]SearchImageItem{
						{
							URL:    "https://example.com/image.jpg",
							Width:  intPtr(800),
							Height: intPtr(600),
						},
					},
				},
			},
		}

		require.NotNil(t, resp)
		require.Len(t, resp.Results, 1)

		item := resp.Results[0]
		assert.NotEmpty(t, item.Title)
		assert.NotEmpty(t, item.URL)
		assert.NotNil(t, item.Snippet)
		assert.NotNil(t, item.Date)
		assert.NotNil(t, item.Score)
		assert.NotNil(t, item.Images)
	})

	t.Run("JSON unmarshaling uses correct field names", func(t *testing.T) {
		// Simulate API response with exact field names from spec
		apiResponse := `{
			"results": [
				{
					"title": "Example Result",
					"url": "https://example.com",
					"snippet": "This is a snippet",
					"date": "2024-01-15",
					"score": 0.89,
					"images": [
						{
							"url": "https://example.com/img.jpg",
							"width": 1024,
							"height": 768
						}
					]
				}
			]
		}`

		var resp SearchResponse
		err := json.Unmarshal([]byte(apiResponse), &resp)
		require.NoError(t, err)

		require.Len(t, resp.Results, 1)
		item := resp.Results[0]
		assert.Equal(t, "Example Result", item.Title)
		assert.Equal(t, "https://example.com", item.URL)
		assert.Equal(t, "This is a snippet", *item.Snippet)
		assert.Equal(t, "2024-01-15", *item.Date)
		assert.Equal(t, 0.89, *item.Score)
		require.NotNil(t, item.Images)
		require.Len(t, *item.Images, 1)
		assert.Equal(t, "https://example.com/img.jpg", (*item.Images)[0].URL)
		assert.Equal(t, 1024, *(*item.Images)[0].Width)
		assert.Equal(t, 768, *(*item.Images)[0].Height)
	})

	t.Run("optional fields can be omitted in response", func(t *testing.T) {
		// Test minimal response with only required fields
		apiResponse := `{
			"results": [
				{
					"title": "Minimal Result",
					"url": "https://example.com"
				}
			]
		}`

		var resp SearchResponse
		err := json.Unmarshal([]byte(apiResponse), &resp)
		require.NoError(t, err)

		require.Len(t, resp.Results, 1)
		item := resp.Results[0]
		assert.Equal(t, "Minimal Result", item.Title)
		assert.Equal(t, "https://example.com", item.URL)
		assert.Nil(t, item.Snippet)
		assert.Nil(t, item.Date)
		assert.Nil(t, item.Score)
		assert.Nil(t, item.Images)
	})
}

// TestSearchHTTPRequestCompliance verifies that HTTP requests are constructed
// according to the API specification.
func TestSearchHTTPRequestCompliance(t *testing.T) {
	var capturedRequest *http.Request
	var capturedBody []byte

	// Create a test server that captures the request
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedRequest = r

		// Read and store body
		body, _ := io.ReadAll(r.Body)
		capturedBody = body

		// Return valid response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"results": []}`))
	}))
	defer server.Close()

	client := NewClient("test-api-key")
	client.SetSearchEndpoint(server.URL)

	req := NewSearchRequest("test query", WithSearchMaxResults(10))
	_, err := client.SendSearchRequest(req)
	require.NoError(t, err)

	t.Run("HTTP method is POST", func(t *testing.T) {
		// Spec: POST https://api.perplexity.ai/search
		assert.Equal(t, http.MethodPost, capturedRequest.Method, "HTTP method must be POST")
	})

	t.Run("Authorization header includes Bearer token", func(t *testing.T) {
		// Spec: Authorization: Bearer {api_key}
		authHeader := capturedRequest.Header.Get("Authorization")
		assert.NotEmpty(t, authHeader, "Authorization header is required")
		assert.Contains(t, authHeader, "Bearer ", "Authorization must use Bearer scheme")
		assert.Contains(t, authHeader, "test-api-key", "Authorization must include API key")
		assert.Equal(t, "Bearer test-api-key", authHeader)
	})

	t.Run("Content-Type header is application/json", func(t *testing.T) {
		// Spec: Content-Type: application/json
		contentType := capturedRequest.Header.Get("Content-Type")
		assert.Equal(t, "application/json", contentType, "Content-Type must be application/json")
	})

	t.Run("integration header identifies perplexity-go", func(t *testing.T) {
		assert.Equal(t, "perplexity-go", capturedRequest.Header.Get("X-Pplx-Integration"))
	})

	t.Run("request body is valid JSON", func(t *testing.T) {
		var body map[string]interface{}
		err := json.Unmarshal(capturedBody, &body)
		assert.NoError(t, err, "Request body must be valid JSON")
		assert.Contains(t, body, "query", "Request body must contain query field")
	})
}

// TestSearchErrorResponseCompliance verifies error handling matches API specification.
func TestSearchErrorResponseCompliance(t *testing.T) {
	t.Run("401 Unauthorized returns ErrUnauthorized", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"error": "Invalid API key"}`))
		}))
		defer server.Close()

		client := NewClient("invalid-key")
		client.SetSearchEndpoint(server.URL)

		req := NewSearchRequest("test")
		_, err := client.SendSearchRequest(req)
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrUnauthorized)
	})

	t.Run("400 Bad Request parsed with ParseErrorMessage", func(t *testing.T) {
		// Use correct error response format: {"error": {"message": "...", "type": "...", "code": 123}}
		errorResponse := `{
			"error": {
				"message": "Invalid query parameter",
				"type": "BadRequest",
				"code": 400
			}
		}`

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(errorResponse))
		}))
		defer server.Close()

		client := NewClient("test-key")
		client.SetSearchEndpoint(server.URL)

		req := NewSearchRequest("test")
		_, err := client.SendSearchRequest(req)
		require.Error(t, err)
		// Verify ParseErrorMessage was used by checking error type
		var respErr *ResponseError
		assert.ErrorAs(t, err, &respErr)
		assert.Equal(t, "Invalid query parameter", respErr.Error())
	})

	t.Run("500 Internal Server Error parsed with ParseErrorMessage", func(t *testing.T) {
		// Use correct error response format
		errorResponse := `{
			"error": {
				"message": "Server encountered an error",
				"type": "InternalServerError",
				"code": 500
			}
		}`

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(errorResponse))
		}))
		defer server.Close()

		client := NewClient("test-key")
		client.SetSearchEndpoint(server.URL)

		req := NewSearchRequest("test")
		_, err := client.SendSearchRequest(req)
		require.Error(t, err)
		// Verify ParseErrorMessage was used
		var respErr *ResponseError
		assert.ErrorAs(t, err, &respErr)
		assert.Equal(t, "Server encountered an error", respErr.Error())
	})

	t.Run("non-JSON error response keeps body and status code", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadGateway)
			_, _ = w.Write([]byte("Not JSON"))
		}))
		defer server.Close()

		client := NewClient("test-key")
		client.SetSearchEndpoint(server.URL)

		_, err := client.SendSearchRequest(NewSearchRequest("test"))
		var respErr *ResponseError
		require.ErrorAs(t, err, &respErr)
		assert.Equal(t, http.StatusBadGateway, respErr.StatusCode)
		assert.Equal(t, "Not JSON", respErr.Error())
	})

	t.Run("422 validation error is parsed", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnprocessableEntity)
			_, _ = w.Write([]byte(`{"detail":[{"loc":["body","max_results"],"msg":"Input should be less than or equal to 20","type":"less_than_equal"}]}`))
		}))
		defer server.Close()

		client := NewClient("test-key")
		client.SetSearchEndpoint(server.URL)

		_, err := client.SendSearchRequest(NewSearchRequest("test"))
		var respErr *ResponseError
		require.ErrorAs(t, err, &respErr)
		assert.Equal(t, http.StatusUnprocessableEntity, respErr.StatusCode)
		require.Len(t, respErr.Detail, 1)
		assert.Equal(t, "less_than_equal", respErr.Detail[0].Type)
		assert.Equal(t, "validation error: body.max_results: Input should be less than or equal to 20", respErr.Error())
	})

	t.Run("429 rate limit exposes status code", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(`{"error":{"message":"Rate limit exceeded","type":"rate_limit","code":429}}`))
		}))
		defer server.Close()

		client := NewClient("test-key")
		client.SetSearchEndpoint(server.URL)

		_, err := client.SendSearchRequest(NewSearchRequest("test"))
		var respErr *ResponseError
		require.ErrorAs(t, err, &respErr)
		assert.Equal(t, http.StatusTooManyRequests, respErr.StatusCode)
		assert.Equal(t, "Rate limit exceeded", respErr.Error())
	})

	t.Run("invalid request is rejected before sending", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Fatal("Handler should not be called for an invalid request")
		}))
		defer server.Close()

		client := NewClient("test-key")
		client.SetSearchEndpoint(server.URL)

		_, err := client.SendSearchRequest(NewSearchRequest(""))
		assert.ErrorIs(t, err, ErrSearchQueryStringEmpty)

		_, err = client.SendSearchRequest(nil)
		assert.ErrorIs(t, err, ErrNilRequest)
	})
}

// TestSearchRequestWithContext verifies context support.
func TestSearchRequestWithContext(t *testing.T) {
	t.Run("SendSearchRequestWithContext method exists and works", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"results": []}`))
		}))
		defer server.Close()

		client := NewClient("test-key")
		client.SetSearchEndpoint(server.URL)

		// Verify the method exists and accepts context
		ctx := context.Background()
		req := NewSearchRequest("test")
		resp, err := client.SendSearchRequestWithContext(ctx, req)
		require.NoError(t, err)
		assert.NotNil(t, resp)
	})

	t.Run("cancelled context returns error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// This should not be reached
			t.Fatal("Handler should not be called with cancelled context")
		}))
		defer server.Close()

		client := NewClient("test-key")
		client.SetSearchEndpoint(server.URL)

		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		req := NewSearchRequest("test")
		_, err := client.SendSearchRequestWithContext(ctx, req)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "context canceled")
	})
}

// Helper functions for pointer types
func intPtr(i int) *int {
	return &i
}

func boolPtr(b bool) *bool {
	return &b
}

func strPtr(s string) *string {
	return &s
}

func floatPtr(f float64) *float64 {
	return &f
}
