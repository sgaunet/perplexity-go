// Package main demonstrates image search capabilities of the Perplexity API.
// This example shows how to request and retrieve images in the API response.
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

	// Create a message requesting images
	msg := []perplexity.Message{
		{
			Role:    "user",
			Content: "Find some photos of the Eiffel Tower",
		},
	}

	// Create a request with image return enabled
	req := perplexity.NewCompletionRequest(
		perplexity.WithMessages(msg),
		perplexity.WithReturnImages(true),        // Enable image returns
		perplexity.WithSearchRecencyFilter(""), // Clear default recency filter (incompatible with images)
	)

	// Validate the request
	if err := validator.ValidateRequest(req); err != nil {
		fmt.Printf("Validation error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("=== Searching for images ===")
	fmt.Println("Query: Find some photos of the Eiffel Tower")
	fmt.Println()

	// Send the request
	res, err := client.SendCompletionRequest(req)
	if err != nil {
		fmt.Printf("API error: %v\n", err)
		os.Exit(1)
	}

	// Print the text response
	fmt.Println("=== Response ===")
	fmt.Println(res.GetLastContent())

	// Print returned images
	images := res.GetImages()
	if len(images) > 0 {
		fmt.Printf("\n=== Images Found: %d ===\n", len(images))
		for i, img := range images {
			fmt.Printf("\nImage %d:\n", i+1)
			fmt.Printf("  Image URL: %s\n", img.ImageURL)
			if img.OriginURL != "" {
				fmt.Printf("  Origin URL: %s\n", img.OriginURL)
			}
			if img.Width > 0 && img.Height > 0 {
				fmt.Printf("  Dimensions: %dx%d\n", img.Width, img.Height)
			}
		}
	} else {
		fmt.Println("\nNo images were returned in the response.")
	}

	// Also show citations if available
	if len(res.GetCitations()) > 0 {
		fmt.Println("\n=== Citations ===")
		for i, c := range res.GetCitations() {
			fmt.Printf("%d. %s\n", i+1, c)
		}
	}
}