# Search API Example

This example demonstrates how to use the Perplexity Search API client to perform web searches.

## Overview

The Perplexity Search API provides direct access to Perplexity's real-time web index without the generative LLM layer, returning raw ranked search results with structured snippets.

This example showcases four search scenarios:
1. **Simple Search** - Basic single query search
2. **Advanced Search** - Content extraction, country, language, domain denylist, recency and server time
3. **Multi-Query Fast Search** - Several queries in one request with `search_type: "fast"`
4. **Date Filters** - Publication and last-updated date filters

## Prerequisites

You need a Perplexity API key with Search API access. The Search API is priced at **$5 per 1,000 requests** (**$1 per 1,000** for Fast Search), separate from chat completion pricing.

## Building

From the project root:

```bash
go build -o bin/search-example ./cmd/search-example
```

Or directly:

```bash
cd cmd/search-example
go build -o search-example .
```

## Running

Set your API key as an environment variable:

```bash
export PPLX_API_KEY=your_api_key_here
./bin/search-example
```

Or in one line:

```bash
PPLX_API_KEY=your_api_key ./bin/search-example
```

## Example Functions

### Example 1: Simple Search

```go
req := perplexity.NewSearchRequest("latest developments in quantum computing")
resp, err := client.SendSearchRequest(req) // validated before sending
```

Prints the search `ID` and, for each result, title, URL, snippet, publication date and last-updated date.

### Example 2: Advanced Search with Filters

```go
req := perplexity.NewSearchRequest(
    "best Go web frameworks",
    perplexity.WithSearchMaxResults(10),
    perplexity.WithSearchMaxTokensPerPage(1024),
    perplexity.WithSearchCountry("US"),
    perplexity.WithSearchLanguageFilter([]string{"en"}),
    perplexity.WithSearchDomains([]string{"-reddit.com", "-pinterest.com"}),
    perplexity.WithSearchRecency(perplexity.SearchRecencyYear),
    perplexity.WithSearchDisplayServerTime(true),
)
```

### Example 3: Multi-Query Fast Search

```go
queries := []string{"Go concurrency patterns", "Go performance optimization", "Go best practices"}
req := perplexity.NewSearchRequest(
    queries,
    perplexity.WithSearchType(perplexity.SearchTypeFast),
    perplexity.WithSearchMaxResults(5),
)
```

Up to 5 queries per request. A multi-query request is billed as one request but consumes one
rate-limit unit per query.

### Example 4: Date Filters

```go
now := time.Now()
req := perplexity.NewSearchRequest(
    "Go release notes",
    perplexity.WithSearchPublishedAfter(now.AddDate(-1, 0, 0)),
    perplexity.WithSearchUpdatedBefore(now),
)
```

Date filters cannot be combined with `WithSearchRecency`.

## Available Search Options

| Option | JSON parameter | Description |
|--------|----------------|-------------|
| `WithSearchType(t)` | `search_type` | `SearchTypeWeb` (default), `SearchTypeFast`, `SearchTypePeople` |
| `WithSearchMaxResults(n)` | `max_results` | 1-20 (up to 50 for people search) |
| `WithSearchMaxTokens(n)` | `max_tokens` | Total content tokens across all results |
| `WithSearchMaxTokensPerPage(n)` | `max_tokens_per_page` | Content tokens extracted per page |
| `WithSearchCountry(code)` | `country` | ISO 3166-1 alpha-2 code (e.g. "US", "GB") |
| `WithSearchLanguageFilter(codes)` | `search_language_filter` | ISO 639-1 codes (e.g. "en", "fr") |
| `WithSearchDomains(domains)` | `search_domain_filter` | Allowlist or denylist (`-` prefix), up to 20 entries |
| `WithSearchRecency(r)` | `search_recency_filter` | hour, day, week, month, year |
| `WithSearchPublishedAfter/Before(t)` | `search_after_date_filter` / `search_before_date_filter` | Publication date |
| `WithSearchUpdatedAfter/Before(t)` | `last_updated_after_filter` / `last_updated_before_filter` | Last-modified date |
| `WithSearchDisplayServerTime(b)` | `display_server_time` | Include `server_time` in the response |

Deprecated (ignored by the API): `WithSearchReturnImages`, `WithSearchReturnSnippets`, `WithSearchLanguagePreference`.

## Response Fields

`SearchResponse`:
- `ID` - Unique identifier of the search
- `Results` - Ranked results
- `ServerTime` - Processing time, when `display_server_time` is set

Each `SearchResultItem` contains:
- `Title` - Page title
- `URL` - Page URL
- `Snippet` - Extracted page content (size driven by `max_tokens` / `max_tokens_per_page`)
- `Date` - Publication date, `YYYY-MM-DD` (optional)
- `LastUpdated` - Last-updated date, `YYYY-MM-DD` (optional)

`Score` and `Images` are deprecated and never populated by the API.

## Error Handling

`SendSearchRequest` validates the request before sending it and returns the validation error
(e.g. `ErrSearchQueryArrayTooLong`, `ErrSearchDomainFilterMixedModes`, `ErrSearchRecencyWithDateFilters`).

API errors:
- `ErrUnauthorized` - Invalid or missing API key (HTTP 401)
- `*ResponseError` - Any other API error, with `StatusCode` and, for HTTP 422, the validation `Detail`
- Network errors - Connection issues, timeouts

```go
var respErr *perplexity.ResponseError
if errors.As(err, &respErr) && respErr.StatusCode == http.StatusTooManyRequests {
    // rate limited: retry with exponential backoff and jitter
}
```

## Tips

1. **Domain Filtering**: use domains without protocol or `www.`
   - Allowlist: `"github.com"`, `"example.com/blog"`, `".gov"`
   - Denylist: `"-reddit.com"`, `"-reddit.com/r/all"`
   - Do not mix allowlist and denylist entries in the same request

2. **Rate Limiting**: the Search API allows 50 query units per second (burst 50) for all tiers.
   Each query of a multi-query request counts as one unit. Retry 429 responses with exponential backoff.

3. **Fast Search**: use `SearchTypeFast` for agent loops and high-volume workloads, `SearchTypeWeb` for hard queries.

4. **Context Support**: use `SendSearchRequestWithContext()` for timeout control:
   ```go
   ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
   defer cancel()
   resp, err := client.SendSearchRequestWithContext(ctx, req)
   ```

## Comparison with Chat Completions

| Feature | Search API | Chat Completions |
|---------|------------|------------------|
| **Purpose** | Raw web search results | AI-generated responses |
| **Response** | Ranked list of URLs with snippets | Natural language text |
| **Pricing** | $5 per 1K requests ($1 for Fast Search) | Variable by model |
| **Latency** | Lower (no generation) | Higher (includes generation) |
| **Use Case** | Research, data gathering | Q&A, summarization, analysis |

## See Also

- [Perplexity Search API Documentation](https://docs.perplexity.ai/api-reference/search-post)
- [Main Library Documentation](../../README.md)
- [GoDoc](https://godoc.org/github.com/sgaunet/perplexity-go/v2)
