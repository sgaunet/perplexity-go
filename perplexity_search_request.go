package perplexity

import (
	"slices"
	"time"
)

// Search types supported by the Search API (search_type parameter).
const (
	// SearchTypeWeb is the default general web search ($5 per 1K requests).
	SearchTypeWeb = "web"
	// SearchTypeFast is a lower-latency, lower-cost web search ($1 per 1K requests).
	SearchTypeFast = "fast"
	// SearchTypePeople is a people search. It is the only type allowing max_results above 20.
	SearchTypePeople = "people"
)

// Recency values supported by the Search API (search_recency_filter parameter).
const (
	SearchRecencyHour  = "hour"
	SearchRecencyDay   = "day"
	SearchRecencyWeek  = "week"
	SearchRecencyMonth = "month"
	SearchRecencyYear  = "year"
)

// searchDateLayout is the MM/DD/YYYY input format expected by the Search API date filters.
const searchDateLayout = "1/2/2006"

// SearchRequest represents a request to the Perplexity Search API.
// The Search API provides direct access to Perplexity's real-time web index
// without the generative LLM layer, returning raw ranked search results.
//
// See https://docs.perplexity.ai/api-reference/search-post
type SearchRequest struct {
	// Query is either a string or a []string (multi-query, up to 5 queries).
	Query any `json:"query"`
	// SearchType selects the search engine: "web" (default), "fast" or "people".
	SearchType *string `json:"search_type,omitempty"`
	// MaxResults is the maximum number of results (1-20, up to 50 for people search).
	MaxResults *int `json:"max_results,omitempty"`
	// MaxTokens is the maximum total webpage content tokens returned across all results.
	MaxTokens *int `json:"max_tokens,omitempty"`
	// MaxTokensPerPage is the maximum webpage content tokens extracted from each result page.
	MaxTokensPerPage *int `json:"max_tokens_per_page,omitempty"`
	// Country is an ISO 3166-1 alpha-2 country code (e.g. "US").
	Country *string `json:"country,omitempty"`
	// SearchDomainFilter restricts (allowlist) or excludes (denylist, "-" prefix) domains.
	// Allowlist and denylist entries cannot be mixed. Up to 20 entries.
	SearchDomainFilter *[]string `json:"search_domain_filter,omitempty"`
	// SearchLanguageFilter restricts results to ISO 639-1 language codes (e.g. "en", "fr").
	SearchLanguageFilter *[]string `json:"search_language_filter,omitempty"`
	// SearchRecencyFilter restricts results to a relative time window: hour, day, week, month or year.
	// It cannot be combined with the date filters.
	SearchRecencyFilter *string `json:"search_recency_filter,omitempty"`
	// SearchAfterDateFilter keeps content published after this date (MM/DD/YYYY).
	SearchAfterDateFilter *string `json:"search_after_date_filter,omitempty"`
	// SearchBeforeDateFilter keeps content published before this date (MM/DD/YYYY).
	SearchBeforeDateFilter *string `json:"search_before_date_filter,omitempty"`
	// LastUpdatedAfterFilter keeps content last updated after this date (MM/DD/YYYY).
	LastUpdatedAfterFilter *string `json:"last_updated_after_filter,omitempty"`
	// LastUpdatedBeforeFilter keeps content last updated before this date (MM/DD/YYYY).
	LastUpdatedBeforeFilter *string `json:"last_updated_before_filter,omitempty"`
	// DisplayServerTime asks the API to include server_time (processing time) in the response.
	DisplayServerTime *bool `json:"display_server_time,omitempty"`

	// Deprecated: return_images is not part of the Search API specification and is ignored by the API.
	ReturnImages *bool `json:"return_images,omitempty"`
	// Deprecated: return_snippets is not part of the Search API specification; snippets are always returned.
	ReturnSnippets *bool `json:"return_snippets,omitempty"`
	// Deprecated: language_preference is a chat completion parameter. Use SearchLanguageFilter instead.
	LanguagePreference *string `json:"language_preference,omitempty"`
}

// SearchRequestOption is a function that modifies a SearchRequest.
type SearchRequestOption func(*SearchRequest)

// NewSearchRequest creates a new SearchRequest with the given query and options.
// The query can be either a single string or an array of strings for multi-query searches.
func NewSearchRequest(query any, opts ...SearchRequestOption) *SearchRequest {
	req := &SearchRequest{
		Query: query,
	}
	for _, opt := range opts {
		opt(req)
	}
	return req
}

