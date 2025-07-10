package perplexity_test

import (
	"encoding/json"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/sgaunet/perplexity-go/v2"
	"github.com/stretchr/testify/assert"
)

func TestWithMessages(t *testing.T) {
	t.Run("creates a new CompletionRequest with messages", func(t *testing.T) {
		msg := []perplexity.Message{
			{
				Role:    "user",
				Content: "hello",
			},
		}
		req := perplexity.NewCompletionRequest(perplexity.WithMessages(msg))
		assert.Equal(t, req.Messages, msg)
	})
}

func TestWithModel(t *testing.T) {
	t.Run("creates a new CompletionRequest with model", func(t *testing.T) {
		model := perplexity.DefaultModel
		req := perplexity.NewCompletionRequest(perplexity.WithModel(model))
		assert.Equal(t, req.Model, model)
	})
	t.Run("Test WithDefaultModel", func(t *testing.T) {
		req := perplexity.NewCompletionRequest(perplexity.WithDefaultModel())
		assert.Equal(t, perplexity.DefaultModel, req.Model)
	})
}

func TestWithMaxTokens(t *testing.T) {
	t.Run("creates a new CompletionRequest with max tokens", func(t *testing.T) {
		maxTokens := 10
		req := perplexity.NewCompletionRequest(perplexity.WithMaxTokens(maxTokens))
		assert.Equal(t, req.MaxTokens, maxTokens)
	})
}

func TestWithTemperature(t *testing.T) {
	t.Run("creates a new CompletionRequest with temperature", func(t *testing.T) {
		temperature := 0.5
		req := perplexity.NewCompletionRequest(perplexity.WithTemperature(temperature))
		assert.Equal(t, req.Temperature, temperature)
	})
}

func TestWithTopP(t *testing.T) {
	t.Run("creates a new CompletionRequest with top p", func(t *testing.T) {
		topP := 0.5
		req := perplexity.NewCompletionRequest(perplexity.WithTopP(topP))
		assert.Equal(t, req.TopP, topP)
	})
}

func TestWithSearchDomainFilter(t *testing.T) {
	t.Run("creates a new CompletionRequest with search domain filter", func(t *testing.T) {
		searchDomainFilter := []string{"filter1", "filter2"}
		req := perplexity.NewCompletionRequest(perplexity.WithSearchDomainFilter(searchDomainFilter))
		assert.Equal(t, req.SearchDomainFilter, searchDomainFilter)
	})
}

func TestWithReturnImages(t *testing.T) {
	t.Run("creates a new CompletionRequest with return images", func(t *testing.T) {
		returnImages := true
		req := perplexity.NewCompletionRequest(perplexity.WithReturnImages(returnImages))
		assert.Equal(t, req.ReturnImages, returnImages)
	})
}

func TestWithReturnRelatedQuestions(t *testing.T) {
	t.Run("creates a new CompletionRequest with return related questions", func(t *testing.T) {
		returnRelatedQuestions := true
		req := perplexity.NewCompletionRequest(perplexity.WithReturnRelatedQuestions(returnRelatedQuestions))
		assert.Equal(t, req.ReturnRelatedQuestions, returnRelatedQuestions)
	})
}

func TestWithSearchRecencyFilter(t *testing.T) {
	t.Run("creates a new CompletionRequest with search recency filter", func(t *testing.T) {
		searchRecencyFilter := "filter"
		req := perplexity.NewCompletionRequest(perplexity.WithSearchRecencyFilter(searchRecencyFilter))
		assert.Equal(t, req.SearchRecencyFilter, searchRecencyFilter)
	})
}

