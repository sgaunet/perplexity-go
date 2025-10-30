package perplexity

import (
	"context"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// skipIfNoAPIKey skips the test if PPLX_API_KEY is not set.
func skipIfNoAPIKey(t *testing.T) string {
	apiKey := os.Getenv("PPLX_API_KEY")
	if apiKey == "" {
		t.Skip("Skipping integration test: PPLX_API_KEY environment variable not set")
	}
	return apiKey
}

// TestSearchIntegration_BasicSingleQuery tests a basic single query search with the real API.
func TestSearchIntegration_BasicSingleQuery(t *testing.T) {
	apiKey := skipIfNoAPIKey(t)

	client := NewClient(apiKey)
	req := NewSearchRequest("golang testing best practices")

	resp, err := client.SendSearchRequest(req)
	require.NoError(t, err, "Search request should succeed")
	require.NotNil(t, resp, "Response should not be nil")

	// Verify response structure
	assert.NotEmpty(t, resp.Results, "Should return at least one result")

	// Verify first result has required fields
	if len(resp.Results) > 0 {
		firstResult := resp.Results[0]
		assert.NotEmpty(t, firstResult.Title, "Result should have a title")
		assert.NotEmpty(t, firstResult.URL, "Result should have a URL")
		assert.Contains(t, firstResult.URL, "http", "URL should be a valid HTTP(S) URL")
	}
}

// TestSearchIntegration_MultiQuery tests multi-query search with the real API.
func TestSearchIntegration_MultiQuery(t *testing.T) {
	apiKey := skipIfNoAPIKey(t)

	client := NewClient(apiKey)
	queries := []string{
		"Go programming language",
		"Rust programming language",
		"Python programming language",
	}
	req := NewSearchRequest(queries)

	resp, err := client.SendSearchRequest(req)
	require.NoError(t, err, "Multi-query search should succeed")
	require.NotNil(t, resp, "Response should not be nil")

	// Multi-query searches should return results
	assert.NotEmpty(t, resp.Results, "Multi-query should return results")
}

// TestSearchIntegration_WithMaxResults tests max_results parameter.
func TestSearchIntegration_WithMaxResults(t *testing.T) {
	apiKey := skipIfNoAPIKey(t)

	client := NewClient(apiKey)
	maxResults := 5
	req := NewSearchRequest("artificial intelligence", WithSearchMaxResults(maxResults))

	resp, err := client.SendSearchRequest(req)
	require.NoError(t, err, "Search with max_results should succeed")
	require.NotNil(t, resp, "Response should not be nil")

	// Verify we don't get more than max_results
	assert.LessOrEqual(t, len(resp.Results), maxResults, "Should not exceed max_results")
}

// TestSearchIntegration_WithReturnImages tests return_images parameter.
func TestSearchIntegration_WithReturnImages(t *testing.T) {
	apiKey := skipIfNoAPIKey(t)

	client := NewClient(apiKey)
	req := NewSearchRequest(
		"beautiful landscapes photography",
		WithSearchReturnImages(true),
		WithSearchMaxResults(3),
	)

	resp, err := client.SendSearchRequest(req)
	require.NoError(t, err, "Search with return_images should succeed")
	require.NotNil(t, resp, "Response should not be nil")

	// Note: Images are not guaranteed even with return_images=true
	// Just verify the request succeeds
	assert.NotEmpty(t, resp.Results, "Should return results")
}

// TestSearchIntegration_WithReturnSnippets tests return_snippets parameter.
func TestSearchIntegration_WithReturnSnippets(t *testing.T) {
	apiKey := skipIfNoAPIKey(t)

	client := NewClient(apiKey)
	req := NewSearchRequest(
		"climate change solutions",
		WithSearchReturnSnippets(true),
		WithSearchMaxResults(3),
	)

	resp, err := client.SendSearchRequest(req)
	require.NoError(t, err, "Search with return_snippets should succeed")
	require.NotNil(t, resp, "Response should not be nil")
	assert.NotEmpty(t, resp.Results, "Should return results")

	// Check if any results have snippets
	hasSnippet := false
	for _, result := range resp.Results {
		if result.Snippet != nil && *result.Snippet != "" {
			hasSnippet = true
			break
		}
	}
	t.Logf("Results with snippets: %v", hasSnippet)
}

// TestSearchIntegration_WithCountry tests country parameter.
func TestSearchIntegration_WithCountry(t *testing.T) {
	apiKey := skipIfNoAPIKey(t)

	client := NewClient(apiKey)
	req := NewSearchRequest(
		"local news today",
		WithSearchCountry("US"),
		WithSearchMaxResults(3),
	)

	resp, err := client.SendSearchRequest(req)
	require.NoError(t, err, "Search with country should succeed")
	require.NotNil(t, resp, "Response should not be nil")
	assert.NotEmpty(t, resp.Results, "Should return results")
}

// TestSearchIntegration_WithDomainFilter tests domain filtering.
func TestSearchIntegration_WithDomainFilter(t *testing.T) {
	apiKey := skipIfNoAPIKey(t)

	client := NewClient(apiKey)
	// Use specific domains (API may not accept wildcard patterns like *.edu)
	req := NewSearchRequest(
		"Go programming tutorials",
		WithSearchDomains([]string{"github.com", "golang.org"}),
		WithSearchMaxResults(5),
	)

	resp, err := client.SendSearchRequest(req)
	require.NoError(t, err, "Search with domain filter should succeed")
	require.NotNil(t, resp, "Response should not be nil")

	// Results might be empty if no matching domains
	t.Logf("Found %d results with domain filter", len(resp.Results))
}

// TestSearchIntegration_AllOptions tests all options combined.
func TestSearchIntegration_AllOptions(t *testing.T) {
	apiKey := skipIfNoAPIKey(t)

	client := NewClient(apiKey)
	// Use specific domains without wildcards
	req := NewSearchRequest(
		"machine learning papers",
		WithSearchMaxResults(3),
		WithSearchReturnImages(false),
		WithSearchReturnSnippets(true),
		WithSearchCountry("US"),
		WithSearchDomains([]string{"arxiv.org", "github.com"}),
	)

	resp, err := client.SendSearchRequest(req)
	require.NoError(t, err, "Search with all options should succeed")
	require.NotNil(t, resp, "Response should not be nil")
	t.Logf("Returned %d results", len(resp.Results))
}

// TestSearchIntegration_ErrorUnauthorized tests 401 error with invalid key.
func TestSearchIntegration_ErrorUnauthorized(t *testing.T) {
	skipIfNoAPIKey(t) // Just to mark as integration test

	client := NewClient("invalid-api-key-12345")
	req := NewSearchRequest("test query")

	_, err := client.SendSearchRequest(req)
	require.Error(t, err, "Invalid API key should return error")
	assert.ErrorIs(t, err, ErrUnauthorized, "Should return ErrUnauthorized")
}

// TestSearchIntegration_ContextWithTimeout tests context timeout handling.
func TestSearchIntegration_ContextWithTimeout(t *testing.T) {
	apiKey := skipIfNoAPIKey(t)

	client := NewClient(apiKey)

	// Use very short timeout to trigger timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	req := NewSearchRequest("test query")
	_, err := client.SendSearchRequestWithContext(ctx, req)

	// Should get a context deadline exceeded or context canceled error
	require.Error(t, err, "Expired context should cause error")
	assert.Contains(t, err.Error(), "context", "Error should mention context")
}

// TestSearchIntegration_ContextCancellation tests context cancellation.
func TestSearchIntegration_ContextCancellation(t *testing.T) {
	apiKey := skipIfNoAPIKey(t)

	client := NewClient(apiKey)
	ctx, cancel := context.WithCancel(context.Background())

	// Cancel immediately
	cancel()

	req := NewSearchRequest("test query")
	_, err := client.SendSearchRequestWithContext(ctx, req)

	require.Error(t, err, "Cancelled context should cause error")
	assert.Contains(t, err.Error(), "context canceled", "Error should indicate context cancellation")
}

// TestSearchIntegration_ContextWithDeadline tests context with deadline.
func TestSearchIntegration_ContextWithDeadline(t *testing.T) {
	apiKey := skipIfNoAPIKey(t)

	client := NewClient(apiKey)

	// Set reasonable deadline (5 seconds)
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(5*time.Second))
	defer cancel()

	req := NewSearchRequest("golang best practices")
	resp, err := client.SendSearchRequestWithContext(ctx, req)

	// Should succeed within 5 seconds
	require.NoError(t, err, "Request should complete within deadline")
	require.NotNil(t, resp, "Response should not be nil")
}

