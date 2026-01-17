// Package main demonstrates the language_preference parameter for both Chat and Search APIs.
// This example shows how to request responses in different languages using ISO 639-1 codes.
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

	// Test 1: Chat completion with French language preference
	fmt.Println("=== Test 1: Chat with language_preference='fr' ===")
	msgs1 := perplexity.NewMessages()
	msgs1.AddUserMessage("What is the capital of France?")

	req1 := perplexity.NewCompletionRequest(
		perplexity.WithMessagesFromMessages(&msgs1),
		perplexity.WithModel("sonar"),
		perplexity.WithLanguagePreference("fr"),
	)

	resp1, err := client.SendCompletionRequest(req1)
	if err != nil {
		fmt.Printf("Completion request failed: %v\n", err)
		os.Exit(1)
	}

	content1 := resp1.GetLastContent()
	fmt.Printf("Question: What is the capital of France?\n")
	fmt.Printf("Response (French): %s\n", content1)

	// Test 2: Chat completion with English (US) language preference
	fmt.Println("\n=== Test 2: Chat with language_preference='en-US' ===")
	msgs2 := perplexity.NewMessages()
	msgs2.AddUserMessage("What is the capital of France?")

	req2 := perplexity.NewCompletionRequest(
		perplexity.WithMessagesFromMessages(&msgs2),
		perplexity.WithModel("sonar"),
		perplexity.WithLanguagePreference("en-US"),
	)

	resp2, err := client.SendCompletionRequest(req2)
	if err != nil {
		fmt.Printf("Completion request failed: %v\n", err)
		os.Exit(1)
	}

	content2 := resp2.GetLastContent()
	fmt.Printf("Question: What is the capital of France?\n")
	fmt.Printf("Response (English): %s\n", content2)

	// Test 3: Search with Spanish language preference
	fmt.Println("\n=== Test 3: Search with language_preference='es' ===")
	searchReq := perplexity.NewSearchRequest(
		"latest news technology",
		perplexity.WithSearchMaxResults(3),
		perplexity.WithSearchLanguagePreference("es"),
		perplexity.WithSearchReturnSnippets(true),
	)

	searchResp, err := client.SendSearchRequest(searchReq)
	if err != nil {
		fmt.Printf("Search request failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Found %d Spanish results\n", len(searchResp.Results))
	for i, result := range searchResp.Results {
		fmt.Printf("  %d. %s\n", i+1, result.Title)
		fmt.Printf("     URL: %s\n", result.URL)
		if result.Snippet != nil && len(*result.Snippet) > 100 {
			fmt.Printf("     Snippet: %s...\n", (*result.Snippet)[:100])
		} else if result.Snippet != nil {
			fmt.Printf("     Snippet: %s\n", *result.Snippet)
		}
		fmt.Println()
	}

	// Test 4: Search with Japanese language preference
	fmt.Println("=== Test 4: Search with language_preference='ja' ===")
	searchReq2 := perplexity.NewSearchRequest(
		"latest technology news",
		perplexity.WithSearchMaxResults(3),
		perplexity.WithSearchLanguagePreference("ja"),
		perplexity.WithSearchReturnSnippets(true),
	)

	searchResp2, err := client.SendSearchRequest(searchReq2)
	if err != nil {
		fmt.Printf("Search request failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Found %d Japanese results\n", len(searchResp2.Results))
	for i, result := range searchResp2.Results {
		fmt.Printf("  %d. %s\n", i+1, result.Title)
		fmt.Printf("     URL: %s\n", result.URL)
		if result.Snippet != nil && len(*result.Snippet) > 100 {
			fmt.Printf("     Snippet: %s...\n", (*result.Snippet)[:100])
		} else if result.Snippet != nil {
			fmt.Printf("     Snippet: %s\n", *result.Snippet)
		}
		fmt.Println()
	}

	fmt.Println("✅ Language preference test completed successfully")
}

/*
Example usage:

1. Set your API key:
   export PPLX_API_KEY=your_api_key_here

2. Build the example:
   go build -o bin/language-preference ./cmd/language-preference

3. Run the example:
   ./bin/language-preference

This example demonstrates:
- Setting language preference for Chat Completions API (French, English)
- Setting language preference for Search API (Spanish, Japanese)
- Valid ISO 639-1 language codes (2-letter codes)
- Valid ISO 639-1 with country codes (e.g., "en-US", "fr-CA")

Supported language preference formats:
- Two-letter language code: "en", "fr", "es", "ja", "de", etc.
- Language code with country: "en-US", "en-GB", "fr-CA", "es-MX", etc.

Note: Language preference works with sonar models for both Chat and Search APIs.
*/
