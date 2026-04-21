// Package main demonstrates Server-Sent Events (SSE) streaming with the Perplexity API.
// This example shows how to handle real-time streaming responses with StreamCompletion:
// the client call runs in its own goroutine, the main goroutine consumes the channel,
// and an error channel hands the final error back to main.
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

	// Create a message for streaming
	msg := []perplexity.Message{
		{
			Role:    "user",
			Content: "What are the latest developments in AI?",
		},
	}

	// Create a request with streaming enabled
	req := perplexity.NewCompletionRequest(
		perplexity.WithMessages(msg),
		perplexity.WithStream(true), // Enable SSE streaming
	)

	// Validate the request
	if err := validator.ValidateRequest(req); err != nil {
		fmt.Printf("Validation error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("=== Starting SSE Streaming ===")
	fmt.Println("Query: What are the latest developments in AI?")
	fmt.Println()
	fmt.Println("Streaming response:")
	fmt.Println("---")

	chResponses := make(chan perplexity.CompletionResponse)
	errCh := make(chan error, 1)
	fullResponse := perplexity.CompletionResponse{}

	// Run the streaming call in its own goroutine. StreamCompletion closes the
	// channel before returning, so the range loop below exits cleanly.
	go func() {
		errCh <- client.StreamCompletion(req, chResponses)
	}()

	// Process streaming responses as they arrive
	tokenCount := 0
	var lastContent string
	for msg := range chResponses {
		// Each message contains the accumulated response
		fullResponse = msg

		// Display incremental content (only the new part)
		currentContent := msg.GetLastContent()
		if len(currentContent) > len(lastContent) {
			newContent := currentContent[len(lastContent):]
			fmt.Print(newContent)
			lastContent = currentContent
		}

		tokenCount++
	}

	if err := <-errCh; err != nil {
		fmt.Printf("\nError during streaming: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\n---")
	fmt.Printf("Streaming completed. Received %d response chunks.\n", tokenCount)

	// Show additional information if available
	if len(fullResponse.GetCitations()) > 0 {
		fmt.Println("\n=== Citations ===")
		for i, c := range fullResponse.GetCitations() {
			fmt.Printf("%d. %s\n", i+1, c)
		}
	}

	if len(fullResponse.GetSearchResults()) > 0 {
		fmt.Println("\n=== Search Results ===")
		for i, sr := range fullResponse.GetSearchResults() {
			fmt.Printf("%d. %s\n", i+1, sr.String())
		}
	}

	// Display usage information
	if fullResponse.Usage.TotalTokens > 0 {
		fmt.Println("\n=== Token Usage ===")
		fmt.Printf("Prompt tokens: %d\n", fullResponse.Usage.PromptTokens)
		fmt.Printf("Completion tokens: %d\n", fullResponse.Usage.CompletionTokens)
		fmt.Printf("Total tokens: %d\n", fullResponse.Usage.TotalTokens)
	}
}
