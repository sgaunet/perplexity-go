package main

import (
	"fmt"
	"os"

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

	// Example 1: Simple search query
	fmt.Println("=== Example 1: Simple Search ===")
	simpleSearch(client)

	fmt.Println("\n=== Example 2: Advanced Search with Options ===")
	advancedSearch(client)

	fmt.Println("\n=== Example 3: Multi-query Search ===")
	multiQuerySearch(client)
}

func simpleSearch(client *perplexity.Client) {
	// Create a simple search request
	req := perplexity.NewSearchRequest("latest developments in quantum computing")

	// Validate request
	validator := perplexity.NewSearchRequestValidator()
	if err := validator.ValidateSearchRequest(req); err != nil {
		fmt.Printf("Validation error: %v\n", err)
		return
	}

	// Send request
	resp, err := client.SendSearchRequest(req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Display results
	fmt.Printf("Found %d results:\n", resp.GetResultCount())
	for i, result := range resp.GetResults() {
		fmt.Printf("\n%d. %s\n", i+1, result.Title)
		fmt.Printf("   URL: %s\n", result.URL)
		if result.Snippet != nil {
			fmt.Printf("   Snippet: %s\n", *result.Snippet)
		}
		if result.Date != nil {
			fmt.Printf("   Date: %s\n", *result.Date)
		}
		if result.Score != nil {
			fmt.Printf("   Relevance Score: %.2f\n", *result.Score)
		}
	}
}

func advancedSearch(client *perplexity.Client) {
	// Create search with options
	// Note: Domain filters should be specific domains (e.g., "github.com")
	// rather than wildcard patterns (e.g., "*.example.com")
	req := perplexity.NewSearchRequest(
		"best Go web frameworks 2025",
		perplexity.WithSearchMaxResults(10),
		perplexity.WithSearchReturnImages(true),
		perplexity.WithSearchReturnSnippets(true),
		perplexity.WithSearchCountry("US"),
		perplexity.WithSearchDomains([]string{"golang.org", "github.com"}),
	)

	// Validate request
	validator := perplexity.NewSearchRequestValidator()
	if err := validator.ValidateSearchRequest(req); err != nil {
		fmt.Printf("Validation error: %v\n", err)
		return
	}

	// Send request
	resp, err := client.SendSearchRequest(req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Display results
	fmt.Printf("Found %d results:\n", resp.GetResultCount())
	for i, result := range resp.GetResults() {
		fmt.Printf("\n%d. %s\n", i+1, result.String())
		if result.Images != nil && len(*result.Images) > 0 {
			fmt.Printf("   Images: %d\n", len(*result.Images))
			for j, img := range *result.Images {
				if j < 2 { // Show first 2 images
					fmt.Printf("     - %s\n", img.String())
				}
			}
		}
	}
}

func multiQuerySearch(client *perplexity.Client) {
	// Create multi-query search request
	queries := []string{
		"Go concurrency patterns",
		"Go performance optimization",
		"Go best practices 2025",
	}

	req := perplexity.NewSearchRequest(
		queries,
		perplexity.WithSearchMaxResults(5),
	)

	// Validate request
	validator := perplexity.NewSearchRequestValidator()
	if err := validator.ValidateSearchRequest(req); err != nil {
		fmt.Printf("Validation error: %v\n", err)
		return
	}

	// Send request
	resp, err := client.SendSearchRequest(req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Display results
	fmt.Printf("Found %d total results across %d queries:\n", resp.GetResultCount(), len(queries))
	for i, result := range resp.GetResults() {
		fmt.Printf("\n%d. %s\n", i+1, result.Title)
		fmt.Printf("   %s\n", result.URL)
	}
}
