package perplexity

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSearchResponseUnmarshal(t *testing.T) {
	t.Run("complete response with all fields", func(t *testing.T) {
		jsonData := `{
			"results": [
				{
					"title": "Example Title",
					"url": "https://example.com",
					"snippet": "This is a snippet",
					"date": "2025-01-15",
					"score": 0.95,
					"images": [
						{
							"url": "https://example.com/image.jpg",
							"width": 800,
							"height": 600
						}
					]
				},
				{
					"title": "Another Title",
					"url": "https://test.org",
					"snippet": "Another snippet",
					"score": 0.87
				}
			]
		}`

		var resp SearchResponse
		err := json.Unmarshal([]byte(jsonData), &resp)
		require.NoError(t, err)

		assert.Len(t, resp.Results, 2)

		// First result
		result1 := resp.Results[0]
		assert.Equal(t, "Example Title", result1.Title)
		assert.Equal(t, "https://example.com", result1.URL)
		assert.Equal(t, "This is a snippet", *result1.Snippet)
		assert.Equal(t, "2025-01-15", *result1.Date)
		assert.Equal(t, 0.95, *result1.Score)
		require.NotNil(t, result1.Images)
		assert.Len(t, *result1.Images, 1)
		assert.Equal(t, "https://example.com/image.jpg", (*result1.Images)[0].URL)
		assert.Equal(t, 800, *(*result1.Images)[0].Width)
		assert.Equal(t, 600, *(*result1.Images)[0].Height)

		// Second result
		result2 := resp.Results[1]
		assert.Equal(t, "Another Title", result2.Title)
		assert.Equal(t, "https://test.org", result2.URL)
		assert.Equal(t, "Another snippet", *result2.Snippet)
		assert.Nil(t, result2.Date)
		assert.Nil(t, result2.Images)
	})

	t.Run("minimal response", func(t *testing.T) {
		jsonData := `{
			"results": [
				{
					"title": "Simple Title",
					"url": "https://simple.com"
				}
			]
		}`

		var resp SearchResponse
		err := json.Unmarshal([]byte(jsonData), &resp)
		require.NoError(t, err)

		assert.Len(t, resp.Results, 1)
		result := resp.Results[0]
		assert.Equal(t, "Simple Title", result.Title)
		assert.Equal(t, "https://simple.com", result.URL)
		assert.Nil(t, result.Snippet)
		assert.Nil(t, result.Date)
		assert.Nil(t, result.Score)
		assert.Nil(t, result.Images)
	})

	t.Run("empty results", func(t *testing.T) {
		jsonData := `{"results": []}`

		var resp SearchResponse
		err := json.Unmarshal([]byte(jsonData), &resp)
		require.NoError(t, err)

		assert.Len(t, resp.Results, 0)
	})
}

func TestSearchResponseGetResults(t *testing.T) {
	t.Run("with results", func(t *testing.T) {
		resp := &SearchResponse{
			Results: []SearchResultItem{
				{Title: "Test 1", URL: "https://test1.com"},
				{Title: "Test 2", URL: "https://test2.com"},
			},
		}
		results := resp.GetResults()
		assert.Len(t, results, 2)
	})

	t.Run("empty results", func(t *testing.T) {
		resp := &SearchResponse{Results: []SearchResultItem{}}
		results := resp.GetResults()
		assert.Len(t, results, 0)
	})

	t.Run("nil response", func(t *testing.T) {
		var resp *SearchResponse
		results := resp.GetResults()
		assert.Len(t, results, 0)
	})
}

func TestSearchResponseGetResultCount(t *testing.T) {
	t.Run("with results", func(t *testing.T) {
		resp := &SearchResponse{
			Results: []SearchResultItem{
				{Title: "Test 1", URL: "https://test1.com"},
				{Title: "Test 2", URL: "https://test2.com"},
				{Title: "Test 3", URL: "https://test3.com"},
			},
		}
		assert.Equal(t, 3, resp.GetResultCount())
	})

	t.Run("empty results", func(t *testing.T) {
		resp := &SearchResponse{Results: []SearchResultItem{}}
		assert.Equal(t, 0, resp.GetResultCount())
	})

	t.Run("nil response", func(t *testing.T) {
		var resp *SearchResponse
		assert.Equal(t, 0, resp.GetResultCount())
	})
}

