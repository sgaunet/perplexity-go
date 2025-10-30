package perplexity

// SearchRequest represents a request to the Perplexity Search API.
// The Search API provides direct access to Perplexity's real-time web index
// without the generative LLM layer, returning raw ranked search results.
type SearchRequest struct {
	Query              interface{} `json:"query" validate:"required"` // string or []string
	MaxResults         *int        `json:"max_results,omitempty"`
	ReturnImages       *bool       `json:"return_images,omitempty"`
	ReturnSnippets     *bool       `json:"return_snippets,omitempty"`
	Country            *string     `json:"country,omitempty"`
	SearchDomainFilter *[]string   `json:"search_domain_filter,omitempty"`
}

// SearchRequestOption is a function that modifies a SearchRequest.
type SearchRequestOption func(*SearchRequest)

// NewSearchRequest creates a new SearchRequest with the given query and options.
// The query can be either a single string or an array of strings for multi-query searches.
func NewSearchRequest(query interface{}, opts ...SearchRequestOption) *SearchRequest {
	req := &SearchRequest{
		Query: query,
	}
	for _, opt := range opts {
		opt(req)
	}
	return req
}

// WithSearchMaxResults sets the maximum number of results to return per query.
func WithSearchMaxResults(maxResults int) SearchRequestOption {
	return func(r *SearchRequest) {
		r.MaxResults = &maxResults
	}
}

// WithSearchReturnImages sets whether to include image URLs in the search results.
func WithSearchReturnImages(include bool) SearchRequestOption {
	return func(r *SearchRequest) {
		r.ReturnImages = &include
	}
}

// WithSearchReturnSnippets sets whether to include text snippets in the search results.
func WithSearchReturnSnippets(include bool) SearchRequestOption {
	return func(r *SearchRequest) {
		r.ReturnSnippets = &include
	}
}

// WithSearchCountry sets the ISO country code to bias or filter results by region.
func WithSearchCountry(country string) SearchRequestOption {
	return func(r *SearchRequest) {
		r.Country = &country
	}
}

// WithSearchDomains sets the domain filters for the search.
func WithSearchDomains(domains []string) SearchRequestOption {
	return func(r *SearchRequest) {
		r.SearchDomainFilter = &domains
	}
}
