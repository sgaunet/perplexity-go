package perplexity

import (
	"errors"
	"fmt"
	"regexp"

	"github.com/go-playground/validator/v10"
)

// Search API validation errors.
var (
	ErrSearchQueryRequired            = errors.New("query is required")
	ErrSearchQueryStringEmpty         = errors.New("query string cannot be empty")
	ErrSearchQueryArrayEmpty          = errors.New("query array cannot be empty")
	ErrSearchQueryInvalidType         = errors.New("query must be a string or array of strings")
	ErrSearchMaxResultsInvalid        = errors.New("max_results must be a positive integer")
	ErrSearchMaxTokensInvalid         = errors.New("max_tokens must be a positive integer")
	ErrSearchCountryInvalid           = errors.New("country must be a valid ISO 3166-1 alpha-2 code (2 uppercase letters)")
	ErrSearchLanguagePreferenceInvalid = errors.New("language_preference must be valid ISO 639 format (e.g., 'en', 'en-US')")
	ErrSearchDomainFilterEntryEmpty   = errors.New("domain filter entry cannot be empty")
	ErrSearchDomainFilterEntryInvalid = errors.New("domain filter entry is not a valid domain or pattern")
	ErrSearchQueryArrayElementEmpty   = errors.New("query array element cannot be empty")
)

// SearchRequestValidator provides validation for SearchRequest objects.
type SearchRequestValidator struct {
	validator *validator.Validate
}

// NewSearchRequestValidator creates a new SearchRequestValidator.
func NewSearchRequestValidator() *SearchRequestValidator {
	return &SearchRequestValidator{
		validator: validator.New(),
	}
}

// ValidateSearchRequest validates a SearchRequest.
func (v *SearchRequestValidator) ValidateSearchRequest(req *SearchRequest) error {
	if req == nil {
		return ErrNilRequest
	}

	// Validate query
	if err := v.validateQuery(req.Query); err != nil {
		return err
	}

	// Validate optional fields
	if err := v.validateOptionalFields(req); err != nil {
		return err
	}

	return nil
}

// validateOptionalFields validates all optional fields in the search request.
func (v *SearchRequestValidator) validateOptionalFields(req *SearchRequest) error {
	if err := v.validateMaxResults(req.MaxResults); err != nil {
		return err
	}
	if err := v.validateMaxTokens(req.MaxTokens); err != nil {
		return err
	}
	if err := v.validateCountryField(req.Country); err != nil {
		return err
	}
	if err := v.validateLanguagePreferenceField(req.LanguagePreference); err != nil {
		return err
	}
	if err := v.validateDomainFilterField(req.SearchDomainFilter); err != nil {
		return err
	}
	return nil
}

// validateMaxResults validates the max_results field.
func (v *SearchRequestValidator) validateMaxResults(maxResults *int) error {
	if maxResults != nil && *maxResults <= 0 {
		return ErrSearchMaxResultsInvalid
	}
	return nil
}

// validateMaxTokens validates the max_tokens field.
func (v *SearchRequestValidator) validateMaxTokens(maxTokens *int) error {
	if maxTokens != nil && *maxTokens <= 0 {
		return ErrSearchMaxTokensInvalid
	}
	return nil
}

// validateCountryField validates the country field.
func (v *SearchRequestValidator) validateCountryField(country *string) error {
	if country != nil && *country != "" {
		return v.validateCountry(*country)
	}
	return nil
}

// validateLanguagePreferenceField validates the language_preference field.
func (v *SearchRequestValidator) validateLanguagePreferenceField(languagePreference *string) error {
	if languagePreference != nil && *languagePreference != "" {
		return v.validateLanguagePreference(*languagePreference)
	}
	return nil
}

// validateDomainFilterField validates the domain filter field.
func (v *SearchRequestValidator) validateDomainFilterField(domainFilter *[]string) error {
	if domainFilter != nil {
		return v.validateDomainFilter(*domainFilter)
	}
	return nil
}

// validateQuery validates that the query is a non-empty string or array of strings.
func (v *SearchRequestValidator) validateQuery(query any) error {
	if query == nil {
		return ErrSearchQueryRequired
	}

	switch q := query.(type) {
	case string:
		if q == "" {
			return ErrSearchQueryStringEmpty
		}
		return nil
	case []string:
		if len(q) == 0 {
			return ErrSearchQueryArrayEmpty
		}
		for i, s := range q {
			if s == "" {
				return fmt.Errorf("%w at index %d", ErrSearchQueryArrayElementEmpty, i)
			}
		}
		return nil
	default:
		return ErrSearchQueryInvalidType
	}
}

// validateCountry validates that the country code is a valid ISO 3166-1 alpha-2 code.
func (v *SearchRequestValidator) validateCountry(country string) error {
	// ISO 3166-1 alpha-2 codes are exactly 2 uppercase letters
	matched, err := regexp.MatchString(`^[A-Z]{2}$`, country)
	if err != nil {
		return fmt.Errorf("failed to validate country code: %w", err)
	}
	if !matched {
		return fmt.Errorf("%w, got: %s", ErrSearchCountryInvalid, country)
	}
	return nil
}

// validateDomainFilter validates domain filter entries.
func (v *SearchRequestValidator) validateDomainFilter(domains []string) error {
	if len(domains) == 0 {
		return nil
	}

	for i, domain := range domains {
		if domain == "" {
			return fmt.Errorf("%w at index %d", ErrSearchDomainFilterEntryEmpty, i)
		}
		// Basic domain validation - should contain at least one dot or be a wildcard pattern
		if domain != "*" && !containsDot(domain) && !isValidDomainPattern(domain) {
			return fmt.Errorf("%w at index %d: %s", ErrSearchDomainFilterEntryInvalid, i, domain)
		}
	}

	return nil
}

// containsDot checks if a string contains a dot.
func containsDot(s string) bool {
	return len(s) > 0 && (s[0] == '.' || regexp.MustCompile(`\.[a-zA-Z]`).MatchString(s))
}

// isValidDomainPattern checks if a string is a valid domain pattern (contains wildcards or valid characters).
func isValidDomainPattern(s string) bool {
	// Allow patterns like *.example.com, example.*, etc.
	matched, _ := regexp.MatchString(`^[\w\*\-\.]+$`, s)
	return matched
}

// validateLanguagePreference validates language preference format.
func (v *SearchRequestValidator) validateLanguagePreference(lang string) error {
	// Check length (ISO 639-1 codes are 2 characters, with optional country code up to 5 total)
	if len(lang) < 2 || len(lang) > 5 {
		return fmt.Errorf("%w: length must be 2-5 characters, got: %s", ErrSearchLanguagePreferenceInvalid, lang)
	}

	// Validate format: lowercase language code, optional uppercase country code with hyphen
	// Examples: "en", "fr", "en-US", "fr-CA"
	matched, err := regexp.MatchString(`^[a-z]{2}(-[A-Z]{2})?$`, lang)
	if err != nil {
		return fmt.Errorf("failed to validate language preference: %w", err)
	}
	if !matched {
		return fmt.Errorf("%w, got: %s", ErrSearchLanguagePreferenceInvalid, lang)
	}

	return nil
}
