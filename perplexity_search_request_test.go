package perplexity

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSearchRequest(t *testing.T) {
	t.Run("with string query", func(t *testing.T) {
		req := NewSearchRequest("test query")
		assert.Equal(t, "test query", req.Query)
		assert.Nil(t, req.MaxResults)
		assert.Nil(t, req.ReturnImages)
		assert.Nil(t, req.ReturnSnippets)
		assert.Nil(t, req.Country)
		assert.Nil(t, req.SearchDomainFilter)
	})

	t.Run("with array query", func(t *testing.T) {
		queries := []string{"query1", "query2"}
		req := NewSearchRequest(queries)
		assert.Equal(t, queries, req.Query)
	})

	t.Run("with all options", func(t *testing.T) {
		maxResults := 10
		maxTokens := 500
		returnImages := true
		returnSnippets := false
		country := "US"
		domains := []string{"example.com", "test.org"}

		req := NewSearchRequest(
			"test query",
			WithSearchMaxResults(maxResults),
			WithSearchMaxTokens(maxTokens),
			WithSearchReturnImages(returnImages),
			WithSearchReturnSnippets(returnSnippets),
			WithSearchCountry(country),
			WithSearchDomains(domains),
		)

		assert.Equal(t, "test query", req.Query)
		assert.Equal(t, &maxResults, req.MaxResults)
		assert.Equal(t, &maxTokens, req.MaxTokens)
		assert.Equal(t, &returnImages, req.ReturnImages)
		assert.Equal(t, &returnSnippets, req.ReturnSnippets)
		assert.Equal(t, &country, req.Country)
		assert.Equal(t, &domains, req.SearchDomainFilter)
	})
}

func TestSearchRequestMarshal(t *testing.T) {
	t.Run("string query with options", func(t *testing.T) {
		maxResults := 5
		returnImages := true
		country := "FR"

		req := NewSearchRequest(
			"quantum computing",
			WithSearchMaxResults(maxResults),
			WithSearchReturnImages(returnImages),
			WithSearchCountry(country),
		)

		data, err := json.Marshal(req)
		require.NoError(t, err)

		var result map[string]interface{}
		err = json.Unmarshal(data, &result)
		require.NoError(t, err)

		assert.Equal(t, "quantum computing", result["query"])
		assert.Equal(t, float64(5), result["max_results"])
		assert.Equal(t, true, result["return_images"])
		assert.Equal(t, "FR", result["country"])
	})

	t.Run("array query", func(t *testing.T) {
		queries := []string{"AI research", "machine learning"}
		req := NewSearchRequest(queries)

		data, err := json.Marshal(req)
		require.NoError(t, err)

		var result map[string]interface{}
		err = json.Unmarshal(data, &result)
		require.NoError(t, err)

		queryArray, ok := result["query"].([]interface{})
		require.True(t, ok)
		assert.Len(t, queryArray, 2)
		assert.Equal(t, "AI research", queryArray[0])
		assert.Equal(t, "machine learning", queryArray[1])
	})

	t.Run("with domain filter", func(t *testing.T) {
		domains := []string{"*.edu", "arxiv.org"}
		req := NewSearchRequest(
			"scientific papers",
			WithSearchDomains(domains),
		)

		data, err := json.Marshal(req)
		require.NoError(t, err)

		var result map[string]interface{}
		err = json.Unmarshal(data, &result)
		require.NoError(t, err)

		domainArray, ok := result["search_domain_filter"].([]interface{})
		require.True(t, ok)
		assert.Len(t, domainArray, 2)
		assert.Equal(t, "*.edu", domainArray[0])
		assert.Equal(t, "arxiv.org", domainArray[1])
	})

	t.Run("with max_tokens", func(t *testing.T) {
		maxTokens := 750
		req := NewSearchRequest(
			"test query",
			WithSearchMaxTokens(maxTokens),
		)

		data, err := json.Marshal(req)
		require.NoError(t, err)

		var result map[string]interface{}
		err = json.Unmarshal(data, &result)
		require.NoError(t, err)

		assert.Equal(t, "test query", result["query"])
		assert.Equal(t, float64(750), result["max_tokens"])
	})

	t.Run("with language_preference", func(t *testing.T) {
		req := NewSearchRequest(
			"search query",
			WithSearchLanguagePreference("es"),
		)

		data, err := json.Marshal(req)
		require.NoError(t, err)

		var result map[string]interface{}
		err = json.Unmarshal(data, &result)
		require.NoError(t, err)

		assert.Equal(t, "search query", result["query"])
		assert.Equal(t, "es", result["language_preference"])
	})

	t.Run("minimal request", func(t *testing.T) {
		req := NewSearchRequest("simple query")

		data, err := json.Marshal(req)
		require.NoError(t, err)

		var result map[string]interface{}
		err = json.Unmarshal(data, &result)
		require.NoError(t, err)

		assert.Equal(t, "simple query", result["query"])
		assert.NotContains(t, result, "max_results")
		assert.NotContains(t, result, "max_tokens")
		assert.NotContains(t, result, "return_images")
		assert.NotContains(t, result, "return_snippets")
		assert.NotContains(t, result, "country")
		assert.NotContains(t, result, "search_domain_filter")
	})
}

