package perplexity_test

import (
	"testing"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/sgaunet/perplexity-go/v2"
	"github.com/stretchr/testify/assert"
)

func TestValidate(t *testing.T) {
	validator := perplexity.NewRequestValidator()

	f := func(testName string, expectedValid bool, opts ...perplexity.CompletionRequestOption) {
		t.Helper()
		req := perplexity.NewCompletionRequest(opts...)
		err := validator.ValidateRequest(req)
		isEqual := assert.Equal(t, expectedValid, err == nil)
		if !isEqual {
			t.Logf("Test %s failed", testName)
		}
	}

	f("returns error if no message to send to the API", false)
	f("returns error if model is empty", false, perplexity.WithMessages([]perplexity.Message{{Role: "user", Content: "hello"}}), perplexity.WithModel(""))
	f("returns error if MaxTokens is negative", false, perplexity.WithMessages([]perplexity.Message{{Role: "user", Content: "hello"}}), perplexity.WithModel(perplexity.DefaultModel), perplexity.WithMaxTokens(-1))
	f("returns error if Temperature is negative", false, perplexity.WithMessages([]perplexity.Message{{Role: "user", Content: "hello"}}), perplexity.WithModel(perplexity.DefaultModel), perplexity.WithTemperature(-1))
	f("accepts Temperature value of 0", true, perplexity.WithMessages([]perplexity.Message{{Role: "user", Content: "hello"}}), perplexity.WithModel(perplexity.DefaultModel), perplexity.WithTemperature(0.0))
	f("returns error if TopP is negative", false, perplexity.WithMessages([]perplexity.Message{{Role: "user", Content: "hello"}}), perplexity.WithModel(perplexity.DefaultModel), perplexity.WithTopP(-1))
	f("returns error if TopK is negative", false, perplexity.WithMessages([]perplexity.Message{{Role: "user", Content: "hello"}}), perplexity.WithModel(perplexity.DefaultModel), perplexity.WithTopK(-1))
	f("returns error if TopK is gt 2048", false, perplexity.WithMessages([]perplexity.Message{{Role: "user", Content: "hello"}}), perplexity.WithModel(perplexity.DefaultModel), perplexity.WithTopK(2049))
	f("returns error if Temperature is gt 2", false, perplexity.WithMessages([]perplexity.Message{{Role: "user", Content: "hello"}}), perplexity.WithModel(perplexity.DefaultModel), perplexity.WithTemperature(2.1))
	f("returns error if Temperature is 2", false, perplexity.WithMessages([]perplexity.Message{{Role: "user", Content: "hello"}}), perplexity.WithModel(perplexity.DefaultModel), perplexity.WithTemperature(2.0))
	f("returns error if TopP is gt 1", false, perplexity.WithMessages([]perplexity.Message{{Role: "user", Content: "hello"}}), perplexity.WithModel(perplexity.DefaultModel), perplexity.WithTopP(1.1))
	f("returns error if SearchDomainFilter contains more than 10 elements", false, perplexity.WithMessages([]perplexity.Message{{Role: "user", Content: "hello"}}), perplexity.WithModel(perplexity.DefaultModel), perplexity.WithSearchDomainFilter([]string{"filter1", "filter2", "filter3", "filter4", "filter5", "filte6", "filter7", "filter8", "filter9", "filter10", "filter11"}))
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

func TestStructuredOutputValidation(t *testing.T) {
	validator := perplexity.NewRequestValidator()

	t.Run("validates JSON schema structured output with sonar model", func(t *testing.T) {
		msg := []perplexity.Message{
			{
				Role:    "user",
				Content: "hello",
			},
		}
		schema := map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"message": map[string]interface{}{
					"type": "string",
				},
			},
		}
		req := perplexity.NewCompletionRequest(
			perplexity.WithMessages(msg),
			perplexity.WithModel("sonar"),
			perplexity.WithJSONSchemaResponseFormat(schema),
		)
		err := validator.ValidateRequest(req)
		assert.NoError(t, err)
	})

	t.Run("validates regex structured output with sonar model", func(t *testing.T) {
		msg := []perplexity.Message{
			{
				Role:    "user",
				Content: "hello",
			},
		}
		regex := `\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}`
		req := perplexity.NewCompletionRequest(
			perplexity.WithMessages(msg),
			perplexity.WithModel("sonar"),
			perplexity.WithRegexResponseFormat(regex),
		)
		err := validator.ValidateRequest(req)
		assert.NoError(t, err)
	})

	t.Run("allows structured output with any model", func(t *testing.T) {
		msg := []perplexity.Message{
			{
				Role:    "user",
				Content: "hello",
			},
		}
		schema := map[string]interface{}{
			"type": "object",
		}
		req := perplexity.NewCompletionRequest(
			perplexity.WithMessages(msg),
			perplexity.WithModel("llama-3.1-sonar-small-128k-online"),
			perplexity.WithJSONSchemaResponseFormat(schema),
		)
		err := validator.ValidateRequest(req)
		assert.NoError(t, err)
	})

	t.Run("rejects invalid json_schema configuration", func(t *testing.T) {
		msg := []perplexity.Message{
			{
				Role:    "user",
				Content: "hello",
			},
		}
		responseFormat := &perplexity.ResponseFormat{
			Type: "json_schema",
			// Missing JSONSchema config
		}
		req := perplexity.NewCompletionRequest(
			perplexity.WithMessages(msg),
			perplexity.WithModel("sonar"),
			perplexity.WithResponseFormat(responseFormat),
		)
		err := validator.ValidateRequest(req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "response format type must match the provided configuration")
	})

	t.Run("rejects invalid regex configuration", func(t *testing.T) {
		msg := []perplexity.Message{
			{
				Role:    "user",
				Content: "hello",
			},
		}
		responseFormat := &perplexity.ResponseFormat{
			Type: "regex",
			// Missing Regex config
		}
		req := perplexity.NewCompletionRequest(
			perplexity.WithMessages(msg),
			perplexity.WithModel("sonar"),
			perplexity.WithResponseFormat(responseFormat),
		)
		err := validator.ValidateRequest(req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "response format type must match the provided configuration")
	})

	t.Run("rejects mixed configuration", func(t *testing.T) {
		msg := []perplexity.Message{
			{
				Role:    "user",
				Content: "hello",
			},
		}
		responseFormat := &perplexity.ResponseFormat{
			Type: "json_schema",
			JSONSchema: &perplexity.JSONSchemaConfig{
				Schema: map[string]interface{}{"type": "object"},
			},
			Regex: &perplexity.RegexConfig{
				Regex: `\d+`,
			},
		}
		req := perplexity.NewCompletionRequest(
			perplexity.WithMessages(msg),
			perplexity.WithModel("sonar"),
			perplexity.WithResponseFormat(responseFormat),
		)
		err := validator.ValidateRequest(req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "response format type must match the provided configuration")
	})

	t.Run("allows request without structured output", func(t *testing.T) {
		msg := []perplexity.Message{
			{
				Role:    "user",
				Content: "hello",
			},
		}
		req := perplexity.NewCompletionRequest(
			perplexity.WithMessages(msg),
			perplexity.WithModel("llama-3.1-sonar-small-128k-online"),
		)
		err := validator.ValidateRequest(req)
		assert.NoError(t, err)
	})
}

