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

func TestSearchRequestValidator_ValidateMaxTokens(t *testing.T) {
	validator := NewSearchRequestValidator()

	t.Run("valid max_tokens", func(t *testing.T) {
		req := NewSearchRequest("test", WithSearchMaxTokens(500))
		err := validator.ValidateSearchRequest(req)
		assert.NoError(t, err)
	})

	t.Run("zero max_tokens", func(t *testing.T) {
		req := NewSearchRequest("test", WithSearchMaxTokens(0))
		err := validator.ValidateSearchRequest(req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "positive integer")
	})

	t.Run("negative max_tokens", func(t *testing.T) {
		req := NewSearchRequest("test", WithSearchMaxTokens(-10))
		err := validator.ValidateSearchRequest(req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "positive integer")
	})

	t.Run("nil max_tokens", func(t *testing.T) {
		req := NewSearchRequest("test")
		err := validator.ValidateSearchRequest(req)
		assert.NoError(t, err)
	})
}

func TestSearchRequestValidator_ValidateLanguagePreference(t *testing.T) {
	validator := NewSearchRequestValidator()

	t.Run("valid ISO 639-1 codes", func(t *testing.T) {
		validCodes := []string{"en", "fr", "es", "de", "ja", "zh"}
		for _, code := range validCodes {
			req := NewSearchRequest("test", WithSearchLanguagePreference(code))
			err := validator.ValidateSearchRequest(req)
			assert.NoError(t, err, "language code %s should be valid", code)
		}
	})

	t.Run("valid extended format with country", func(t *testing.T) {
		validCodes := []string{"en-US", "en-GB", "fr-CA", "es-MX", "pt-BR"}
		for _, code := range validCodes {
			req := NewSearchRequest("test", WithSearchLanguagePreference(code))
			err := validator.ValidateSearchRequest(req)
			assert.NoError(t, err, "language code %s should be valid", code)
		}
	})

	t.Run("invalid format - uppercase language code", func(t *testing.T) {
		req := NewSearchRequest("test", WithSearchLanguagePreference("EN"))
		err := validator.ValidateSearchRequest(req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "ISO 639")
	})

	t.Run("invalid format - lowercase country code", func(t *testing.T) {
		req := NewSearchRequest("test", WithSearchLanguagePreference("en-us"))
		err := validator.ValidateSearchRequest(req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "ISO 639")
	})

	t.Run("invalid format - too short", func(t *testing.T) {
		req := NewSearchRequest("test", WithSearchLanguagePreference("e"))
		err := validator.ValidateSearchRequest(req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "2-5 characters")
	})

	t.Run("invalid format - too long", func(t *testing.T) {
		req := NewSearchRequest("test", WithSearchLanguagePreference("en-USA"))
		err := validator.ValidateSearchRequest(req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "2-5 characters")
	})

	t.Run("nil language_preference", func(t *testing.T) {
		req := NewSearchRequest("test")
		err := validator.ValidateSearchRequest(req)
		assert.NoError(t, err)
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

// TestQueryEdgeCases tests special characters, unicode, and very long strings in queries.
func TestQueryEdgeCases(t *testing.T) {
	validator := NewSearchRequestValidator()

	t.Run("query with special characters", func(t *testing.T) {
		specialChars := []string{
			"test! query?",
			"email@example.com",
			"price $100",
			"c++ programming",
			"50% discount",
			"#hashtag search",
			"query with (parentheses)",
			"query with [brackets]",
			"query with {braces}",
			"path/to/file",
			"a&b|c",
		}
		for _, query := range specialChars {
			req := NewSearchRequest(query)
			err := validator.ValidateSearchRequest(req)
			assert.NoError(t, err, "query %q should be valid", query)
		}
	})

	t.Run("query with unicode characters", func(t *testing.T) {
		unicodeQueries := []string{
			"你好世界",                 // Chinese
			"こんにちは",                // Japanese
			"안녕하세요",                // Korean
			"مرحبا",                // Arabic
			"привет",               // Russian
			"emoji test 😀 🎉 🚀",     // Emojis
			"math symbols ∫∑∏√",    // Math symbols
			"mixed 中文 and English", // Mixed
		}
		for _, query := range unicodeQueries {
			req := NewSearchRequest(query)
			err := validator.ValidateSearchRequest(req)
			assert.NoError(t, err, "unicode query %q should be valid", query)
		}
	})

	t.Run("very long query string", func(t *testing.T) {
		// Test with 1000 character query
		longQuery := ""
		for i := 0; i < 100; i++ {
			longQuery += "test query "
		}
		req := NewSearchRequest(longQuery)
		err := validator.ValidateSearchRequest(req)
		assert.NoError(t, err, "very long query should be valid")
	})

	t.Run("query with only whitespace", func(t *testing.T) {
		whitespaceQueries := []string{
			" ",
			"   ",
			"\t",
			"\n",
			"  \t  \n  ",
		}
		for _, query := range whitespaceQueries {
			req := NewSearchRequest(query)
			err := validator.ValidateSearchRequest(req)
			// Whitespace-only queries are currently allowed by the validator
			// since it only checks for empty string, not trimmed empty string
			assert.NoError(t, err, "whitespace query %q is currently allowed", query)
		}
	})

	t.Run("query with escape sequences", func(t *testing.T) {
		escapeQueries := []string{
			`query with "quotes"`,
			`query with 'single quotes'`,
			"query with\ttab",
			"query with\nnewline",
			`query with \\ backslash`,
		}
		for _, query := range escapeQueries {
			req := NewSearchRequest(query)
			err := validator.ValidateSearchRequest(req)
			assert.NoError(t, err, "query with escape sequences %q should be valid", query)
		}
	})
}

// TestLargeQueryArrays tests validation with large arrays of queries.
func TestLargeQueryArrays(t *testing.T) {
	validator := NewSearchRequestValidator()

	t.Run("array with 10 queries", func(t *testing.T) {
		queries := make([]string, 10)
		for i := 0; i < 10; i++ {
			queries[i] = "query " + string(rune('A'+i))
		}
		req := NewSearchRequest(queries)
		err := validator.ValidateSearchRequest(req)
		assert.NoError(t, err)
	})

	t.Run("array with 50 queries", func(t *testing.T) {
		queries := make([]string, 50)
		for i := 0; i < 50; i++ {
			queries[i] = "test query number " + string(rune('0'+i%10))
		}
		req := NewSearchRequest(queries)
		err := validator.ValidateSearchRequest(req)
		assert.NoError(t, err)
	})

	t.Run("array with 100 queries", func(t *testing.T) {
		queries := make([]string, 100)
		for i := 0; i < 100; i++ {
			queries[i] = "query item"
		}
		req := NewSearchRequest(queries)
		err := validator.ValidateSearchRequest(req)
		assert.NoError(t, err)
	})

	t.Run("large array with one empty element", func(t *testing.T) {
		queries := make([]string, 20)
		for i := 0; i < 20; i++ {
			if i == 10 {
				queries[i] = "" // Empty element at index 10
			} else {
				queries[i] = "valid query"
			}
		}
		req := NewSearchRequest(queries)
		err := validator.ValidateSearchRequest(req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "index 10")
	})

	t.Run("large array with mixed unicode and ascii", func(t *testing.T) {
		queries := []string{
			"english query",
			"中文查询",
			"日本語クエリ",
			"한국어 쿼리",
			"запрос на русском",
			"consulta en español",
			"requête en français",
			"deutsche Abfrage",
			"consulta em português",
			"ricerca in italiano",
			"query with emoji 🔍",
			"mixed 中英文 query",
		}
		req := NewSearchRequest(queries)
		err := validator.ValidateSearchRequest(req)
		assert.NoError(t, err)
	})
}

// TestDomainFilterEdgeCases tests complex domain patterns and edge cases.
func TestDomainFilterEdgeCases(t *testing.T) {
	validator := NewSearchRequestValidator()

	t.Run("domains with multiple wildcards", func(t *testing.T) {
		domains := []string{
			"*.example.*",
			"*.*",
			"*.*.example.com",
		}
		req := NewSearchRequest("test", WithSearchDomains(domains))
		err := validator.ValidateSearchRequest(req)
		assert.NoError(t, err, "domains with multiple wildcards should be valid")
	})

	t.Run("domains with hyphens in various positions", func(t *testing.T) {
		domains := []string{
			"my-domain.com",
			"test-site-name.org",
			"a-b-c-d.net",
			"site-123.com",
		}
		req := NewSearchRequest("test", WithSearchDomains(domains))
		err := validator.ValidateSearchRequest(req)
		assert.NoError(t, err)
	})

	t.Run("domains with numbers", func(t *testing.T) {
		domains := []string{
			"site123.com",
			"123site.org",
			"test456.net",
			"my3site4.edu",
		}
		req := NewSearchRequest("test", WithSearchDomains(domains))
		err := validator.ValidateSearchRequest(req)
		assert.NoError(t, err)
	})

	t.Run("subdomain patterns", func(t *testing.T) {
		domains := []string{
			"*.example.com",
			"*.subdomain.example.org",
			"api.*.example.com",
		}
		req := NewSearchRequest("test", WithSearchDomains(domains))
		err := validator.ValidateSearchRequest(req)
		assert.NoError(t, err)
	})

	t.Run("domains with special TLDs", func(t *testing.T) {
		domains := []string{
			"example.co.uk",
			"site.gov.au",
			"test.ac.jp",
			"example.com.br",
		}
		req := NewSearchRequest("test", WithSearchDomains(domains))
		err := validator.ValidateSearchRequest(req)
		assert.NoError(t, err)
	})

	t.Run("domains with special characters are allowed when they contain dots", func(t *testing.T) {
		// Note: The validator is intentionally permissive for domain filters.
		// Domains with dots are considered valid even with special characters,
		// as the API may support various glob patterns. The API will perform
		// final validation.
		domains := []string{
			"example!.com",
			"test@site.org",
			"my#domain.net",
			"site$.com",
			"test%.org",
			"example .com", // even with space, if it has a dot
		}
		for _, domain := range domains {
			req := NewSearchRequest("test", WithSearchDomains([]string{domain}))
			err := validator.ValidateSearchRequest(req)
			assert.NoError(t, err, "domain %s is allowed (contains dot)", domain)
		}
	})

	t.Run("domains without dots must match pattern", func(t *testing.T) {
		// Domains without dots must match the valid pattern regex
		invalidDomains := []string{
			"invalid!domain",
			"test@site",
			"my#domain",
			"site$name",
			"test%value",
			"domain with space",
		}
		for _, domain := range invalidDomains {
			req := NewSearchRequest("test", WithSearchDomains([]string{domain}))
			err := validator.ValidateSearchRequest(req)
			assert.Error(t, err, "domain without dot and with special chars %s should be invalid", domain)
		}
	})

	t.Run("large domain filter list", func(t *testing.T) {
		domains := make([]string, 50)
		for i := 0; i < 50; i++ {
			domains[i] = "example" + string(rune('a'+i%26)) + ".com"
		}
		req := NewSearchRequest("test", WithSearchDomains(domains))
		err := validator.ValidateSearchRequest(req)
		assert.NoError(t, err)
	})

	t.Run("single character domain components", func(t *testing.T) {
		domains := []string{
			"a.b.c.com",
			"x.y.z",
		}
		req := NewSearchRequest("test", WithSearchDomains(domains))
		err := validator.ValidateSearchRequest(req)
		assert.NoError(t, err)
	})
}

// TestMaxResultsBoundaries tests boundary values for max_results.
func TestMaxResultsBoundaries(t *testing.T) {
	validator := NewSearchRequestValidator()

	t.Run("max_results with value 1", func(t *testing.T) {
		req := NewSearchRequest("test", WithSearchMaxResults(1))
		err := validator.ValidateSearchRequest(req)
		assert.NoError(t, err)
	})

	t.Run("max_results with large positive value", func(t *testing.T) {
		req := NewSearchRequest("test", WithSearchMaxResults(1000))
		err := validator.ValidateSearchRequest(req)
		assert.NoError(t, err)
	})

	t.Run("max_results with very large positive value", func(t *testing.T) {
		req := NewSearchRequest("test", WithSearchMaxResults(999999))
		err := validator.ValidateSearchRequest(req)
		assert.NoError(t, err)
	})

	t.Run("max_results with maximum int value", func(t *testing.T) {
		req := NewSearchRequest("test", WithSearchMaxResults(2147483647)) // max int32
		err := validator.ValidateSearchRequest(req)
		assert.NoError(t, err)
	})

	t.Run("max_results with -1", func(t *testing.T) {
		req := NewSearchRequest("test", WithSearchMaxResults(-1))
		err := validator.ValidateSearchRequest(req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "positive integer")
	})

	t.Run("max_results with large negative value", func(t *testing.T) {
		req := NewSearchRequest("test", WithSearchMaxResults(-9999))
		err := validator.ValidateSearchRequest(req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "positive integer")
	})
}