func TestWebSearchOptionsValidation(t *testing.T) {
	validate := validator.New()
	t.Run("valid search_context_size values", func(t *testing.T) {
		for _, val := range []string{"low", "medium", "high", ""} {
			req := &perplexity.CompletionRequest{
				Messages:         []perplexity.Message{{Role: "user", Content: "test"}},
				Model:            perplexity.DefaultModel,
				MaxTokens:        10,
				Temperature:      1.0,
				TopP:             0.5,
				TopK:             10,
				PresencePenalty:  0.0,
				FrequencyPenalty: 1.0,
				WebSearchOptions: &perplexity.WebSearchOptions{SearchContextSize: val},
			}
			assert.NoError(t, validate.Struct(req))
		}
	})
	t.Run("invalid search_context_size value", func(t *testing.T) {
		req := &perplexity.CompletionRequest{
			Messages:         []perplexity.Message{{Role: "user", Content: "test"}},
			Model:            perplexity.DefaultModel,
			MaxTokens:        10,
			Temperature:      1.0,
			TopP:             0.5,
			TopK:             10,
			PresencePenalty:  0.0,
			FrequencyPenalty: 1.0,
			WebSearchOptions: &perplexity.WebSearchOptions{SearchContextSize: "super"},
		}
		assert.Error(t, validate.Struct(req))
	})
	t.Run("valid country code", func(t *testing.T) {
		req := &perplexity.CompletionRequest{
			Messages:         []perplexity.Message{{Role: "user", Content: "test"}},
			Model:            perplexity.DefaultModel,
			MaxTokens:        10,
			Temperature:      1.0,
			TopP:             0.5,
			TopK:             10,
			PresencePenalty:  0.0,
			FrequencyPenalty: 1.0,
			WebSearchOptions: &perplexity.WebSearchOptions{
				UserLocation: &perplexity.UserLocation{
					Latitude:  48.85,
					Longitude: 2.35,
					Country:   "FR",
				},
			},
		}
		assert.NoError(t, validate.Struct(req))
	})
	t.Run("invalid country code (too long)", func(t *testing.T) {
		req := &perplexity.CompletionRequest{
			Messages:         []perplexity.Message{{Role: "user", Content: "test"}},
			Model:            perplexity.DefaultModel,
			MaxTokens:        10,
			Temperature:      1.0,
			TopP:             0.5,
			TopK:             10,
			PresencePenalty:  0.0,
			FrequencyPenalty: 1.0,
			WebSearchOptions: &perplexity.WebSearchOptions{
				UserLocation: &perplexity.UserLocation{
					Latitude:  48.85,
					Longitude: 2.35,
					Country:   "FRA",
				},
			},
		}
		assert.Error(t, validate.Struct(req))
	})
	t.Run("omitempty: field omitted if not set", func(t *testing.T) {
		req := &perplexity.CompletionRequest{
			Messages:         []perplexity.Message{{Role: "user", Content: "test"}},
			Model:            perplexity.DefaultModel,
			MaxTokens:        10,
			Temperature:      1.0,
			TopP:             0.5,
			TopK:             10,
			PresencePenalty:  0.0,
			FrequencyPenalty: 1.0,
		}
		b, err := json.Marshal(req)
		assert.NoError(t, err)
		assert.NotContains(t, string(b), "web_search_options")
	})
	// Also check that if set, the field is present
	t.Run("web_search_options present when set", func(t *testing.T) {
		req := &perplexity.CompletionRequest{
			Messages:         []perplexity.Message{{Role: "user", Content: "test"}},
			Model:            perplexity.DefaultModel,
			MaxTokens:        10,
			Temperature:      1.0,
			TopP:             0.5,
			TopK:             10,
			PresencePenalty:  0.0,
			FrequencyPenalty: 1.0,
			WebSearchOptions: &perplexity.WebSearchOptions{SearchContextSize: "high"},
		}
		b, err := json.Marshal(req)
		assert.NoError(t, err)
		assert.Contains(t, string(b), "web_search_options")
		assert.Contains(t, string(b), "high")
	})
}

func TestWithTopK(t *testing.T) {
	t.Run("creates a new CompletionRequest with top k", func(t *testing.T) {
		topK := 10
		req := perplexity.NewCompletionRequest(perplexity.WithTopK(topK))
		assert.Equal(t, req.TopK, topK)
	})
}

