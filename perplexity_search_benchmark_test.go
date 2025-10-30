package perplexity

import (
	"encoding/json"
	"testing"
)

// Benchmark for request marshaling with single query
func BenchmarkSearchRequestMarshalSingleQuery(b *testing.B) {
	req := NewSearchRequest("golang best practices")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := json.Marshal(req)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Benchmark for request marshaling with multiple queries
func BenchmarkSearchRequestMarshalMultiQuery(b *testing.B) {
	queries := []string{
		"Go programming language",
		"Rust programming language",
		"Python programming language",
		"JavaScript programming language",
		"TypeScript programming language",
	}
	req := NewSearchRequest(queries)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := json.Marshal(req)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Benchmark for request marshaling with all options
func BenchmarkSearchRequestMarshalWithOptions(b *testing.B) {
	req := NewSearchRequest(
		"machine learning papers",
		WithSearchMaxResults(10),
		WithSearchReturnImages(true),
		WithSearchReturnSnippets(true),
		WithSearchCountry("US"),
		WithSearchDomains([]string{"arxiv.org", "github.com", "python.org"}),
	)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := json.Marshal(req)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Benchmark for response unmarshaling with minimal data
func BenchmarkSearchResponseUnmarshalMinimal(b *testing.B) {
	data := []byte(`{
		"results": [
			{
				"title": "Test Result",
				"url": "https://example.com/test"
			}
		]
	}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var resp SearchResponse
		err := json.Unmarshal(data, &resp)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Benchmark for response unmarshaling with full data
func BenchmarkSearchResponseUnmarshalFull(b *testing.B) {
	data := []byte(`{
		"results": [
			{
				"title": "Test Result 1",
				"url": "https://example.com/test1",
				"snippet": "This is a test snippet with some content that describes the result",
				"date": "2025-01-15",
				"score": 0.95,
				"images": [
					{
						"url": "https://example.com/image1.jpg",
						"width": 800,
						"height": 600
					},
					{
						"url": "https://example.com/image2.jpg",
						"width": 1024,
						"height": 768
					}
				]
			},
			{
				"title": "Test Result 2",
				"url": "https://example.com/test2",
				"snippet": "Another test snippet with different content",
				"date": "2025-01-14",
				"score": 0.87
			}
		]
	}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var resp SearchResponse
		err := json.Unmarshal(data, &resp)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Benchmark for response unmarshaling with large result set
func BenchmarkSearchResponseUnmarshalLarge(b *testing.B) {
	// Build large response with 50 results
	results := make([]map[string]interface{}, 50)
	for i := 0; i < 50; i++ {
		results[i] = map[string]interface{}{
			"title":   "Test Result",
			"url":     "https://example.com/test",
			"snippet": "This is a test snippet with some content that describes the result in detail",
			"date":    "2025-01-15",
			"score":   0.95,
		}
	}

	responseData := map[string]interface{}{
		"results": results,
	}
	data, err := json.Marshal(responseData)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var resp SearchResponse
		err := json.Unmarshal(data, &resp)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Benchmark for validation of simple request
func BenchmarkSearchRequestValidationSimple(b *testing.B) {
	validator := NewSearchRequestValidator()
	req := NewSearchRequest("golang best practices")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := validator.ValidateSearchRequest(req)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Benchmark for validation of complex request
func BenchmarkSearchRequestValidationComplex(b *testing.B) {
	validator := NewSearchRequestValidator()
	queries := []string{
		"Go programming language",
		"Rust programming language",
		"Python programming language",
		"JavaScript programming language",
		"TypeScript programming language",
	}
	req := NewSearchRequest(
		queries,
		WithSearchMaxResults(10),
		WithSearchReturnImages(true),
		WithSearchReturnSnippets(true),
		WithSearchCountry("US"),
		WithSearchDomains([]string{"github.com", "golang.org", "rust-lang.org", "python.org", "typescriptlang.org"}),
	)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := validator.ValidateSearchRequest(req)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Benchmark for GetResults helper method
func BenchmarkSearchResponseGetResults(b *testing.B) {
	resp := &SearchResponse{
		Results: []SearchResultItem{
			{Title: "Result 1", URL: "https://example.com/1"},
			{Title: "Result 2", URL: "https://example.com/2"},
			{Title: "Result 3", URL: "https://example.com/3"},
			{Title: "Result 4", URL: "https://example.com/4"},
			{Title: "Result 5", URL: "https://example.com/5"},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = resp.GetResults()
	}
}

// Benchmark for GetResultCount helper method
func BenchmarkSearchResponseGetResultCount(b *testing.B) {
	resp := &SearchResponse{
		Results: make([]SearchResultItem, 50),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = resp.GetResultCount()
	}
}

// Benchmark for String() method on SearchResultItem
func BenchmarkSearchResultItemString(b *testing.B) {
	item := &SearchResultItem{
		Title: "Test Result with a moderately long title",
		URL:   "https://example.com/very/long/path/to/resource/with/multiple/segments",
	}
	date := "2025-01-15"
	score := 0.95
	item.Date = &date
	item.Score = &score

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = item.String()
	}
}

// Benchmark for String() method on SearchResponse
func BenchmarkSearchResponseString(b *testing.B) {
	resp := &SearchResponse{
		Results: []SearchResultItem{
			{Title: "Result 1", URL: "https://example.com/1"},
			{Title: "Result 2", URL: "https://example.com/2"},
			{Title: "Result 3", URL: "https://example.com/3"},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = resp.String()
	}
}
