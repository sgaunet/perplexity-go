// Package main demonstrates basic completion requests with the Perplexity API.
// This example shows how to use web search options, including search context size
// and user location, and how to handle responses with citations and search results.
package main

import (
	"fmt"
	"os"

	"github.com/sgaunet/perplexity-go/v2"
)

func main() {
	// Initialize the client with API key from environment variable
	client := perplexity.NewClient(os.Getenv("PPLX_API_KEY"))
	validator := perplexity.NewRequestValidator()

	// Example message that would benefit from web search
	msg := []perplexity.Message{
		{
			Role:    "user",
			Content: "What are the latest developments in AI?",
		},
	}

	// Create web search options with context size and user location
	webSearchOpts := &perplexity.WebSearchOptions{
		SearchContextSize: "high", // Use high for comprehensive answers
		UserLocation: &perplexity.UserLocation{
			Latitude:  48.8566, // Paris coordinates as an example
			Longitude: 2.3522,
			Country:   "FR",
		},
	}

	// Create request with messages and web search options
	req := perplexity.NewCompletionRequest(
		perplexity.WithMessages(msg),
		perplexity.WithWebSearchOptions(webSearchOpts),
	)

	// Alternatively, you can set options individually:
	// req := perplexity.NewCompletionRequest(
	// 	perplexity.WithMessages(msg),
	// 	perplexity.WithSearchContextSize("high"),
	// 	perplexity.WithUserLocation(48.8566, 2.3522, "FR"),
	// )

	// Validate the request
	if err := validator.ValidateRequest(req); err != nil {
		fmt.Printf("Validation error: %v\n", err)
		os.Exit(1)
	}

	// Send the request
	res, err := client.SendCompletionRequest(req)
	if err != nil {
		fmt.Printf("API error: %v\n", err)
		os.Exit(1)
	}

	// Print the response
	fmt.Println("=== Response ===")
	fmt.Println(res.GetLastContent())

	// Print citations if available
	if len(res.GetCitations()) > 0 {
		fmt.Println("\n=== Citations ===")
		for i, c := range res.GetCitations() {
			fmt.Printf("%d. %s\n", i+1, c)
		}
	}

	// Print search results if available
	if len(res.GetSearchResults()) > 0 {
		fmt.Println("\n=== Search Results ===")
		for i, sr := range res.GetSearchResults() {
			fmt.Printf("%d. %s\n", i+1, sr.String())
		}
	}

	// Print any images returned
	if images := res.GetImages(); len(images) > 0 {
		fmt.Println("\n=== Images ===")
		for i, img := range images {
			fmt.Printf("%d. %s\n", i+1, img.String())
		}
	}
}