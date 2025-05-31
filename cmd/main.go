package main

import (
	"fmt"
	"os"
	"sync"

	"github.com/sgaunet/perplexity-go/v2"
)

// This example demonstrates how to create a completion request with web search options
func main() {
	client := perplexity.NewClient(os.Getenv("PPLX_API_KEY"))

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
	if err := req.Validate(); err != nil {
		fmt.Printf("Validation error: %v\n", err)
		os.Exit(1)
	}

	// Send the request
	res, err := client.SendCompletionRequest(req)
	if err != nil {
		fmt.Printf("API error: %v\n", err)
		os.Exit(1)
	}

	// Print the response and citations
	fmt.Println("=== Response ===")
	fmt.Println(res.GetLastContent())

	if len(res.GetCitations()) > 0 {
		fmt.Println("\n=== Citations ===")
		for i, c := range res.GetCitations() {
			fmt.Printf("%d. %s\n", i+1, c)
		}
	}
	fmt.Println("*************")

	// Support also server-sent events
	req = perplexity.NewCompletionRequest(perplexity.WithMessages(msg), perplexity.WithStream(true))
	err = req.Validate()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	var wg sync.WaitGroup
	chResponses := make(chan perplexity.CompletionResponse, 5)
	fullResponse := perplexity.CompletionResponse{}

	waitAfterGoroutine := make(chan struct{})
	wg.Add(1)
	go func() {
		waitAfterGoroutine <- struct{}{}
		err = client.SendSSEHTTPRequest(&wg, req, chResponses)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
	}()

	<-waitAfterGoroutine
	for msg := range chResponses {
		fullResponse = msg
	}
	// perplexity.TreatSSEData(chResponses)
	wg.Wait()
	fmt.Println("----------------")
	fmt.Println(fullResponse.GetLastContent())
	// if err != nil {
	// 	fmt.Printf("Error: %v\n", err)
	// 	os.Exit(1)
	// }
}
