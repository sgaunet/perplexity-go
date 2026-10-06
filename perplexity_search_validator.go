package perplexity

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
)

// Search API limits, from https://docs.perplexity.ai/api-reference/search-post.
const (
	// SearchMaxQueries is the maximum number of queries in a multi-query request.
	SearchMaxQueries = 5
	// SearchMaxResultsLimit is the maximum max_results for web and fast search.
	SearchMaxResultsLimit = 20
	// SearchMaxResultsPeopleLimit is the maximum max_results for people search.
	SearchMaxResultsPeopleLimit = 50
	// SearchMaxTokensLimit is the maximum value for max_tokens and max_tokens_per_page.
	SearchMaxTokensLimit = 1_000_000
	// SearchMaxDomainFilterEntries is the maximum number of search_domain_filter entries.
	SearchMaxDomainFilterEntries = 20
	// SearchMaxDomainFilterEntryLength is the maximum length of a search_domain_filter entry.
	SearchMaxDomainFilterEntryLength = 253
	// SearchMaxLanguageFilterEntries is the maximum number of search_language_filter entries.
	SearchMaxLanguageFilterEntries = 20
)

// Search API validation errors.
var (
	ErrSearchQueryRequired             = errors.New("query is required")
	ErrSearchQueryStringEmpty          = errors.New("query string cannot be empty")
	ErrSearchQueryArrayEmpty           = errors.New("query array cannot be empty")
	ErrSearchQueryInvalidType          = errors.New("query must be a string or array of strings")
	ErrSearchQueryArrayTooLong         = fmt.Errorf("query array cannot contain more than %d queries", SearchMaxQueries)
	ErrSearchMaxResultsInvalid         = errors.New("max_results must be a positive integer")
	ErrSearchMaxResultsTooLarge        = fmt.Errorf("max_results cannot exceed %d (%d for people search)", SearchMaxResultsLimit, SearchMaxResultsPeopleLimit)
	ErrSearchMaxTokensInvalid          = errors.New("max_tokens must be a positive integer")
	ErrSearchMaxTokensPerPageInvalid   = errors.New("max_tokens_per_page must be a positive integer")
	ErrSearchTokensTooLarge            = fmt.Errorf("max_tokens and max_tokens_per_page cannot exceed %d", SearchMaxTokensLimit)
	ErrSearchTypeInvalid               = errors.New("search_type must be one of web, fast, people")
	ErrSearchCountryInvalid            = errors.New("country must be a valid ISO 3166-1 alpha-2 code (2 uppercase letters)")
	ErrSearchLanguagePreferenceInvalid = errors.New("language_preference must be valid ISO 639 format (e.g., 'en', 'en-US')")
	ErrSearchLanguageFilterInvalid     = errors.New("search_language_filter entries must be lowercase ISO 639-1 codes (e.g., 'en')")
	ErrSearchLanguageFilterTooLong     = fmt.Errorf("search_language_filter cannot contain more than %d entries", SearchMaxLanguageFilterEntries)
	ErrSearchDomainFilterEntryEmpty    = errors.New("domain filter entry cannot be empty")
	ErrSearchDomainFilterEntryInvalid  = errors.New("domain filter entry is not a valid domain or pattern")
	ErrSearchDomainFilterTooLong       = fmt.Errorf("search_domain_filter cannot contain more than %d entries", SearchMaxDomainFilterEntries)
	ErrSearchDomainFilterMixedModes    = errors.New("search_domain_filter cannot mix allowlist and denylist (\"-\" prefixed) entries")
	ErrSearchQueryArrayElementEmpty    = errors.New("query array element cannot be empty")
	ErrSearchRecencyInvalid            = errors.New("search_recency_filter must be one of hour, day, week, month, year")
	ErrSearchRecencyWithDateFilters    = errors.New("search_recency_filter cannot be combined with date filters")
	ErrSearchDateFilterInvalid         = errors.New("search date filter must be in MM/DD/YYYY format (e.g., 3/1/2025, 12/31/2024)")
	ErrSearchDateRangeInvalid          = errors.New("search date filter range is invalid: 'after' date must not be later than 'before' date")
)

var (
	countryCodeRegex        = regexp.MustCompile(`^[A-Z]{2}$`)
	languageCodeRegex       = regexp.MustCompile(`^[a-z]{2}$`)
	languagePreferenceRegex = regexp.MustCompile(`^[a-z]{2}(-[A-Z]{2})?$`)
	dotLetterRegex          = regexp.MustCompile(`\.[a-zA-Z]`)
	domainPatternRegex      = regexp.MustCompile(`^[\w\*\-\.]+$`)
)

// SearchRequestValidator provides validation for SearchRequest objects.
type SearchRequestValidator struct{}

// NewSearchRequestValidator creates a new SearchRequestValidator.
func NewSearchRequestValidator() *SearchRequestValidator {
	return &SearchRequestValidator{}
}

