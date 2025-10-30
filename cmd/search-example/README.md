# Search API Example

This example demonstrates how to use the Perplexity Search API client to perform web searches.

## Overview

The Perplexity Search API provides direct access to Perplexity's real-time web index without the generative LLM layer, returning raw ranked search results with structured snippets.

This example showcases three different search scenarios:
1. **Simple Search** - Basic single query search
2. **Advanced Search** - Search with options (max results, images, snippets, domain filters)
3. **Multi-Query Search** - Execute multiple queries in a single request

## Prerequisites

You need a Perplexity API key with Search API access. The Search API is priced at **$5 per 1,000 requests** (separate from chat completion pricing).

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

Performs a basic search query with default options:

```go
req := perplexity.NewSearchRequest("latest developments in quantum computing")
```

**Features demonstrated:**
- Basic request creation
- Request validation with `SearchRequestValidator`
- Sending requests with `SendSearchRequest()`
- Accessing result fields (Title, URL, Snippet, Date, Score)

**Output includes:**
- Result count
- For each result: Title, URL, snippet (if available), date, relevance score

### Example 2: Advanced Search with Options

Shows how to use search options for more control:

```go
req := perplexity.NewSearchRequest(
    "best Go web frameworks 2025",
    perplexity.WithSearchMaxResults(10),
    perplexity.WithSearchReturnImages(true),
    perplexity.WithSearchReturnSnippets(true),
    perplexity.WithSearchCountry("US"),
    perplexity.WithSearchDomains([]string{"golang.org", "github.com"}),
)
```

**Options demonstrated:**
- `WithSearchMaxResults(10)` - Limit number of results
- `WithSearchReturnImages(true)` - Include images in results
- `WithSearchReturnSnippets(true)` - Include text snippets
- `WithSearchCountry("US")` - Country-specific results
- `WithSearchDomains([]string{"golang.org", "github.com"})` - Filter by domains

**Note:** Domain filters must be specific domains (e.g., "github.com"), not wildcard patterns (e.g., "*.com").

**Output includes:**
- All fields from Example 1
- Image URLs with dimensions (when available)
- Formatted with `result.String()` and `img.String()` methods

### Example 3: Multi-Query Search

Demonstrates searching multiple queries in a single request:

```go
queries := []string{
    "Go concurrency patterns",
    "Go performance optimization",
    "Go best practices 2025",
}
req := perplexity.NewSearchRequest(queries, perplexity.WithSearchMaxResults(5))
```

**Features demonstrated:**
- Passing multiple queries as `[]string`
- Results aggregated across all queries
- More efficient than individual requests

**Output includes:**
- Total result count across all queries
- Simplified view (title + URL only)

## Available Search Options

| Option | Description | Type |
|--------|-------------|------|
| `WithSearchMaxResults(n)` | Maximum number of results to return | `int` |
| `WithSearchReturnImages(bool)` | Include images in results | `bool` |
| `WithSearchReturnSnippets(bool)` | Include text snippets | `bool` |
| `WithSearchCountry(code)` | Country code for localized results (e.g., "US", "GB") | `string` |
| `WithSearchDomains(domains)` | Filter results to specific domains | `[]string` |

## Response Fields

Each `SearchResultItem` contains:

- `Title` - Result title (required)
- `URL` - Result URL (required)
- `Snippet` - Text snippet from the page (optional)
- `Date` - Publication/modification date (optional)
- `Score` - Relevance score (optional)
- `Images` - Array of related images (optional)
  - `URL` - Image URL
  - `Width` - Image width in pixels (optional)
  - `Height` - Image height in pixels (optional)

## Error Handling

The example demonstrates proper error handling:

1. **Environment variable check** - Exits if `PPLX_API_KEY` not set
2. **Request validation** - Validates before sending with `validator.ValidateSearchRequest()`
3. **API errors** - Checks `client.SendSearchRequest()` return value

Common errors:
- `ErrUnauthorized` - Invalid or missing API key
- `ErrBadRequest` - Invalid request parameters (e.g., empty query, invalid domains)
- Network errors - Connection issues, timeouts

## Validation

The example uses `SearchRequestValidator` to validate requests before sending:

```go
validator := perplexity.NewSearchRequestValidator()
if err := validator.ValidateSearchRequest(req); err != nil {
    fmt.Printf("Validation error: %v\n", err)
    return
}
```

**Validation rules:**
- Query cannot be empty
- Domain filters must be valid domain names
- Max results must be positive (if specified)

## Tips

1. **Domain Filtering**: Use specific domains without wildcards
   - ✅ Good: `"github.com"`, `"golang.org"`
   - ❌ Bad: `"*.github.com"`, `"*.org"`

2. **Rate Limiting**: The Search API has rate limits. For production use, implement:
   - Exponential backoff for retries
   - Request queuing
   - Concurrent request limiting (~3-5 recommended)

3. **Context Support**: Use `SendSearchRequestWithContext()` for timeout control:
   ```go
   ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
   defer cancel()
   resp, err := client.SendSearchRequestWithContext(ctx, req)
   ```

4. **Thread Safety**: The client and response objects are safe for concurrent use

## Comparison with Chat Completions

| Feature | Search API | Chat Completions |
|---------|------------|------------------|
| **Purpose** | Raw web search results | AI-generated responses |
| **Response** | Ranked list of URLs with snippets | Natural language text |
| **Pricing** | $5 per 1K requests | Variable by model |
| **Latency** | Lower (no generation) | Higher (includes generation) |
| **Use Case** | Research, data gathering | Q&A, summarization, analysis |

## See Also

- [Perplexity Search API Documentation](https://docs.perplexity.ai/reference/post_search)
- [Main Library Documentation](../../README.md)
- [GoDoc](https://godoc.org/github.com/sgaunet/perplexity-go/v2)