func TestSearchResponseString(t *testing.T) {
	t.Run("valid response", func(t *testing.T) {
		snippet := "Test snippet"
		resp := &SearchResponse{
			Results: []SearchResultItem{
				{
					Title:   "Test Title",
					URL:     "https://test.com",
					Snippet: &snippet,
				},
			},
		}
		str := resp.String()
		assert.Contains(t, str, "Test Title")
		assert.Contains(t, str, "https://test.com")
		assert.Contains(t, str, "Test snippet")
	})

	t.Run("nil response", func(t *testing.T) {
		var resp *SearchResponse
		assert.Equal(t, "", resp.String())
	})

	t.Run("empty response", func(t *testing.T) {
		resp := &SearchResponse{}
		assert.Equal(t, "", resp.String())
	})
}

func TestSearchResultItemString(t *testing.T) {
	t.Run("complete item", func(t *testing.T) {
		date := "2025-01-15"
		score := 0.95
		item := &SearchResultItem{
			Title: "Example Title",
			URL:   "https://example.com",
			Date:  &date,
			Score: &score,
		}
		str := item.String()
		assert.Contains(t, str, "Example Title")
		assert.Contains(t, str, "https://example.com")
		assert.Contains(t, str, "2025-01-15")
		assert.Contains(t, str, "0.95")
	})

	t.Run("minimal item", func(t *testing.T) {
		item := &SearchResultItem{
			Title: "Simple Title",
			URL:   "https://simple.com",
		}
		str := item.String()
		assert.Equal(t, "Simple Title (https://simple.com)", str)
	})

	t.Run("nil item", func(t *testing.T) {
		var item *SearchResultItem
		assert.Equal(t, "", item.String())
	})
}

func TestSearchImageItemString(t *testing.T) {
	t.Run("complete image", func(t *testing.T) {
		width := 1920
		height := 1080
		img := &SearchImageItem{
			URL:    "https://example.com/image.jpg",
			Width:  &width,
			Height: &height,
		}
		str := img.String()
		assert.Contains(t, str, "https://example.com/image.jpg")
		assert.Contains(t, str, "1920x1080")
	})

	t.Run("minimal image", func(t *testing.T) {
		img := &SearchImageItem{
			URL: "https://example.com/image.jpg",
		}
		str := img.String()
		assert.Equal(t, "https://example.com/image.jpg", str)
	})

	t.Run("nil image", func(t *testing.T) {
		var img *SearchImageItem
		assert.Equal(t, "", img.String())
	})
}

func TestSearchResponseMarshal(t *testing.T) {
	t.Run("complete response", func(t *testing.T) {
		snippet := "Test snippet"
		date := "2025-01-15"
		score := 0.85
		width := 800
		height := 600

		resp := &SearchResponse{
			Results: []SearchResultItem{
				{
					Title:   "Test Title",
					URL:     "https://test.com",
					Snippet: &snippet,
					Date:    &date,
					Score:   &score,
					Images: &[]SearchImageItem{
						{
							URL:    "https://test.com/img.jpg",
							Width:  &width,
							Height: &height,
						},
					},
				},
			},
		}

		data, err := json.Marshal(resp)
		require.NoError(t, err)

		var result map[string]interface{}
		err = json.Unmarshal(data, &result)
		require.NoError(t, err)

		results, ok := result["results"].([]interface{})
		require.True(t, ok)
		assert.Len(t, results, 1)

		firstResult := results[0].(map[string]interface{})
		assert.Equal(t, "Test Title", firstResult["title"])
		assert.Equal(t, "https://test.com", firstResult["url"])
		assert.Equal(t, "Test snippet", firstResult["snippet"])
		assert.Equal(t, "2025-01-15", firstResult["date"])
		assert.Equal(t, 0.85, firstResult["score"])
	})
}