func TestSearchRecencyFilterValidationRule(t *testing.T) {
	validate := validator.New()
	tests := []struct {
		name  string
		value string
		valid bool
	}{
		{"empty value", "", true},
		{"year is valid", "year", true},
		{"month is valid", "month", true},
		{"week is valid", "week", true},
		{"day is valid", "day", true},
		{"hour is valid", "hour", true},
		{"invalid value foo", "foo", false},
		{"invalid value 2022", "2022", false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := &perplexity.CompletionRequest{
				Messages:            []perplexity.Message{{Role: "user", Content: "test"}},
				Model:               perplexity.DefaultModel,
				MaxTokens:           10,
				Temperature:         1.0,
				TopP:                0.5,
				SearchRecencyFilter: test.value,
				TopK:                10,
				PresencePenalty:     0.0,
				FrequencyPenalty:    1.0,
			}
			err := validate.Struct(req)
			if test.valid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestSearchModeValidationRule(t *testing.T) {
	validate := validator.New()
	tests := []struct {
		name  string
		value string
		valid bool
	}{
		{"empty value", "", true},
		{"academic is valid", "academic", true},
		{"web is valid", "web", true},
		{"invalid value foo", "foo", false},
		{"invalid value google", "google", false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := &perplexity.CompletionRequest{
				Messages:         []perplexity.Message{{Role: "user", Content: "test"}},
				Model:            perplexity.DefaultModel,
				MaxTokens:        10,
				Temperature:      1.0,
				TopP:             0.5,
				SearchMode:       test.value,
				TopK:             10,
				PresencePenalty:  0.0,
				FrequencyPenalty: 1.0,
			}
			err := validate.Struct(req)
			if test.valid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestValidateSearchDomainFilter(t *testing.T) {
	t.Run("returns no error for empty filter", func(t *testing.T) {
		req := perplexity.NewCompletionRequest(
			perplexity.WithMessages([]perplexity.Message{{Role: "user", Content: "test"}}),
			perplexity.WithModel(perplexity.DefaultModel),
		)
		err := req.ValidateSearchDomainFilter()
		assert.NoError(t, err)
	})

	t.Run("returns no error for valid filter length", func(t *testing.T) {
		req := perplexity.NewCompletionRequest(
			perplexity.WithMessages([]perplexity.Message{{Role: "user", Content: "test"}}),
			perplexity.WithModel(perplexity.DefaultModel),
			perplexity.WithSearchDomainFilter([]string{"example.com", "test.com"}),
		)
		err := req.ValidateSearchDomainFilter()
		assert.NoError(t, err)
	})

	t.Run("returns no error for maximum allowed length", func(t *testing.T) {
		req := perplexity.NewCompletionRequest(
			perplexity.WithMessages([]perplexity.Message{{Role: "user", Content: "test"}}),
			perplexity.WithModel(perplexity.DefaultModel),
			perplexity.WithSearchDomainFilter([]string{"domain1.com", "domain2.com", "domain3.com"}),
		)
		err := req.ValidateSearchDomainFilter()
		assert.NoError(t, err)
	})

	t.Run("returns error for exceeding maximum length", func(t *testing.T) {
		req := perplexity.NewCompletionRequest(
			perplexity.WithMessages([]perplexity.Message{{Role: "user", Content: "test"}}),
			perplexity.WithModel(perplexity.DefaultModel),
			perplexity.WithSearchDomainFilter([]string{"domain1.com", "domain2.com", "domain3.com", "domain4.com", "domain5.com", "domain6.com", "domain7.com", "domain8.com", "domain9.com", "domain10.com", "domain11.com"}),
		)
		err := req.ValidateSearchDomainFilter()
		assert.Error(t, err)
		assert.Equal(t, perplexity.ErrDomainFilterTooLong, err)
	})
}

func TestValidateSearchDomain(t *testing.T) {
	t.Run("returns no error if SearchDomain is empty", func(t *testing.T) {
		req := perplexity.NewCompletionRequest()
		err := req.ValidateSearchDomain()
		assert.NoError(t, err)
	})

	t.Run("returns no error if SearchDomain is set to 'sec'", func(t *testing.T) {
		req := perplexity.NewCompletionRequest(perplexity.WithSearchDomain("sec"))
		err := req.ValidateSearchDomain()
		assert.NoError(t, err)
	})

	t.Run("returns error if SearchDomain is set to invalid value", func(t *testing.T) {
		req := perplexity.NewCompletionRequest(perplexity.WithSearchDomain("invalid"))
		err := req.ValidateSearchDomain()
		assert.Error(t, err)
		assert.Equal(t, perplexity.ErrSearchDomainInvalid, err)
	})

	t.Run("returns error if SearchDomain is set to 'web'", func(t *testing.T) {
		req := perplexity.NewCompletionRequest(perplexity.WithSearchDomain("web"))
		err := req.ValidateSearchDomain()
		assert.Error(t, err)
		assert.Equal(t, perplexity.ErrSearchDomainInvalid, err)
	})
}

func TestValidateImageDomainFilter(t *testing.T) {
	validator := perplexity.NewRequestValidator()

	t.Run("returns no error for empty filter", func(t *testing.T) {
		req := perplexity.NewCompletionRequest(
			perplexity.WithMessages([]perplexity.Message{{Role: "user", Content: "test"}}),
			perplexity.WithModel(perplexity.DefaultModel),
		)
		// Test through full validation since individual method is not exported
		err := validator.ValidateRequest(req)
		assert.NoError(t, err)
	})

	t.Run("returns no error for valid domains", func(t *testing.T) {
		req := perplexity.NewCompletionRequest(
			perplexity.WithMessages([]perplexity.Message{{Role: "user", Content: "test"}}),
			perplexity.WithModel(perplexity.DefaultModel),
			perplexity.WithImageDomainFilter([]string{"example.com", "test.com"}),
		)
		err := validator.ValidateRequest(req)
		assert.NoError(t, err)
	})

	t.Run("returns no error for maximum allowed length", func(t *testing.T) {
		domains := []string{
			"domain1.com", "domain2.com", "domain3.com", "domain4.com", "domain5.com",
			"domain6.com", "domain7.com", "domain8.com", "domain9.com", "domain10.com",
		}
		req := perplexity.NewCompletionRequest(
			perplexity.WithMessages([]perplexity.Message{{Role: "user", Content: "test"}}),
			perplexity.WithModel(perplexity.DefaultModel),
			perplexity.WithImageDomainFilter(domains),
		)
		err := validator.ValidateRequest(req)
		assert.NoError(t, err)
	})

	t.Run("returns error for exceeding maximum length", func(t *testing.T) {
		domains := []string{
			"domain1.com", "domain2.com", "domain3.com", "domain4.com", "domain5.com",
			"domain6.com", "domain7.com", "domain8.com", "domain9.com", "domain10.com",
			"domain11.com",
		}
		req := perplexity.NewCompletionRequest(
			perplexity.WithMessages([]perplexity.Message{{Role: "user", Content: "test"}}),
			perplexity.WithModel(perplexity.DefaultModel),
			perplexity.WithImageDomainFilter(domains),
		)
		err := validator.ValidateRequest(req)
		assert.Error(t, err)
		assert.Equal(t, perplexity.ErrDomainFilterTooLong, err)
	})

	t.Run("returns error for empty domain entry", func(t *testing.T) {
		req := perplexity.NewCompletionRequest(
			perplexity.WithMessages([]perplexity.Message{{Role: "user", Content: "test"}}),
			perplexity.WithModel(perplexity.DefaultModel),
			perplexity.WithImageDomainFilter([]string{"example.com", "", "test.com"}),
		)
		err := validator.ValidateRequest(req)
		assert.Error(t, err)
		assert.Equal(t, perplexity.ErrDomainFilterEmpty, err)
	})

	t.Run("returns error for domain with protocol", func(t *testing.T) {
		req := perplexity.NewCompletionRequest(
			perplexity.WithMessages([]perplexity.Message{{Role: "user", Content: "test"}}),
			perplexity.WithModel(perplexity.DefaultModel),
			perplexity.WithImageDomainFilter([]string{"https://example.com"}),
		)
		err := validator.ValidateRequest(req)
		assert.Error(t, err)
		assert.Equal(t, perplexity.ErrDomainFilterProtocolNotAllowed, err)
	})

	t.Run("returns error for domain with http protocol", func(t *testing.T) {
		req := perplexity.NewCompletionRequest(
			perplexity.WithMessages([]perplexity.Message{{Role: "user", Content: "test"}}),
			perplexity.WithModel(perplexity.DefaultModel),
			perplexity.WithImageDomainFilter([]string{"http://example.com"}),
		)
		err := validator.ValidateRequest(req)
		assert.Error(t, err)
		assert.Equal(t, perplexity.ErrDomainFilterProtocolNotAllowed, err)
	})

	t.Run("returns error for domain with subdomain", func(t *testing.T) {
		req := perplexity.NewCompletionRequest(
			perplexity.WithMessages([]perplexity.Message{{Role: "user", Content: "test"}}),
			perplexity.WithModel(perplexity.DefaultModel),
			perplexity.WithImageDomainFilter([]string{"sub.example.com"}),
		)
		err := validator.ValidateRequest(req)
		assert.Error(t, err)
		assert.Equal(t, perplexity.ErrDomainFilterSubdomainNotAllowed, err)
	})

	t.Run("returns error for invalid domain format", func(t *testing.T) {
		req := perplexity.NewCompletionRequest(
			perplexity.WithMessages([]perplexity.Message{{Role: "user", Content: "test"}}),
			perplexity.WithModel(perplexity.DefaultModel),
			perplexity.WithImageDomainFilter([]string{"invalid@domain"}),
		)
		err := validator.ValidateRequest(req)
		assert.Error(t, err)
		assert.Equal(t, perplexity.ErrDomainFilterInvalidFormat, err)
	})

	t.Run("returns no error for exclusion prefix domain", func(t *testing.T) {
		req := perplexity.NewCompletionRequest(
			perplexity.WithMessages([]perplexity.Message{{Role: "user", Content: "test"}}),
			perplexity.WithModel(perplexity.DefaultModel),
			perplexity.WithImageDomainFilter([]string{"-gettyimages.com"}),
		)
		err := validator.ValidateRequest(req)
		assert.NoError(t, err)
	})

	t.Run("returns error for invalid exclusion prefix domain", func(t *testing.T) {
		req := perplexity.NewCompletionRequest(
			perplexity.WithMessages([]perplexity.Message{{Role: "user", Content: "test"}}),
			perplexity.WithModel(perplexity.DefaultModel),
			perplexity.WithImageDomainFilter([]string{"-invalid@domain"}),
		)
		err := validator.ValidateRequest(req)
		assert.Error(t, err)
		assert.Equal(t, perplexity.ErrDomainFilterInvalidFormat, err)
	})

	t.Run("returns error for domain with www prefix", func(t *testing.T) {
		req := perplexity.NewCompletionRequest(
			perplexity.WithMessages([]perplexity.Message{{Role: "user", Content: "test"}}),
			perplexity.WithModel(perplexity.DefaultModel),
			perplexity.WithImageDomainFilter([]string{"www.example.com"}),
		)
		err := req.Validate()
		assert.Error(t, err)
		assert.Equal(t, perplexity.ErrDomainFilterWWWNotAllowed, err)
	})

	t.Run("returns error for domain with www prefix and exclusion", func(t *testing.T) {
		req := perplexity.NewCompletionRequest(
			perplexity.WithMessages([]perplexity.Message{{Role: "user", Content: "test"}}),
			perplexity.WithModel(perplexity.DefaultModel),
			perplexity.WithImageDomainFilter([]string{"-www.example.com"}),
		)
		err := req.Validate()
		assert.Error(t, err)
		assert.Equal(t, perplexity.ErrDomainFilterWWWNotAllowed, err)
	})

	t.Run("returns error for domain with exclusion and http protocol", func(t *testing.T) {
		req := perplexity.NewCompletionRequest(
			perplexity.WithMessages([]perplexity.Message{{Role: "user", Content: "test"}}),
			perplexity.WithModel(perplexity.DefaultModel),
			perplexity.WithImageDomainFilter([]string{"-http://example.com"}),
		)
		err := req.Validate()
		assert.Error(t, err)
		assert.Equal(t, perplexity.ErrDomainFilterProtocolNotAllowed, err)
	})
}

func TestValidateImageFormatFilter(t *testing.T) {
	validator := perplexity.NewRequestValidator()

	t.Run("returns no error for empty filter", func(t *testing.T) {
		req := perplexity.NewCompletionRequest(
			perplexity.WithMessages([]perplexity.Message{{Role: "user", Content: "test"}}),
			perplexity.WithModel(perplexity.DefaultModel),
		)
		err := validator.ValidateRequest(req)
		assert.NoError(t, err)
	})

	t.Run("returns no error for valid formats", func(t *testing.T) {
		req := perplexity.NewCompletionRequest(
			perplexity.WithMessages([]perplexity.Message{{Role: "user", Content: "test"}}),
			perplexity.WithModel(perplexity.DefaultModel),
			perplexity.WithImageFormatFilter([]string{"jpg", "png", "webp"}),
		)
		err := validator.ValidateRequest(req)
		assert.NoError(t, err)
	})

	t.Run("returns error for empty format entry", func(t *testing.T) {
		req := perplexity.NewCompletionRequest(
			perplexity.WithMessages([]perplexity.Message{{Role: "user", Content: "test"}}),
			perplexity.WithModel(perplexity.DefaultModel),
			perplexity.WithImageFormatFilter([]string{"jpg", "", "png"}),
		)
		err := validator.ValidateRequest(req)
		assert.Error(t, err)
		assert.Equal(t, perplexity.ErrImageFormatFilterEmpty, err)
	})

	t.Run("returns error for format with dot prefix", func(t *testing.T) {
		req := perplexity.NewCompletionRequest(
			perplexity.WithMessages([]perplexity.Message{{Role: "user", Content: "test"}}),
			perplexity.WithModel(perplexity.DefaultModel),
			perplexity.WithImageFormatFilter([]string{".jpg"}),
		)
		err := validator.ValidateRequest(req)
		assert.Error(t, err)
		assert.Equal(t, perplexity.ErrImageFormatFilterDotPrefixNotAllowed, err)
	})

	t.Run("returns error for uppercase format", func(t *testing.T) {
		req := perplexity.NewCompletionRequest(
			perplexity.WithMessages([]perplexity.Message{{Role: "user", Content: "test"}}),
			perplexity.WithModel(perplexity.DefaultModel),
			perplexity.WithImageFormatFilter([]string{"JPG"}),
		)
		err := validator.ValidateRequest(req)
		assert.Error(t, err)
		assert.Equal(t, perplexity.ErrImageFormatFilterMustBeLowercase, err)
	})

	t.Run("returns error for format with special characters", func(t *testing.T) {
		req := perplexity.NewCompletionRequest(
			perplexity.WithMessages([]perplexity.Message{{Role: "user", Content: "test"}}),
			perplexity.WithModel(perplexity.DefaultModel),
			perplexity.WithImageFormatFilter([]string{"j-p-g"}),
		)
		err := validator.ValidateRequest(req)
		assert.Error(t, err)
		assert.Equal(t, perplexity.ErrImageFormatFilterInvalidFormat, err)
	})

	t.Run("returns error for format with numbers only", func(t *testing.T) {
		req := perplexity.NewCompletionRequest(
			perplexity.WithMessages([]perplexity.Message{{Role: "user", Content: "test"}}),
			perplexity.WithModel(perplexity.DefaultModel),
			perplexity.WithImageFormatFilter([]string{"j@g"}),
		)
		err := validator.ValidateRequest(req)
		assert.Error(t, err)
		assert.Equal(t, perplexity.ErrImageFormatFilterInvalidFormat, err)
	})

	t.Run("returns no error for valid alphanumeric format", func(t *testing.T) {
		req := perplexity.NewCompletionRequest(
			perplexity.WithMessages([]perplexity.Message{{Role: "user", Content: "test"}}),
			perplexity.WithModel(perplexity.DefaultModel),
			perplexity.WithImageFormatFilter([]string{"jpeg", "png", "webp", "gif", "bmp"}),
		)
		err := validator.ValidateRequest(req)
		assert.NoError(t, err)
	})

	t.Run("returns no error for format with numbers", func(t *testing.T) {
		req := perplexity.NewCompletionRequest(
			perplexity.WithMessages([]perplexity.Message{{Role: "user", Content: "test"}}),
			perplexity.WithModel(perplexity.DefaultModel),
			perplexity.WithImageFormatFilter([]string{"jpg", "png", "webp", "gif", "bmp", "tiff", "svg"}),
		)
		err := validator.ValidateRequest(req)
		assert.NoError(t, err)
	})
}

func TestValidateSearchRecencyFilterWithImages(t *testing.T) {
	t.Run("returns error when ReturnImages is true and SearchRecencyFilter is set", func(t *testing.T) {
		req := perplexity.NewCompletionRequest(
			perplexity.WithMessages([]perplexity.Message{{Role: "user", Content: "test"}}),
			perplexity.WithModel(perplexity.DefaultModel),
			perplexity.WithReturnImages(true),
			perplexity.WithSearchRecencyFilter("day"),
		)
		err := req.ValidateSearchRecencyFilter()
		assert.Error(t, err)
		assert.Equal(t, perplexity.ErrSearchRecencyFilter, err)
	})

	t.Run("returns no error when ReturnImages is false and SearchRecencyFilter is set", func(t *testing.T) {
		req := perplexity.NewCompletionRequest(
			perplexity.WithMessages([]perplexity.Message{{Role: "user", Content: "test"}}),
			perplexity.WithModel(perplexity.DefaultModel),
			perplexity.WithReturnImages(false),
			perplexity.WithSearchRecencyFilter("day"),
		)
		err := req.ValidateSearchRecencyFilter()
		assert.NoError(t, err)
	})

	t.Run("returns no error when ReturnImages is true and SearchRecencyFilter is empty", func(t *testing.T) {
		req := perplexity.NewCompletionRequest(
			perplexity.WithMessages([]perplexity.Message{{Role: "user", Content: "test"}}),
			perplexity.WithModel(perplexity.DefaultModel),
			perplexity.WithReturnImages(true),
			perplexity.WithSearchRecencyFilter(""),
		)
		err := req.ValidateSearchRecencyFilter()
		assert.NoError(t, err)
	})
}

func TestValidateRegexAndImagesCompatibility(t *testing.T) {
	validator := perplexity.NewRequestValidator()

	t.Run("returns error when regex and images are used together", func(t *testing.T) {
		req := perplexity.NewCompletionRequest(
			perplexity.WithMessages([]perplexity.Message{{Role: "user", Content: "test"}}),
			perplexity.WithModel("sonar"),
			perplexity.WithReturnImages(true),
			perplexity.WithSearchRecencyFilter(""),
			perplexity.WithRegexResponseFormat(`\d+`),
		)
		err := validator.ValidateRequest(req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "regex and images are not compatible")
	})

	t.Run("returns no error when regex is used without images", func(t *testing.T) {
		req := perplexity.NewCompletionRequest(
			perplexity.WithMessages([]perplexity.Message{{Role: "user", Content: "test"}}),
			perplexity.WithModel("sonar"),
			perplexity.WithReturnImages(false),
			perplexity.WithRegexResponseFormat(`\d+`),
		)
		err := validator.ValidateRequest(req)
		assert.NoError(t, err)
	})
}

// Test backward compatibility methods
func TestBackwardCompatibility(t *testing.T) {
	t.Run("Validate method still works", func(t *testing.T) {
		req := perplexity.NewCompletionRequest(
			perplexity.WithMessages([]perplexity.Message{{Role: "user", Content: "test"}}),
			perplexity.WithModel(perplexity.DefaultModel),
		)
		err := req.Validate()
		assert.NoError(t, err)
	})

	t.Run("ValidateSearchDomainFilter method still works", func(t *testing.T) {
		req := perplexity.NewCompletionRequest(
			perplexity.WithMessages([]perplexity.Message{{Role: "user", Content: "test"}}),
			perplexity.WithModel(perplexity.DefaultModel),
			perplexity.WithSearchDomainFilter([]string{"domain1.com", "domain2.com", "domain3.com", "domain4.com", "domain5.com", "domain6.com", "domain7.com", "domain8.com", "domain9.com", "domain10.com", "domain11.com"}),
		)
		err := req.ValidateSearchDomainFilter()
		assert.Error(t, err)
		assert.Equal(t, perplexity.ErrDomainFilterTooLong, err)
	})

	t.Run("ValidateSearchRecencyFilter method still works", func(t *testing.T) {
		req := perplexity.NewCompletionRequest(
			perplexity.WithMessages([]perplexity.Message{{Role: "user", Content: "test"}}),
			perplexity.WithModel(perplexity.DefaultModel),
			perplexity.WithReturnImages(true),
			perplexity.WithSearchRecencyFilter("day"),
		)
		err := req.ValidateSearchRecencyFilter()
		assert.Error(t, err)
		assert.Equal(t, perplexity.ErrSearchRecencyFilter, err)
	})

	t.Run("ValidateStructuredOutput method still works", func(t *testing.T) {
		req := perplexity.NewCompletionRequest(
			perplexity.WithMessages([]perplexity.Message{{Role: "user", Content: "test"}}),
			perplexity.WithModel("llama-3.1-sonar-small-128k-online"),
			perplexity.WithJSONSchemaResponseFormat(map[string]interface{}{"type": "object"}),
		)
		err := req.ValidateStructuredOutput()
		assert.NoError(t, err)
	})

	t.Run("ValidateSearchDomain method still works", func(t *testing.T) {
		req := perplexity.NewCompletionRequest(
			perplexity.WithMessages([]perplexity.Message{{Role: "user", Content: "test"}}),
			perplexity.WithModel(perplexity.DefaultModel),
			perplexity.WithSearchDomain("sec"),
		)
		err := req.ValidateSearchDomain()
		assert.NoError(t, err)
	})
}

func TestValidateRequestNil(t *testing.T) {
	validator := perplexity.NewRequestValidator()

	t.Run("returns error for nil request", func(t *testing.T) {
		err := validator.ValidateRequest(nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "request cannot be nil")
		assert.Contains(t, err.Error(), "validation failed")
	})
}

func TestValidateDateFilters(t *testing.T) {
	validator := perplexity.NewRequestValidator()

	t.Run("returns no error for valid date formats", func(t *testing.T) {
		req := perplexity.NewCompletionRequest(
			perplexity.WithMessages([]perplexity.Message{{Role: "user", Content: "test"}}),
			perplexity.WithModel(perplexity.DefaultModel),
		)
		// Manually set date filters to test validation
		req.SearchAfterDateFilter = "3/1/2025"
		req.SearchBeforeDateFilter = "12/31/2024"
		req.LastUpdatedAfterFilter = "1/15/2025"
		req.LastUpdatedBeforeFilter = "7/4/2024"

		err := validator.ValidateRequest(req)
		assert.NoError(t, err)
	})

	t.Run("returns no error for empty date filters", func(t *testing.T) {
		req := perplexity.NewCompletionRequest(
			perplexity.WithMessages([]perplexity.Message{{Role: "user", Content: "test"}}),
			perplexity.WithModel(perplexity.DefaultModel),
		)
		// All date filters are empty by default
		err := validator.ValidateRequest(req)
		assert.NoError(t, err)
	})

	t.Run("returns error for invalid date format with leading zeros", func(t *testing.T) {
		req := perplexity.NewCompletionRequest(
			perplexity.WithMessages([]perplexity.Message{{Role: "user", Content: "test"}}),
			perplexity.WithModel(perplexity.DefaultModel),
		)
		req.SearchAfterDateFilter = "03/01/2025" // Leading zero in month

		err := validator.ValidateRequest(req)
		assert.Error(t, err)
		assert.Equal(t, perplexity.ErrDateFilterInvalidFormat, err)
	})

	t.Run("returns error for ISO date format", func(t *testing.T) {
		req := perplexity.NewCompletionRequest(
			perplexity.WithMessages([]perplexity.Message{{Role: "user", Content: "test"}}),
			perplexity.WithModel(perplexity.DefaultModel),
		)
		req.SearchBeforeDateFilter = "2025-03-01" // ISO format

		err := validator.ValidateRequest(req)
		assert.Error(t, err)
		assert.Equal(t, perplexity.ErrDateFilterInvalidFormat, err)
	})

	t.Run("returns error for two-digit year", func(t *testing.T) {
		req := perplexity.NewCompletionRequest(
			perplexity.WithMessages([]perplexity.Message{{Role: "user", Content: "test"}}),
			perplexity.WithModel(perplexity.DefaultModel),
		)
		req.LastUpdatedAfterFilter = "3/1/25" // Two-digit year

		err := validator.ValidateRequest(req)
		assert.Error(t, err)
		assert.Equal(t, perplexity.ErrDateFilterInvalidFormat, err)
	})

	t.Run("returns error for invalid separator", func(t *testing.T) {
		req := perplexity.NewCompletionRequest(
			perplexity.WithMessages([]perplexity.Message{{Role: "user", Content: "test"}}),
			perplexity.WithModel(perplexity.DefaultModel),
		)
		req.LastUpdatedBeforeFilter = "3-1-2025" // Dashes instead of slashes

		err := validator.ValidateRequest(req)
		assert.Error(t, err)
		assert.Equal(t, perplexity.ErrDateFilterInvalidFormat, err)
	})

	t.Run("returns error for invalid month", func(t *testing.T) {
		req := perplexity.NewCompletionRequest(
			perplexity.WithMessages([]perplexity.Message{{Role: "user", Content: "test"}}),
			perplexity.WithModel(perplexity.DefaultModel),
		)
		req.SearchAfterDateFilter = "13/1/2025" // Month > 12

		err := validator.ValidateRequest(req)
		assert.Error(t, err)
		assert.Equal(t, perplexity.ErrDateFilterInvalidFormat, err)
	})

	t.Run("returns error for invalid day", func(t *testing.T) {
		req := perplexity.NewCompletionRequest(
			perplexity.WithMessages([]perplexity.Message{{Role: "user", Content: "test"}}),
			perplexity.WithModel(perplexity.DefaultModel),
		)
		req.SearchBeforeDateFilter = "3/32/2025" // Day > 31

		err := validator.ValidateRequest(req)
		assert.Error(t, err)
		assert.Equal(t, perplexity.ErrDateFilterInvalidFormat, err)
	})

	t.Run("returns no error for single-digit months and days", func(t *testing.T) {
		req := perplexity.NewCompletionRequest(
			perplexity.WithMessages([]perplexity.Message{{Role: "user", Content: "test"}}),
			perplexity.WithModel(perplexity.DefaultModel),
		)
		req.SearchAfterDateFilter = "1/1/2025"    // Single-digit month and day
		req.SearchBeforeDateFilter = "9/9/2024"   // Single-digit month and day
		req.LastUpdatedAfterFilter = "12/1/2025"  // Single-digit day
		req.LastUpdatedBeforeFilter = "1/31/2024" // Single-digit month

		err := validator.ValidateRequest(req)
		assert.NoError(t, err)
	})

	t.Run("functional options always produce valid formats", func(t *testing.T) {
		// Using functional options should always produce valid date formats
		req := perplexity.NewCompletionRequest(
			perplexity.WithMessages([]perplexity.Message{{Role: "user", Content: "test"}}),
			perplexity.WithModel(perplexity.DefaultModel),
			perplexity.WithSearchAfterDateFilter(time.Date(2025, 3, 1, 0, 0, 0, 0, time.UTC)),
			perplexity.WithSearchBeforeDateFilter(time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)),
			perplexity.WithLastUpdatedAfterFilter(time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)),
			perplexity.WithLastUpdatedBeforeFilter(time.Date(2024, 7, 4, 0, 0, 0, 0, time.UTC)),
		)

		err := validator.ValidateRequest(req)
		assert.NoError(t, err)
		// Deprecated functions now set the new API-compliant fields
		assert.Equal(t, "3/1/2025", req.PublishedAfter)
		assert.Equal(t, "12/31/2024", req.PublishedBefore)
		assert.Equal(t, "1/15/2025", req.LastUpdatedAfterFilter)
		assert.Equal(t, "7/4/2024", req.LastUpdatedBeforeFilter)
	})
}

func TestValidateReasoningEffort(t *testing.T) {
	validator := perplexity.NewRequestValidator()

	t.Run("returns no error when reasoning_effort is empty", func(t *testing.T) {
		req := perplexity.NewCompletionRequest(
			perplexity.WithMessages([]perplexity.Message{{Role: "user", Content: "test"}}),
			perplexity.WithModel(perplexity.DefaultModel),
		)
		err := validator.ValidateRequest(req)
		assert.NoError(t, err)
	})

	t.Run("returns no error when reasoning_effort is used with sonar-deep-research model", func(t *testing.T) {
		req := perplexity.NewCompletionRequest(
			perplexity.WithMessages([]perplexity.Message{{Role: "user", Content: "test"}}),
			perplexity.WithModel(perplexity.ModelSonarDeepResearch),
			perplexity.WithReasoningEffort(perplexity.ReasoningEffortLow),
		)
		err := validator.ValidateRequest(req)
		assert.NoError(t, err)

		// Test medium
		req = perplexity.NewCompletionRequest(
			perplexity.WithMessages([]perplexity.Message{{Role: "user", Content: "test"}}),
			perplexity.WithModel(perplexity.ModelSonarDeepResearch),
			perplexity.WithReasoningEffort(perplexity.ReasoningEffortMedium),
		)
		err = validator.ValidateRequest(req)
		assert.NoError(t, err)

		// Test high
		req = perplexity.NewCompletionRequest(
			perplexity.WithMessages([]perplexity.Message{{Role: "user", Content: "test"}}),
			perplexity.WithModel(perplexity.ModelSonarDeepResearch),
			perplexity.WithReasoningEffort(perplexity.ReasoningEffortHigh),
		)
		err = validator.ValidateRequest(req)
		assert.NoError(t, err)
	})

	t.Run("returns error when reasoning_effort is used with non-sonar-deep-research model", func(t *testing.T) {
		req := perplexity.NewCompletionRequest(
			perplexity.WithMessages([]perplexity.Message{{Role: "user", Content: "test"}}),
			perplexity.WithModel("sonar"),
			perplexity.WithReasoningEffort(perplexity.ReasoningEffortHigh),
		)
		err := validator.ValidateRequest(req)
		assert.Error(t, err)
		assert.Equal(t, perplexity.ErrReasoningEffortModelRequirement, err)
	})

	t.Run("returns error when reasoning_effort is used with other models", func(t *testing.T) {
		models := []string{
			"llama-3.1-sonar-small-128k-online",
			"llama-3.1-sonar-large-128k-online",
			"llama-3.1-sonar-huge-128k-online",
		}

		for _, model := range models {
			req := perplexity.NewCompletionRequest(
				perplexity.WithMessages([]perplexity.Message{{Role: "user", Content: "test"}}),
				perplexity.WithModel(model),
				perplexity.WithReasoningEffort(perplexity.ReasoningEffortMedium),
			)
			err := validator.ValidateRequest(req)
			assert.Error(t, err)
			assert.Equal(t, perplexity.ErrReasoningEffortModelRequirement, err)
		}
	})

	t.Run("returns error for invalid reasoning_effort value", func(t *testing.T) {
		req := perplexity.NewCompletionRequest(
			perplexity.WithMessages([]perplexity.Message{{Role: "user", Content: "test"}}),
			perplexity.WithModel(perplexity.ModelSonarDeepResearch),
		)
		// Manually set an invalid value to test validation
		req.ReasoningEffort = "ultra-high"

		err := validator.ValidateRequest(req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Key: 'CompletionRequest.ReasoningEffort'")
	})
}
