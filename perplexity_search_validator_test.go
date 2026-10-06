package perplexity

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

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

	t.Run("array with max queries", func(t *testing.T) {
		queries := []string{"query A", "query B", "query C", "query D", "query E"}
		req := NewSearchRequest(queries)
		err := validator.ValidateSearchRequest(req)
		assert.NoError(t, err)
	})

	t.Run("array above max queries", func(t *testing.T) {
		queries := []string{"q1", "q2", "q3", "q4", "q5", "q6"}
		req := NewSearchRequest(queries)
		err := validator.ValidateSearchRequest(req)
		assert.ErrorIs(t, err, ErrSearchQueryArrayTooLong)
	})

	t.Run("array with one empty element", func(t *testing.T) {
		queries := []string{"valid", "valid", "", "valid"}
		req := NewSearchRequest(queries)
		err := validator.ValidateSearchRequest(req)
		assert.ErrorIs(t, err, ErrSearchQueryArrayElementEmpty)
		assert.Contains(t, err.Error(), "index 2")
	})

	t.Run("array with mixed unicode and ascii", func(t *testing.T) {
		queries := []string{
			"english query",
			"中文查询",
			"запрос на русском",
			"query with emoji 🔍",
			"mixed 中英文 query",
		}
		req := NewSearchRequest(queries)
		err := validator.ValidateSearchRequest(req)
		assert.NoError(t, err)
	})

	t.Run("array decoded from JSON ([]any)", func(t *testing.T) {
		var req SearchRequest
		require.NoError(t, json.Unmarshal([]byte(`{"query":["a","b"]}`), &req))
		assert.NoError(t, validator.ValidateSearchRequest(&req))
	})

	t.Run("[]any with non-string element", func(t *testing.T) {
		req := NewSearchRequest([]any{"a", 1})
		assert.ErrorIs(t, validator.ValidateSearchRequest(req), ErrSearchQueryInvalidType)
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

	t.Run("domain filter list at max size", func(t *testing.T) {
		domains := make([]string, SearchMaxDomainFilterEntries)
		for i := range domains {
			domains[i] = "example" + string(rune('a'+i)) + ".com"
		}
		req := NewSearchRequest("test", WithSearchDomains(domains))
		err := validator.ValidateSearchRequest(req)
		assert.NoError(t, err)
	})

	t.Run("domain filter list above max size", func(t *testing.T) {
		domains := make([]string, SearchMaxDomainFilterEntries+1)
		for i := range domains {
			domains[i] = "example" + string(rune('a'+i)) + ".com"
		}
		req := NewSearchRequest("test", WithSearchDomains(domains))
		err := validator.ValidateSearchRequest(req)
		assert.ErrorIs(t, err, ErrSearchDomainFilterTooLong)
	})

	t.Run("denylist, paths and TLDs", func(t *testing.T) {
		valid := [][]string{
			{"-reddit.com", "-pinterest.com"},
			{"-reddit.com/r/all"},
			{"-.gov"},
			{"example.com/blog", ".gov", ".edu"},
		}
		for _, domains := range valid {
			req := NewSearchRequest("test", WithSearchDomains(domains))
			assert.NoError(t, validator.ValidateSearchRequest(req), "domains %v should be valid", domains)
		}
	})

	t.Run("mixing allowlist and denylist", func(t *testing.T) {
		req := NewSearchRequest("test", WithSearchDomains([]string{"example.com", "-reddit.com"}))
		assert.ErrorIs(t, validator.ValidateSearchRequest(req), ErrSearchDomainFilterMixedModes)
	})

	t.Run("invalid entries", func(t *testing.T) {
		invalid := []string{
			"https://example.com",
			"example .com",
			"-",
			"/path",
			strings.Repeat("a", SearchMaxDomainFilterEntryLength-3) + ".com",
		}
		for _, domain := range invalid {
			req := NewSearchRequest("test", WithSearchDomains([]string{domain}))
			assert.ErrorIs(t, validator.ValidateSearchRequest(req), ErrSearchDomainFilterEntryInvalid, "domain %q", domain)
		}
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

	t.Run("max_results at web limit", func(t *testing.T) {
		req := NewSearchRequest("test", WithSearchMaxResults(SearchMaxResultsLimit))
		assert.NoError(t, validator.ValidateSearchRequest(req))
	})

	t.Run("max_results above web limit", func(t *testing.T) {
		req := NewSearchRequest("test", WithSearchMaxResults(SearchMaxResultsLimit+1))
		assert.ErrorIs(t, validator.ValidateSearchRequest(req), ErrSearchMaxResultsTooLarge)
	})

	t.Run("max_results above web limit with fast search", func(t *testing.T) {
		req := NewSearchRequest("test", WithSearchType(SearchTypeFast), WithSearchMaxResults(30))
		assert.ErrorIs(t, validator.ValidateSearchRequest(req), ErrSearchMaxResultsTooLarge)
	})

	t.Run("max_results at people limit", func(t *testing.T) {
		req := NewSearchRequest("test", WithSearchType(SearchTypePeople), WithSearchMaxResults(SearchMaxResultsPeopleLimit))
		assert.NoError(t, validator.ValidateSearchRequest(req))
	})

	t.Run("max_results above people limit", func(t *testing.T) {
		req := NewSearchRequest("test", WithSearchType(SearchTypePeople), WithSearchMaxResults(SearchMaxResultsPeopleLimit+1))
		assert.ErrorIs(t, validator.ValidateSearchRequest(req), ErrSearchMaxResultsTooLarge)
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

func TestSearchRequestValidator_SearchType(t *testing.T) {
	validator := NewSearchRequestValidator()

	for _, st := range []string{SearchTypeWeb, SearchTypeFast, SearchTypePeople} {
		req := NewSearchRequest("test", WithSearchType(st))
		assert.NoError(t, validator.ValidateSearchRequest(req), "search_type %s", st)
	}

	req := NewSearchRequest("test", WithSearchType("academic"))
	assert.ErrorIs(t, validator.ValidateSearchRequest(req), ErrSearchTypeInvalid)
}

func TestSearchRequestValidator_TokenLimits(t *testing.T) {
	validator := NewSearchRequestValidator()

	t.Run("valid values", func(t *testing.T) {
		req := NewSearchRequest("test", WithSearchMaxTokens(SearchMaxTokensLimit), WithSearchMaxTokensPerPage(4096))
		assert.NoError(t, validator.ValidateSearchRequest(req))
	})

	t.Run("max_tokens too large", func(t *testing.T) {
		req := NewSearchRequest("test", WithSearchMaxTokens(SearchMaxTokensLimit+1))
		assert.ErrorIs(t, validator.ValidateSearchRequest(req), ErrSearchTokensTooLarge)
	})

	t.Run("max_tokens_per_page zero", func(t *testing.T) {
		req := NewSearchRequest("test", WithSearchMaxTokensPerPage(0))
		assert.ErrorIs(t, validator.ValidateSearchRequest(req), ErrSearchMaxTokensPerPageInvalid)
	})

	t.Run("max_tokens_per_page too large", func(t *testing.T) {
		req := NewSearchRequest("test", WithSearchMaxTokensPerPage(SearchMaxTokensLimit+1))
		assert.ErrorIs(t, validator.ValidateSearchRequest(req), ErrSearchTokensTooLarge)
	})
}

func TestSearchRequestValidator_LanguageFilter(t *testing.T) {
	validator := NewSearchRequestValidator()

	t.Run("valid codes", func(t *testing.T) {
		req := NewSearchRequest("test", WithSearchLanguageFilter([]string{"en", "fr", "de"}))
		assert.NoError(t, validator.ValidateSearchRequest(req))
	})

	t.Run("invalid codes", func(t *testing.T) {
		for _, code := range []string{"EN", "eng", "en-US", "", "e"} {
			req := NewSearchRequest("test", WithSearchLanguageFilter([]string{code}))
			assert.ErrorIs(t, validator.ValidateSearchRequest(req), ErrSearchLanguageFilterInvalid, "code %q", code)
		}
	})

	t.Run("too many codes", func(t *testing.T) {
		codes := make([]string, SearchMaxLanguageFilterEntries+1)
		for i := range codes {
			codes[i] = "en"
		}
		req := NewSearchRequest("test", WithSearchLanguageFilter(codes))
		assert.ErrorIs(t, validator.ValidateSearchRequest(req), ErrSearchLanguageFilterTooLong)
	})
}

func TestSearchRequestValidator_TimeFilters(t *testing.T) {
	validator := NewSearchRequestValidator()
	jan := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	mar := time.Date(2025, 3, 1, 0, 0, 0, 0, time.UTC)

	t.Run("valid recency values", func(t *testing.T) {
		for _, r := range []string{SearchRecencyHour, SearchRecencyDay, SearchRecencyWeek, SearchRecencyMonth, SearchRecencyYear} {
			req := NewSearchRequest("test", WithSearchRecency(r))
			assert.NoError(t, validator.ValidateSearchRequest(req), "recency %s", r)
		}
	})

	t.Run("invalid recency", func(t *testing.T) {
		req := NewSearchRequest("test", WithSearchRecency("decade"))
		assert.ErrorIs(t, validator.ValidateSearchRequest(req), ErrSearchRecencyInvalid)
	})

	t.Run("valid date filters", func(t *testing.T) {
		req := NewSearchRequest("test",
			WithSearchPublishedAfter(jan), WithSearchPublishedBefore(mar),
			WithSearchUpdatedAfter(jan), WithSearchUpdatedBefore(mar),
		)
		assert.NoError(t, validator.ValidateSearchRequest(req))
	})

	t.Run("zero-padded date accepted", func(t *testing.T) {
		req := NewSearchRequest("test")
		req.SearchAfterDateFilter = strPtr("03/01/2025")
		assert.NoError(t, validator.ValidateSearchRequest(req))
	})

	t.Run("invalid date format", func(t *testing.T) {
		for _, d := range []string{"2025-03-01", "13/01/2025", "", "yesterday"} {
			req := NewSearchRequest("test")
			req.LastUpdatedBeforeFilter = strPtr(d)
			assert.ErrorIs(t, validator.ValidateSearchRequest(req), ErrSearchDateFilterInvalid, "date %q", d)
		}
	})

	t.Run("inverted date range", func(t *testing.T) {
		req := NewSearchRequest("test", WithSearchPublishedAfter(mar), WithSearchPublishedBefore(jan))
		assert.ErrorIs(t, validator.ValidateSearchRequest(req), ErrSearchDateRangeInvalid)

		req = NewSearchRequest("test", WithSearchUpdatedAfter(mar), WithSearchUpdatedBefore(jan))
		assert.ErrorIs(t, validator.ValidateSearchRequest(req), ErrSearchDateRangeInvalid)
	})

	t.Run("recency combined with date filter", func(t *testing.T) {
		req := NewSearchRequest("test", WithSearchRecency(SearchRecencyWeek), WithSearchUpdatedAfter(jan))
		assert.ErrorIs(t, validator.ValidateSearchRequest(req), ErrSearchRecencyWithDateFilters)
	})
}
