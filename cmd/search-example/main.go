package main

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	perplexity "github.com/sgaunet/perplexity-go/v2"
)

func main() {
	// Get API key from environment
	apiKey := os.Getenv("PPLX_API_KEY")
	if apiKey == "" {
		fmt.Println("Error: PPLX_API_KEY environment variable is not set")
		fmt.Println("Usage: export PPLX_API_KEY=your_api_key && ./search-example")
		os.Exit(1)
	}

	// Create client
	client := perplexity.NewClient(apiKey)

	fmt.Println("=== Example 1: Simple Search ===")
	simpleSearch(client)

	fmt.Println("\n=== Example 2: Advanced Search with Filters ===")
	advancedSearch(client)

	fmt.Println("\n=== Example 3: Multi-query Fast Search ===")
	multiQuerySearch(client)

	fmt.Println("\n=== Example 4: Date Filters ===")
	dateFilteredSearch(client)
}

func simpleSearch(client *perplexity.Client) {
	req := perplexity.NewSearchRequest("latest developments in quantum computing")

	// The request is validated by the client before being sent
	resp, err := client.SendSearchRequest(req)
	if err != nil {
		printError(err)
		return
	}

	fmt.Printf("Search %s found %d results:\n", resp.ID, resp.GetResultCount())
	for i, result := range resp.GetResults() {
		fmt.Printf("\n%d. %s\n", i+1, result.Title)
		fmt.Printf("   URL: %s\n", result.URL)
		if result.Snippet != nil {
			fmt.Printf("   Snippet: %s\n", truncate(*result.Snippet, 200))
		}
		if result.Date != nil {
			fmt.Printf("   Published: %s\n", *result.Date)
		}
		if result.LastUpdated != nil {
			fmt.Printf("   Last updated: %s\n", *result.LastUpdated)
		}
	}
}

func advancedSearch(client *perplexity.Client) {
	// Domain filters are either an allowlist ("github.com", "example.com/blog", ".gov")
	// or a denylist (every entry prefixed with "-"), never both.
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

	resp, err := client.SendSearchRequest(req)
	if err != nil {
		printError(err)
		return
	}

	if resp.ServerTime != nil {
		fmt.Printf("Server time: %s\n", *resp.ServerTime)
	}
	fmt.Printf("Found %d results:\n", resp.GetResultCount())
	for i, result := range resp.GetResults() {
		fmt.Printf("%d. %s\n", i+1, result.String())
	}
}

func multiQuerySearch(client *perplexity.Client) {
	// Up to 5 queries per request; billed as one request.
	queries := []string{
		"Go concurrency patterns",
		"Go performance optimization",
		"Go best practices",
	}

	req := perplexity.NewSearchRequest(
		queries,
		perplexity.WithSearchType(perplexity.SearchTypeFast), // lower latency and cost
		perplexity.WithSearchMaxResults(5),
	)

	resp, err := client.SendSearchRequest(req)
	if err != nil {
		printError(err)
		return
	}

	fmt.Printf("Found %d total results across %d queries:\n", resp.GetResultCount(), len(queries))
	for i, result := range resp.GetResults() {
		fmt.Printf("\n%d. %s\n", i+1, result.Title)
		fmt.Printf("   %s\n", result.URL)
	}
}

func dateFilteredSearch(client *perplexity.Client) {
	// Date filters cannot be combined with a recency filter.
	now := time.Now()
	req := perplexity.NewSearchRequest(
		"Go release notes",
		perplexity.WithSearchMaxResults(5),
		perplexity.WithSearchPublishedAfter(now.AddDate(-1, 0, 0)),
		perplexity.WithSearchUpdatedBefore(now),
	)

	resp, err := client.SendSearchRequest(req)
	if err != nil {
		printError(err)
		return
	}

	for i, result := range resp.GetResults() {
		fmt.Printf("%d. %s\n", i+1, result.String())
	}
}

func printError(err error) {
	var respErr *perplexity.ResponseError
	switch {
	case errors.As(err, &respErr) && respErr.StatusCode == http.StatusTooManyRequests:
		fmt.Printf("Rate limited, retry later: %v\n", err)
	case errors.As(err, &respErr):
		fmt.Printf("API error (HTTP %d): %v\n", respErr.StatusCode, err)
	default:
		fmt.Printf("Error: %v\n", err)
	}
}

func truncate(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n]) + "..."
}