// ValidateSearchRequest validates a SearchRequest against the Search API constraints.
// It returns the first error found.
func (v *SearchRequestValidator) ValidateSearchRequest(req *SearchRequest) error {
	if req == nil {
		return ErrNilRequest
	}

	checks := []func(*SearchRequest) error{
		func(r *SearchRequest) error { return v.validateQuery(r.Query) },
		v.validateSearchType,
		v.validateMaxResultsField,
		v.validateTokenFields,
		func(r *SearchRequest) error { return v.validateCountryField(r.Country) },
		func(r *SearchRequest) error { return v.validateLanguagePreferenceField(r.LanguagePreference) },
		func(r *SearchRequest) error { return v.validateLanguageFilterField(r.SearchLanguageFilter) },
		func(r *SearchRequest) error { return v.validateDomainFilterField(r.SearchDomainFilter) },
		v.validateTimeFilters,
	}
	for _, check := range checks {
		if err := check(req); err != nil {
			return err
		}
	}
	return nil
}

// validateSearchType validates the search_type field.
func (v *SearchRequestValidator) validateSearchType(req *SearchRequest) error {
	if req.SearchType == nil {
		return nil
	}
	switch *req.SearchType {
	case SearchTypeWeb, SearchTypeFast, SearchTypePeople:
		return nil
	default:
		return fmt.Errorf("%w, got: %s", ErrSearchTypeInvalid, *req.SearchType)
	}
}

// validateMaxResultsField validates max_results, whose upper bound depends on search_type.
func (v *SearchRequestValidator) validateMaxResultsField(req *SearchRequest) error {
	if err := v.validateMaxResults(req.MaxResults); err != nil {
		return err
	}
	if req.MaxResults == nil {
		return nil
	}
	limit := SearchMaxResultsLimit
	if req.SearchType != nil && *req.SearchType == SearchTypePeople {
		limit = SearchMaxResultsPeopleLimit
	}
	if *req.MaxResults > limit {
		return fmt.Errorf("%w, got: %d", ErrSearchMaxResultsTooLarge, *req.MaxResults)
	}
	return nil
}

// validateMaxResults validates that max_results is positive.
func (v *SearchRequestValidator) validateMaxResults(maxResults *int) error {
	if maxResults != nil && *maxResults <= 0 {
		return ErrSearchMaxResultsInvalid
	}
	return nil
}

// validateTokenFields validates max_tokens and max_tokens_per_page.
func (v *SearchRequestValidator) validateTokenFields(req *SearchRequest) error {
	if err := v.validateMaxTokens(req.MaxTokens); err != nil {
		return err
	}
	if req.MaxTokensPerPage != nil && *req.MaxTokensPerPage <= 0 {
		return ErrSearchMaxTokensPerPageInvalid
	}
	for _, tokens := range []*int{req.MaxTokens, req.MaxTokensPerPage} {
		if tokens != nil && *tokens > SearchMaxTokensLimit {
			return fmt.Errorf("%w, got: %d", ErrSearchTokensTooLarge, *tokens)
		}
	}
	return nil
}

// validateMaxTokens validates that max_tokens is positive.
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

// validateLanguagePreferenceField validates the deprecated language_preference field.
func (v *SearchRequestValidator) validateLanguagePreferenceField(languagePreference *string) error {
	if languagePreference != nil && *languagePreference != "" {
		return v.validateLanguagePreference(*languagePreference)
	}
	return nil
}

// validateLanguageFilterField validates the search_language_filter field.
func (v *SearchRequestValidator) validateLanguageFilterField(languages *[]string) error {
	if languages == nil {
		return nil
	}
	if len(*languages) > SearchMaxLanguageFilterEntries {
		return ErrSearchLanguageFilterTooLong
	}
	for i, lang := range *languages {
		if !languageCodeRegex.MatchString(lang) {
			return fmt.Errorf("%w at index %d, got: %q", ErrSearchLanguageFilterInvalid, i, lang)
		}
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

// validateQuery validates that the query is a non-empty string or array of up to
// SearchMaxQueries non-empty strings. A []any holding only strings (as produced by
// json.Unmarshal) is accepted.
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
		return v.validateQueryArray(q)
	case []any:
		queries := make([]string, len(q))
		for i, item := range q {
			s, ok := item.(string)
			if !ok {
				return ErrSearchQueryInvalidType
			}
			queries[i] = s
		}
		return v.validateQueryArray(queries)
	default:
		return ErrSearchQueryInvalidType
	}
}

// validateQueryArray validates a multi-query array.
func (v *SearchRequestValidator) validateQueryArray(queries []string) error {
	if len(queries) == 0 {
		return ErrSearchQueryArrayEmpty
	}
	if len(queries) > SearchMaxQueries {
		return fmt.Errorf("%w, got: %d", ErrSearchQueryArrayTooLong, len(queries))
	}
	for i, s := range queries {
		if s == "" {
			return fmt.Errorf("%w at index %d", ErrSearchQueryArrayElementEmpty, i)
		}
	}
	return nil
}