func TestWithPresencePenalty(t *testing.T) {
	t.Run("creates a new CompletionRequest with presence penalty", func(t *testing.T) {
		presencePenalty := 0.5
		req := perplexity.NewCompletionRequest(perplexity.WithPresencePenalty(presencePenalty))
		assert.Equal(t, req.PresencePenalty, presencePenalty)
	})
}

func TestWithFrequencyPenalty(t *testing.T) {
	t.Run("creates a new CompletionRequest with frequency penalty", func(t *testing.T) {
		frequencyPenalty := 0.5
		req := perplexity.NewCompletionRequest(perplexity.WithFrequencyPenalty(frequencyPenalty))
		assert.Equal(t, req.FrequencyPenalty, frequencyPenalty)
	})
}

func TestWithSearchContextSize(t *testing.T) {
	tests := []struct {
		name     string
		size     string
		expected string
	}{
		{"low context size", "low", "low"},
		{"medium context size", "medium", "medium"},
		{"high context size", "high", "high"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := perplexity.NewCompletionRequest(perplexity.WithSearchContextSize(tt.size))
			assert.NotNil(t, req.WebSearchOptions)
			assert.Equal(t, tt.expected, req.WebSearchOptions.SearchContextSize)
		})
	}

	t.Run("initializes WebSearchOptions if nil", func(t *testing.T) {
		req := perplexity.NewCompletionRequest()
		req.WebSearchOptions = nil
		req = perplexity.NewCompletionRequest(perplexity.WithSearchContextSize("medium"))
		assert.NotNil(t, req.WebSearchOptions)
		assert.Equal(t, "medium", req.WebSearchOptions.SearchContextSize)
	})
}

func TestWithUserLocation(t *testing.T) {
	tests := []struct {
		name      string
		latitude  float64
		longitude float64
		country   string
	}{
		{
			name:      "US location",
			latitude:  37.7749,
			longitude: -122.4194,
			country:   "US",
		},
		{
			name:      "FR location",
			latitude:  48.8566,
			longitude: 2.3522,
			country:   "FR",
		},
		{
			name:      "JP location",
			latitude:  35.6762,
			longitude: 139.6503,
			country:   "JP",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := perplexity.NewCompletionRequest(
				perplexity.WithUserLocation(tt.latitude, tt.longitude, tt.country),
			)

			assert.NotNil(t, req.WebSearchOptions)
			assert.NotNil(t, req.WebSearchOptions.UserLocation)
			assert.Equal(t, tt.latitude, req.WebSearchOptions.UserLocation.Latitude)
			assert.Equal(t, tt.longitude, req.WebSearchOptions.UserLocation.Longitude)
			assert.Equal(t, tt.country, req.WebSearchOptions.UserLocation.Country)
		})
	}

	t.Run("initializes WebSearchOptions if nil", func(t *testing.T) {
		req := perplexity.NewCompletionRequest()
		req.WebSearchOptions = nil
		req = perplexity.NewCompletionRequest(
			perplexity.WithUserLocation(48.8566, 2.3522, "FR"),
		)

		assert.NotNil(t, req.WebSearchOptions)
		assert.NotNil(t, req.WebSearchOptions.UserLocation)
		assert.Equal(t, 48.8566, req.WebSearchOptions.UserLocation.Latitude)
		assert.Equal(t, 2.3522, req.WebSearchOptions.UserLocation.Longitude)
		assert.Equal(t, "FR", req.WebSearchOptions.UserLocation.Country)
	})

	t.Run("overwrites existing user location", func(t *testing.T) {
		req := perplexity.NewCompletionRequest(
			perplexity.WithUserLocation(40.7128, -74.0060, "US"),
		)

		assert.NotNil(t, req.WebSearchOptions)
		assert.NotNil(t, req.WebSearchOptions.UserLocation)
		assert.Equal(t, 40.7128, req.WebSearchOptions.UserLocation.Latitude)
		assert.Equal(t, -74.0060, req.WebSearchOptions.UserLocation.Longitude)
		assert.Equal(t, "US", req.WebSearchOptions.UserLocation.Country)
	})
}