// TestSearchIntegration_ConcurrentRequests tests concurrent API usage.
func TestSearchIntegration_ConcurrentRequests(t *testing.T) {
	apiKey := skipIfNoAPIKey(t)

	client := NewClient(apiKey)
	numRequests := 5

	var wg sync.WaitGroup
	errors := make(chan error, numRequests)
	responses := make(chan *SearchResponse, numRequests)

	// Launch concurrent requests
	for i := 0; i < numRequests; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			req := NewSearchRequest("test query", WithSearchMaxResults(1))
			resp, err := client.SendSearchRequest(req)

			if err != nil {
				errors <- err
			} else {
				responses <- resp
			}
		}(i)
	}

	wg.Wait()
	close(errors)
	close(responses)

	// Check results
	errorCount := 0
	successCount := 0

	for err := range errors {
		errorCount++
		t.Logf("Error in concurrent request: %v", err)
	}

	for range responses {
		successCount++
	}

	t.Logf("Concurrent requests: %d succeeded, %d failed", successCount, errorCount)

	// At least some requests should succeed
	assert.Greater(t, successCount, 0, "At least some concurrent requests should succeed")
}

// TestSearchIntegration_ThreadSafety tests thread safety of response objects.
func TestSearchIntegration_ThreadSafety(t *testing.T) {
	apiKey := skipIfNoAPIKey(t)

	client := NewClient(apiKey)
	req := NewSearchRequest("golang concurrency", WithSearchMaxResults(5))

	resp, err := client.SendSearchRequest(req)
	require.NoError(t, err, "Initial request should succeed")
	require.NotNil(t, resp, "Response should not be nil")

	if len(resp.Results) == 0 {
		t.Skip("No results to test thread safety")
	}

	// Access response from multiple goroutines
	var wg sync.WaitGroup
	numGoroutines := 10

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			// Read operations on response
			_ = resp.GetResults()
			_ = resp.GetResultCount()
			_ = resp.String()

			// Access individual results
			for _, result := range resp.Results {
				_ = result.String()
			}
		}()
	}

	wg.Wait()
	// If we reach here without panics, thread safety is good
	t.Log("Thread safety test completed without issues")
}

