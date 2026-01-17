// Package main demonstrates the max_tokens parameter for the Search API.
// This example shows how controlling max_tokens affects the length of search result snippets.
package main

import (
	"fmt"
	"os"

	"github.com/sgaunet/perplexity-go/v2"
)

func main() {
	// Initialize the client with API key from environment variable
	apiKey := os.Getenv("PPLX_API_KEY")
	if apiKey == "" {
		fmt.Println("Error: PPLX_API_KEY environment variable is required")
		os.Exit(1)
	}

	client := perplexity.NewClient(apiKey)

	// Test 1: Search with max_tokens = 500
	fmt.Println("=== Test 1: Search with max_tokens=500 ===")
	req1 := perplexity.NewSearchRequest(
		"What is quantum computing?",
		perplexity.WithSearchMaxResults(3),
		perplexity.WithSearchMaxTokens(500),
		perplexity.WithSearchReturnSnippets(true),
	)

	resp1, err := client.SendSearchRequest(req1)
	if err != nil {
		fmt.Printf("Search request failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Found %d results with max_tokens=500\n", len(resp1.Results))
	for i, result := range resp1.Results {
		snippetLen := 0
		if result.Snippet != nil {
			snippetLen = len(*result.Snippet)
		}
		fmt.Printf("  Result %d: %s\n", i+1, result.Title)
		fmt.Printf("           URL: %s\n", result.URL)
		fmt.Printf("           Snippet length: %d chars\n", snippetLen)
		if result.Snippet != nil && len(*result.Snippet) > 100 {
			fmt.Printf("           Preview: %s...\n", (*result.Snippet)[:100])
		} else if result.Snippet != nil {
			fmt.Printf("           Preview: %s\n", *result.Snippet)
		}
		fmt.Println()
	}

	// Test 2: Search with max_tokens = 1000
	fmt.Println("=== Test 2: Search with max_tokens=1000 ===")
	req2 := perplexity.NewSearchRequest(
		"What is quantum computing?",
		perplexity.WithSearchMaxResults(3),
		perplexity.WithSearchMaxTokens(1000),
		perplexity.WithSearchReturnSnippets(true),
	)

	resp2, err := client.SendSearchRequest(req2)
	if err != nil {
		fmt.Printf("Search request failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Found %d results with max_tokens=1000\n", len(resp2.Results))
	for i, result := range resp2.Results {
		snippetLen := 0
		if result.Snippet != nil {
			snippetLen = len(*result.Snippet)
		}
		fmt.Printf("  Result %d: %s\n", i+1, result.Title)
		fmt.Printf("           URL: %s\n", result.URL)
		fmt.Printf("           Snippet length: %d chars\n", snippetLen)
		if result.Snippet != nil && len(*result.Snippet) > 100 {
			fmt.Printf("           Preview: %s...\n", (*result.Snippet)[:100])
		} else if result.Snippet != nil {
			fmt.Printf("           Preview: %s\n", *result.Snippet)
		}
		fmt.Println()
	}

	// Test 3: Search without max_tokens (default)
	fmt.Println("=== Test 3: Search without max_tokens (default) ===")
	req3 := perplexity.NewSearchRequest(
		"What is quantum computing?",
		perplexity.WithSearchMaxResults(3),
		perplexity.WithSearchReturnSnippets(true),
	)

	resp3, err := client.SendSearchRequest(req3)
	if err != nil {
		fmt.Printf("Search request failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Found %d results with default max_tokens\n", len(resp3.Results))
	for i, result := range resp3.Results {
		snippetLen := 0
		if result.Snippet != nil {
			snippetLen = len(*result.Snippet)
		}
		fmt.Printf("  Result %d: %s\n", i+1, result.Title)
		fmt.Printf("           URL: %s\n", result.URL)
		fmt.Printf("           Snippet length: %d chars\n", snippetLen)
		if result.Snippet != nil && len(*result.Snippet) > 100 {
			fmt.Printf("           Preview: %s...\n", (*result.Snippet)[:100])
		} else if result.Snippet != nil {
			fmt.Printf("           Preview: %s\n", *result.Snippet)
		}
		fmt.Println()
	}

	fmt.Println("✅ Max tokens test completed successfully")
}

/*
Example usage:

1. Set your API key:
   export PPLX_API_KEY=your_api_key_here

2. Build the example:
   go build -o bin/search-max-tokens ./cmd/search-max-tokens

3. Run the example:
   ./bin/search-max-tokens

This example demonstrates:
- How max_tokens controls the extraction length per page in Search API results
- Comparing snippet lengths between different max_tokens values (500, 1000, default)
- The impact of max_tokens on result detail and token consumption

Note: The max_tokens parameter controls how much content is extracted from each
search result page, allowing you to balance between detail and token usage.
*/