func TestSearchRequestUnmarshal(t *testing.T) {
	t.Run("string query", func(t *testing.T) {
		jsonData := `{
			"query": "test search",
			"max_results": 10,
			"return_images": true,
			"country": "US"
		}`

		var req SearchRequest
		err := json.Unmarshal([]byte(jsonData), &req)
		require.NoError(t, err)

		assert.Equal(t, "test search", req.Query)
		assert.Equal(t, 10, *req.MaxResults)
		assert.Equal(t, true, *req.ReturnImages)
		assert.Equal(t, "US", *req.Country)
	})

	t.Run("array query", func(t *testing.T) {
		jsonData := `{
			"query": ["query1", "query2", "query3"]
		}`

		var req SearchRequest
		err := json.Unmarshal([]byte(jsonData), &req)
		require.NoError(t, err)

		queryArray, ok := req.Query.([]interface{})
		require.True(t, ok)
		assert.Len(t, queryArray, 3)
	})
}

func TestSearchRequestOptions(t *testing.T) {
	t.Run("WithSearchMaxResults", func(t *testing.T) {
		req := NewSearchRequest("test", WithSearchMaxResults(20))
		assert.Equal(t, 20, *req.MaxResults)
	})

	t.Run("WithSearchMaxTokens", func(t *testing.T) {
		req := NewSearchRequest("test", WithSearchMaxTokens(1000))
		assert.Equal(t, 1000, *req.MaxTokens)
	})

	t.Run("WithSearchReturnImages", func(t *testing.T) {
		req := NewSearchRequest("test", WithSearchReturnImages(true))
		assert.Equal(t, true, *req.ReturnImages)
	})

	t.Run("WithSearchReturnSnippets", func(t *testing.T) {
		req := NewSearchRequest("test", WithSearchReturnSnippets(false))
		assert.Equal(t, false, *req.ReturnSnippets)
	})

	t.Run("WithSearchCountry", func(t *testing.T) {
		req := NewSearchRequest("test", WithSearchCountry("GB"))
		assert.Equal(t, "GB", *req.Country)
	})

	t.Run("WithSearchLanguagePreference", func(t *testing.T) {
		req := NewSearchRequest("test", WithSearchLanguagePreference("fr"))
		assert.Equal(t, "fr", *req.LanguagePreference)
	})

	t.Run("WithSearchDomains", func(t *testing.T) {
		domains := []string{"example.com"}
		req := NewSearchRequest("test", WithSearchDomains(domains))
		assert.Equal(t, domains, *req.SearchDomainFilter)
	})
}