// stringPtrOrNil returns nil for an empty string so the field is omitted from the request.
func stringPtrOrNil(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// WithSearchType sets the search type: SearchTypeWeb, SearchTypeFast or SearchTypePeople.
func WithSearchType(searchType string) SearchRequestOption {
	return func(r *SearchRequest) {
		r.SearchType = stringPtrOrNil(searchType)
	}
}

// WithSearchMaxResults sets the maximum number of results to return.
func WithSearchMaxResults(maxResults int) SearchRequestOption {
	return func(r *SearchRequest) {
		r.MaxResults = &maxResults
	}
}

// WithSearchMaxTokens sets the maximum total webpage content tokens returned across all results.
func WithSearchMaxTokens(maxTokens int) SearchRequestOption {
	return func(r *SearchRequest) {
		r.MaxTokens = &maxTokens
	}
}

// WithSearchMaxTokensPerPage sets the maximum webpage content tokens extracted from each result page.
func WithSearchMaxTokensPerPage(maxTokensPerPage int) SearchRequestOption {
	return func(r *SearchRequest) {
		r.MaxTokensPerPage = &maxTokensPerPage
	}
}

// WithSearchReturnImages sets whether to include image URLs in the search results.
//
// Deprecated: return_images is not part of the Search API specification and is ignored by the API.
func WithSearchReturnImages(include bool) SearchRequestOption {
	return func(r *SearchRequest) {
		r.ReturnImages = &include
	}
}

// WithSearchReturnSnippets sets whether to include text snippets in the search results.
//
// Deprecated: return_snippets is not part of the Search API specification; snippets are always returned.
func WithSearchReturnSnippets(include bool) SearchRequestOption {
	return func(r *SearchRequest) {
		r.ReturnSnippets = &include
	}
}

// WithSearchCountry sets the ISO 3166-1 alpha-2 country code to bias results by region.
// An empty string leaves the field unset.
func WithSearchCountry(country string) SearchRequestOption {
	return func(r *SearchRequest) {
		r.Country = stringPtrOrNil(country)
	}
}

// WithSearchLanguagePreference sets the preferred language for search results.
//
// Deprecated: language_preference is a chat completion parameter and is not part of the
// Search API specification. Use WithSearchLanguageFilter instead.
func WithSearchLanguagePreference(lang string) SearchRequestOption {
	return func(r *SearchRequest) {
		r.LanguagePreference = stringPtrOrNil(lang)
	}
}

// WithSearchLanguageFilter restricts results to the given ISO 639-1 language codes (e.g. "en", "fr").
func WithSearchLanguageFilter(languages []string) SearchRequestOption {
	return func(r *SearchRequest) {
		langs := slices.Clone(languages)
		r.SearchLanguageFilter = &langs
	}
}

// WithSearchDomains sets the domain filter (search_domain_filter).
// Use plain domains ("example.com"), paths ("example.com/blog") or TLDs (".gov") for an allowlist,
// or prefix every entry with "-" for a denylist. Both modes cannot be mixed.
func WithSearchDomains(domains []string) SearchRequestOption {
	return func(r *SearchRequest) {
		d := slices.Clone(domains)
		r.SearchDomainFilter = &d
	}
}

// WithSearchRecency restricts results to a relative time window
// (SearchRecencyHour, SearchRecencyDay, SearchRecencyWeek, SearchRecencyMonth or SearchRecencyYear).
func WithSearchRecency(recency string) SearchRequestOption {
	return func(r *SearchRequest) {
		r.SearchRecencyFilter = stringPtrOrNil(recency)
	}
}

// WithSearchPublishedAfter keeps content published after the given date.
func WithSearchPublishedAfter(date time.Time) SearchRequestOption {
	return func(r *SearchRequest) {
		s := date.Format(searchDateLayout)
		r.SearchAfterDateFilter = &s
	}
}

// WithSearchPublishedBefore keeps content published before the given date.
func WithSearchPublishedBefore(date time.Time) SearchRequestOption {
	return func(r *SearchRequest) {
		s := date.Format(searchDateLayout)
		r.SearchBeforeDateFilter = &s
	}
}

// WithSearchUpdatedAfter keeps content last updated after the given date.
func WithSearchUpdatedAfter(date time.Time) SearchRequestOption {
	return func(r *SearchRequest) {
		s := date.Format(searchDateLayout)
		r.LastUpdatedAfterFilter = &s
	}
}

// WithSearchUpdatedBefore keeps content last updated before the given date.
func WithSearchUpdatedBefore(date time.Time) SearchRequestOption {
	return func(r *SearchRequest) {
		s := date.Format(searchDateLayout)
		r.LastUpdatedBeforeFilter = &s
	}
}

// WithSearchDisplayServerTime asks the API to include server_time (processing time) in the response.
func WithSearchDisplayServerTime(display bool) SearchRequestOption {
	return func(r *SearchRequest) {
		r.DisplayServerTime = &display
	}
}