func TestWithResponseFormat(t *testing.T) {
	t.Run("creates a new CompletionRequest with response format", func(t *testing.T) {
		responseFormat := &perplexity.ResponseFormat{
			Type: "json_schema",
			JSONSchema: &perplexity.JSONSchemaConfig{
				Schema: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"name": map[string]interface{}{
							"type": "string",
						},
					},
				},
			},
		}
		req := perplexity.NewCompletionRequest(perplexity.WithResponseFormat(responseFormat))
		assert.Equal(t, responseFormat, req.ResponseFormat)
	})
}

func TestWithJSONSchemaResponseFormat(t *testing.T) {
	t.Run("creates a new CompletionRequest with JSON schema response format", func(t *testing.T) {
		schema := map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"name": map[string]interface{}{
					"type": "string",
				},
				"age": map[string]interface{}{
					"type": "integer",
				},
			},
		}
		req := perplexity.NewCompletionRequest(perplexity.WithJSONSchemaResponseFormat(schema))
		assert.NotNil(t, req.ResponseFormat)
		assert.Equal(t, "json_schema", req.ResponseFormat.Type)
		assert.NotNil(t, req.ResponseFormat.JSONSchema)
		assert.Equal(t, schema, req.ResponseFormat.JSONSchema.Schema)
		assert.Nil(t, req.ResponseFormat.Regex)
	})
}

func TestWithRegexResponseFormat(t *testing.T) {
	t.Run("creates a new CompletionRequest with regex response format", func(t *testing.T) {
		regex := `\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}`
		req := perplexity.NewCompletionRequest(perplexity.WithRegexResponseFormat(regex))
		assert.NotNil(t, req.ResponseFormat)
		assert.Equal(t, "regex", req.ResponseFormat.Type)
		assert.NotNil(t, req.ResponseFormat.Regex)
		assert.Equal(t, regex, req.ResponseFormat.Regex.Regex)
		assert.Nil(t, req.ResponseFormat.JSONSchema)
	})
}

func TestStructuredOutputJSONSerialization(t *testing.T) {
	t.Run("serializes JSON schema response format correctly", func(t *testing.T) {
		schema := map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"name": map[string]interface{}{
					"type": "string",
				},
			},
		}
		responseFormat := &perplexity.ResponseFormat{
			Type: "json_schema",
			JSONSchema: &perplexity.JSONSchemaConfig{
				Schema: schema,
			},
		}

		data, err := json.Marshal(responseFormat)
		assert.NoError(t, err)

		var result map[string]interface{}
		err = json.Unmarshal(data, &result)
		assert.NoError(t, err)

		assert.Equal(t, "json_schema", result["type"])
		assert.NotNil(t, result["json_schema"])
		assert.Nil(t, result["regex"])

		jsonSchema := result["json_schema"].(map[string]interface{})
		assert.Equal(t, schema, jsonSchema["schema"])
	})

	t.Run("serializes regex response format correctly", func(t *testing.T) {
		regex := `\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}`
		responseFormat := &perplexity.ResponseFormat{
			Type: "regex",
			Regex: &perplexity.RegexConfig{
				Regex: regex,
			},
		}

		data, err := json.Marshal(responseFormat)
		assert.NoError(t, err)

		var result map[string]interface{}
		err = json.Unmarshal(data, &result)
		assert.NoError(t, err)

		assert.Equal(t, "regex", result["type"])
		assert.NotNil(t, result["regex"])
		assert.Nil(t, result["json_schema"])

		regexConfig := result["regex"].(map[string]interface{})
		assert.Equal(t, regex, regexConfig["regex"])
	})
}