// validateCountry validates that the country code is a valid ISO 3166-1 alpha-2 code.
func (v *SearchRequestValidator) validateCountry(country string) error {
	if !countryCodeRegex.MatchString(country) {
		return fmt.Errorf("%w, got: %s", ErrSearchCountryInvalid, country)
	}
	return nil
}

// validateDomainFilter validates domain filter entries: count, length, format, and
// that allowlist and denylist ("-" prefixed) entries are not mixed.
func (v *SearchRequestValidator) validateDomainFilter(domains []string) error {
	if len(domains) == 0 {
		return nil
	}
	if len(domains) > SearchMaxDomainFilterEntries {
		return fmt.Errorf("%w, got: %d", ErrSearchDomainFilterTooLong, len(domains))
	}

	denyCount := 0
	for i, domain := range domains {
		if domain == "" {
			return fmt.Errorf("%w at index %d", ErrSearchDomainFilterEntryEmpty, i)
		}
		if strings.HasPrefix(domain, "-") {
			denyCount++
		}
		if !isValidDomainFilterEntry(domain) {
			return fmt.Errorf("%w at index %d: %s", ErrSearchDomainFilterEntryInvalid, i, domain)
		}
	}
	if denyCount > 0 && denyCount < len(domains) {
		return ErrSearchDomainFilterMixedModes
	}

	return nil
}

// isValidDomainFilterEntry checks a single search_domain_filter entry. Entries are a domain,
// a TLD (".gov") or a domain with a path ("example.com/blog"), optionally prefixed with "-".
// Protocols and whitespace are rejected.
func isValidDomainFilterEntry(entry string) bool {
	if len(entry) > SearchMaxDomainFilterEntryLength ||
		strings.Contains(entry, "://") ||
		strings.ContainsFunc(entry, func(r rune) bool { return r == ' ' || r == '\t' || r == '\n' || r == '\r' }) {
		return false
	}
	host, _, _ := strings.Cut(strings.TrimPrefix(entry, "-"), "/")
	if host == "" {
		return false
	}
	return host == "*" || containsDot(host) || isValidDomainPattern(host)
}

// containsDot checks if a string starts with a dot or contains a dot followed by a letter.
func containsDot(s string) bool {
	return len(s) > 0 && (s[0] == '.' || dotLetterRegex.MatchString(s))
}

// isValidDomainPattern checks if a string is a valid domain pattern (contains wildcards or valid characters).
func isValidDomainPattern(s string) bool {
	return domainPatternRegex.MatchString(s)
}

// validateLanguagePreference validates the deprecated language preference format.
func (v *SearchRequestValidator) validateLanguagePreference(lang string) error {
	// ISO 639-1 codes are 2 characters, with optional country code up to 5 total
	if len(lang) < 2 || len(lang) > 5 {
		return fmt.Errorf("%w: length must be 2-5 characters, got: %s", ErrSearchLanguagePreferenceInvalid, lang)
	}
	if !languagePreferenceRegex.MatchString(lang) {
		return fmt.Errorf("%w, got: %s", ErrSearchLanguagePreferenceInvalid, lang)
	}
	return nil
}

// validateTimeFilters validates the recency filter, the date filters and their compatibility.
func (v *SearchRequestValidator) validateTimeFilters(req *SearchRequest) error {
	dateFilters := []struct {
		name  string
		value *string
	}{
		{"search_after_date_filter", req.SearchAfterDateFilter},
		{"search_before_date_filter", req.SearchBeforeDateFilter},
		{"last_updated_after_filter", req.LastUpdatedAfterFilter},
		{"last_updated_before_filter", req.LastUpdatedBeforeFilter},
	}

	parsed := make([]time.Time, len(dateFilters))
	hasDateFilter := false
	for i, f := range dateFilters {
		if f.value == nil {
			continue
		}
		hasDateFilter = true
		t, err := time.Parse(searchDateLayout, *f.value)
		if err != nil {
			return fmt.Errorf("%w: %s=%q", ErrSearchDateFilterInvalid, f.name, *f.value)
		}
		parsed[i] = t
	}

	// Pairs: (published after, published before), (updated after, updated before)
	for _, pair := range [][2]int{{0, 1}, {2, 3}} {
		after, before := parsed[pair[0]], parsed[pair[1]]
		if !after.IsZero() && !before.IsZero() && after.After(before) {
			return fmt.Errorf("%w: %s > %s", ErrSearchDateRangeInvalid, dateFilters[pair[0]].name, dateFilters[pair[1]].name)
		}
	}

	return v.validateRecency(req.SearchRecencyFilter, hasDateFilter)
}

// validateRecency validates search_recency_filter, which cannot be combined with date filters.
func (v *SearchRequestValidator) validateRecency(recency *string, hasDateFilter bool) error {
	if recency == nil {
		return nil
	}
	switch *recency {
	case SearchRecencyHour, SearchRecencyDay, SearchRecencyWeek, SearchRecencyMonth, SearchRecencyYear:
	default:
		return fmt.Errorf("%w, got: %s", ErrSearchRecencyInvalid, *recency)
	}
	if hasDateFilter {
		return ErrSearchRecencyWithDateFilters
	}
	return nil
}
