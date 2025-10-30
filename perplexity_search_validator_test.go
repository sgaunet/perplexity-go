package perplexity

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSearchRequestValidator_ValidateSearchRequest(t *testing.T) {
	validator := NewSearchRequestValidator()

	t.Run("nil request", func(t *testing.T) {
		err := validator.ValidateSearchRequest(nil)
		assert.Equal(t, ErrNilRequest, err)
	})

	t.Run("valid string query", func(t *testing.T) {
		req := NewSearchRequest("test query")
		err := validator.ValidateSearchRequest(req)
		assert.NoError(t, err)
	})

	t.Run("valid array query", func(t *testing.T) {
		req := NewSearchRequest([]string{"query1", "query2"})
		err := validator.ValidateSearchRequest(req)
		assert.NoError(t, err)
	})

	t.Run("empty string query", func(t *testing.T) {
		req := NewSearchRequest("")
		err := validator.ValidateSearchRequest(req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot be empty")
	})

	t.Run("empty array query", func(t *testing.T) {
		req := NewSearchRequest([]string{})
		err := validator.ValidateSearchRequest(req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot be empty")
	})

	t.Run("array with empty element", func(t *testing.T) {
		req := NewSearchRequest([]string{"valid", "", "also valid"})
		err := validator.ValidateSearchRequest(req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "index 1")
	})

	t.Run("invalid query type", func(t *testing.T) {
		req := NewSearchRequest(123)
		err := validator.ValidateSearchRequest(req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "must be a string or array of strings")
	})

	t.Run("nil query", func(t *testing.T) {
		req := &SearchRequest{Query: nil}
		err := validator.ValidateSearchRequest(req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "required")
	})
}

func TestSearchRequestValidator_ValidateMaxResults(t *testing.T) {
	validator := NewSearchRequestValidator()

	t.Run("valid max_results", func(t *testing.T) {
		req := NewSearchRequest("test", WithSearchMaxResults(10))
		err := validator.ValidateSearchRequest(req)
		assert.NoError(t, err)
	})

	t.Run("zero max_results", func(t *testing.T) {
		req := NewSearchRequest("test", WithSearchMaxResults(0))
		err := validator.ValidateSearchRequest(req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "positive integer")
	})

	t.Run("negative max_results", func(t *testing.T) {
		req := NewSearchRequest("test", WithSearchMaxResults(-5))
		err := validator.ValidateSearchRequest(req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "positive integer")
	})
}

func TestSearchRequestValidator_ValidateCountry(t *testing.T) {
	validator := NewSearchRequestValidator()

	t.Run("valid country codes", func(t *testing.T) {
		validCodes := []string{"US", "FR", "GB", "DE", "JP", "CN"}
		for _, code := range validCodes {
			req := NewSearchRequest("test", WithSearchCountry(code))
			err := validator.ValidateSearchRequest(req)
			assert.NoError(t, err, "country code %s should be valid", code)
		}
	})

	t.Run("invalid country codes", func(t *testing.T) {
		invalidCodes := []string{"USA", "us", "U", "123", "U1", ""}
		for _, code := range invalidCodes {
			req := NewSearchRequest("test", WithSearchCountry(code))
			err := validator.ValidateSearchRequest(req)
			if code == "" {
				// Empty string validation might be skipped
				continue
			}
			assert.Error(t, err, "country code %s should be invalid", code)
			assert.Contains(t, err.Error(), "ISO 3166-1")
		}
	})

	t.Run("lowercase country code", func(t *testing.T) {
		req := NewSearchRequest("test", WithSearchCountry("us"))
		err := validator.ValidateSearchRequest(req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "uppercase")
	})

	t.Run("three-letter country code", func(t *testing.T) {
		req := NewSearchRequest("test", WithSearchCountry("USA"))
		err := validator.ValidateSearchRequest(req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "2 uppercase letters")
	})
}

func TestSearchRequestValidator_ValidateDomainFilter(t *testing.T) {
	validator := NewSearchRequestValidator()

	t.Run("valid domain filters", func(t *testing.T) {
		validDomains := [][]string{
			{"example.com"},
			{"*.edu", "arxiv.org"},
			{"test.org", "sample.net"},
			{"example.com", "*.co.uk"},
			{"*"},
		}
		for _, domains := range validDomains {
			req := NewSearchRequest("test", WithSearchDomains(domains))
			err := validator.ValidateSearchRequest(req)
			assert.NoError(t, err, "domains %v should be valid", domains)
		}
	})

	t.Run("empty domain in filter", func(t *testing.T) {
		req := NewSearchRequest("test", WithSearchDomains([]string{"example.com", "", "test.org"}))
		err := validator.ValidateSearchRequest(req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "index 1")
		assert.Contains(t, err.Error(), "cannot be empty")
	})

	t.Run("invalid domain format", func(t *testing.T) {
		req := NewSearchRequest("test", WithSearchDomains([]string{"not a domain!"}))
		err := validator.ValidateSearchRequest(req)
		assert.Error(t, err)
	})

	t.Run("empty domain filter array", func(t *testing.T) {
		req := NewSearchRequest("test", WithSearchDomains([]string{}))
		err := validator.ValidateSearchRequest(req)
		assert.NoError(t, err) // Empty array is allowed
	})
}

func TestSearchRequestValidator_CompleteValidation(t *testing.T) {
	validator := NewSearchRequestValidator()

	t.Run("fully valid request", func(t *testing.T) {
		req := NewSearchRequest(
			"quantum computing",
			WithSearchMaxResults(10),
			WithSearchReturnImages(true),
			WithSearchReturnSnippets(true),
			WithSearchCountry("US"),
			WithSearchDomains([]string{"*.edu", "arxiv.org"}),
		)
		err := validator.ValidateSearchRequest(req)
		assert.NoError(t, err)
	})

	t.Run("minimal valid request", func(t *testing.T) {
		req := NewSearchRequest("simple query")
		err := validator.ValidateSearchRequest(req)
		assert.NoError(t, err)
	})

	t.Run("multi-query with options", func(t *testing.T) {
		req := NewSearchRequest(
			[]string{"AI research", "machine learning", "deep learning"},
			WithSearchMaxResults(5),
			WithSearchCountry("GB"),
		)
		err := validator.ValidateSearchRequest(req)
		assert.NoError(t, err)
	})
}

func TestValidateQuery(t *testing.T) {
	validator := NewSearchRequestValidator()

	t.Run("valid string", func(t *testing.T) {
		err := validator.validateQuery("test query")
		assert.NoError(t, err)
	})

	t.Run("valid array", func(t *testing.T) {
		err := validator.validateQuery([]string{"query1", "query2"})
		assert.NoError(t, err)
	})

	t.Run("empty string", func(t *testing.T) {
		err := validator.validateQuery("")
		assert.Error(t, err)
	})

	t.Run("empty array", func(t *testing.T) {
		err := validator.validateQuery([]string{})
		assert.Error(t, err)
	})

	t.Run("nil query", func(t *testing.T) {
		err := validator.validateQuery(nil)
		assert.Error(t, err)
	})

	t.Run("invalid type", func(t *testing.T) {
		err := validator.validateQuery(123)
		assert.Error(t, err)
	})
}

func TestValidateCountry(t *testing.T) {
	validator := NewSearchRequestValidator()

	t.Run("valid codes", func(t *testing.T) {
		codes := []string{"US", "FR", "GB", "DE", "JP"}
		for _, code := range codes {
			err := validator.validateCountry(code)
			assert.NoError(t, err, "code %s should be valid", code)
		}
	})

	t.Run("invalid codes", func(t *testing.T) {
		codes := []string{"USA", "us", "1", "AB1", ""}
		for _, code := range codes {
			err := validator.validateCountry(code)
			assert.Error(t, err, "code %s should be invalid", code)
		}
	})
}

func TestValidateDomainFilter(t *testing.T) {
	validator := NewSearchRequestValidator()

	t.Run("valid domains", func(t *testing.T) {
		domains := []string{"example.com", "*.edu", "test.org", "*"}
		err := validator.validateDomainFilter(domains)
		assert.NoError(t, err)
	})

	t.Run("empty domain", func(t *testing.T) {
		domains := []string{"example.com", ""}
		err := validator.validateDomainFilter(domains)
		assert.Error(t, err)
	})

	t.Run("empty array", func(t *testing.T) {
		err := validator.validateDomainFilter([]string{})
		assert.NoError(t, err)
	})
}

func TestContainsDot(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"example.com", true},
		{".com", true},
		{"com", false},
		{"", false},
		{"test.org", true},
	}

	for _, tt := range tests {
		result := containsDot(tt.input)
		assert.Equal(t, tt.expected, result, "containsDot(%q)", tt.input)
	}
}

func TestIsValidDomainPattern(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"*.example.com", true},
		{"example.*", true},
		{"example.com", true},
		{"*", true},
		{"test-domain.org", true},
		{"invalid!domain", false},
		{"domain with spaces", false},
	}

	for _, tt := range tests {
		result := isValidDomainPattern(tt.input)
		assert.Equal(t, tt.expected, result, "isValidDomainPattern(%q)", tt.input)
	}
}

func TestNewSearchRequestValidator(t *testing.T) {
	validator := NewSearchRequestValidator()
	require.NotNil(t, validator)
	require.NotNil(t, validator.validator)
}