// TestSearchIntegration_ValidateRealResponseParsing verifies response parsing with real data.
func TestSearchIntegration_ValidateRealResponseParsing(t *testing.T) {
	apiKey := skipIfNoAPIKey(t)

	client := NewClient(apiKey)
	req := NewSearchRequest(
		"artificial intelligence trends 2024",
		WithSearchMaxResults(5),
		WithSearchReturnSnippets(true),
	)

	resp, err := client.SendSearchRequest(req)
	require.NoError(t, err, "Request should succeed")
	require.NotNil(t, resp, "Response should not be nil")
	require.NotEmpty(t, resp.Results, "Should have results")

	// Validate each result field
	for i, result := range resp.Results {
		assert.NotEmpty(t, result.Title, "Result %d should have title", i)
		assert.NotEmpty(t, result.URL, "Result %d should have URL", i)

		// Optional fields - just check if present, they're valid
		if result.Snippet != nil {
			t.Logf("Result %d has snippet: %s", i, *result.Snippet)
		}
		if result.Date != nil {
			t.Logf("Result %d has date: %s", i, *result.Date)
		}
		if result.Score != nil {
			assert.GreaterOrEqual(t, *result.Score, 0.0, "Score should be non-negative")
			t.Logf("Result %d has score: %.3f", i, *result.Score)
		}
		if result.Images != nil && len(*result.Images) > 0 {
			t.Logf("Result %d has %d images", i, len(*result.Images))
			for j, img := range *result.Images {
				assert.NotEmpty(t, img.URL, "Image %d should have URL", j)
			}
		}
	}
}

// TestSearchIntegration_EmptyQuery tests validation of empty query.
func TestSearchIntegration_EmptyQuery(t *testing.T) {
	apiKey := skipIfNoAPIKey(t)

	client := NewClient(apiKey)

	// Validator should catch this before sending to API
	validator := NewSearchRequestValidator()
	req := NewSearchRequest("")
	err := validator.ValidateSearchRequest(req)
	require.Error(t, err, "Empty query should fail validation")

	// Even if we bypass validation, API should reject
	_, err = client.SendSearchRequest(req)
	require.Error(t, err, "Empty query should be rejected")
}

// TestSearchIntegration_VeryLongQuery tests handling of very long queries.
func TestSearchIntegration_VeryLongQuery(t *testing.T) {
	apiKey := skipIfNoAPIKey(t)

	client := NewClient(apiKey)

	// Create a very long query (500+ characters)
	longQuery := "artificial intelligence machine learning deep learning neural networks " +
		"natural language processing computer vision robotics automation data science " +
		"big data analytics cloud computing edge computing quantum computing blockchain " +
		"cryptocurrency distributed systems microservices containerization kubernetes " +
		"devops continuous integration continuous deployment agile methodologies " +
		"software engineering best practices testing debugging performance optimization"

	req := NewSearchRequest(longQuery, WithSearchMaxResults(3))

	resp, err := client.SendSearchRequest(req)
	// API should either succeed or return a specific error
	if err != nil {
		t.Logf("Long query returned error: %v", err)
	} else {
		require.NotNil(t, resp, "Response should not be nil")
		t.Logf("Long query succeeded with %d results", len(resp.Results))
	}
}
