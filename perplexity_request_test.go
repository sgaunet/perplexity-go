package perplexity_test

import (
	"testing"

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

func TestWithTopK(t *testing.T) {
	t.Run("creates a new CompletionRequest with top k", func(t *testing.T) {
		topK := 10
		req := perplexity.NewCompletionRequest(perplexity.WithTopK(topK))
		assert.Equal(t, req.TopK, topK)
	})
}

func TestValidate(t *testing.T) {
	f := func(testName string, expectedValid bool, opts ...perplexity.CompletionRequestOption) {
		t.Helper()
		req := perplexity.NewCompletionRequest(opts...)
		err := req.Validate()
		isEqual := assert.Equal(t, expectedValid, err == nil)
		if !isEqual {
			t.Logf("Test %s failed", testName)
		}
	}

	f("returns error if no message to send to the API", false)
	f("returns error if model is empty", false, perplexity.WithMessages([]perplexity.Message{{Role: "user", Content: "hello"}}), perplexity.WithModel(""))
	f("returns error if MaxTokens is negative", false, perplexity.WithMessages([]perplexity.Message{{Role: "user", Content: "hello"}}), perplexity.WithModel(perplexity.DefaultModel), perplexity.WithMaxTokens(-1))
	f("returns error if Temperature is negative", false, perplexity.WithMessages([]perplexity.Message{{Role: "user", Content: "hello"}}), perplexity.WithModel(perplexity.DefaultModel), perplexity.WithTemperature(-1))
	f("returns error if TopP is negative", false, perplexity.WithMessages([]perplexity.Message{{Role: "user", Content: "hello"}}), perplexity.WithModel(perplexity.DefaultModel), perplexity.WithTopP(-1))
	f("returns error if TopK is negative", false, perplexity.WithMessages([]perplexity.Message{{Role: "user", Content: "hello"}}), perplexity.WithModel(perplexity.DefaultModel), perplexity.WithTopK(-1))
	f("returns error if TopK is gt 2048", false, perplexity.WithMessages([]perplexity.Message{{Role: "user", Content: "hello"}}), perplexity.WithModel(perplexity.DefaultModel), perplexity.WithTopK(2049))
	f("returns error if Temperature is gt 2", false, perplexity.WithMessages([]perplexity.Message{{Role: "user", Content: "hello"}}), perplexity.WithModel(perplexity.DefaultModel), perplexity.WithTemperature(2.1))
	f("returns error if TopP is gt 1", false, perplexity.WithMessages([]perplexity.Message{{Role: "user", Content: "hello"}}), perplexity.WithModel(perplexity.DefaultModel), perplexity.WithTopP(1.1))
	f("returns error if SearchDomainFilter contains more than 3 elements", false, perplexity.WithMessages([]perplexity.Message{{Role: "user", Content: "hello"}}), perplexity.WithModel(perplexity.DefaultModel), perplexity.WithSearchDomainFilter([]string{"filter1", "filter2", "filter3", "filter4"}))
	f("returns error return_images and searchRecencyFilter are set", false, perplexity.WithMessages([]perplexity.Message{{Role: "user", Content: "hello"}}), perplexity.WithModel(perplexity.DefaultModel), perplexity.WithMaxTokens(10), perplexity.WithTemperature(0.5), perplexity.WithTopP(0.5), perplexity.WithSearchDomainFilter([]string{"filter1", "filter2"}), perplexity.WithReturnImages(true), perplexity.WithReturnRelatedQuestions(true), perplexity.WithSearchRecencyFilter("filter"), perplexity.WithTopK(10))
	f("returns no error", true, perplexity.WithMessages([]perplexity.Message{{Role: "user", Content: "hello"}}), perplexity.WithModel(perplexity.DefaultModel), perplexity.WithMaxTokens(10), perplexity.WithTemperature(0.5), perplexity.WithTopP(0.5), perplexity.WithSearchDomainFilter([]string{"filter1", "filter2"}), perplexity.WithReturnRelatedQuestions(true), perplexity.WithTopK(10))
}

func TestValidateSearchRecencyFilter(t *testing.T) {
	f := func(testName string, expectedValid bool, opts ...perplexity.CompletionRequestOption) {
		t.Helper()
		req := perplexity.NewCompletionRequest(opts...)
		err := req.ValidateSearchRecencyFilter()
		isEqual := assert.Equal(t, expectedValid, err == nil)
		if !isEqual {
			t.Logf("Test %s failed", testName)
		}
	}

	f("returns no error if SearchRecencyFilter is empty", true)
	f("returns no error if SearchRecencyFilter is set to 'hour'", true, perplexity.WithSearchRecencyFilter("hour"))
	f("returns no error if SearchRecencyFilter is set to 'day'", true, perplexity.WithSearchRecencyFilter("day"))
	f("returns no error if SearchRecencyFilter is set to 'week'", true, perplexity.WithSearchRecencyFilter("week"))
	f("returns no error if SearchRecencyFilter is set to 'month'", true, perplexity.WithSearchRecencyFilter("month"))
	f("returns error if SearchRecencyFilter is set to 'year'", false, perplexity.WithSearchRecencyFilter("year"))
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
