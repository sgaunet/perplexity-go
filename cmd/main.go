// Package main provides the command-line interface for the Perplexity API client.
package main

import (
	"fmt"
	"os"
	"sync"

	"github.com/sgaunet/perplexity-go/v2"
)

// This example demonstrates how to create a completion request with web search options.
func main() {
	client := perplexity.NewClient(os.Getenv("PPLX_API_KEY"))
	validator := perplexity.NewRequestValidator()

	demonstrateBaseCompletion(client, validator)
	printBreakToConsole()

	demonstrateCompletionWithStructuredOutput(client, validator)
	printBreakToConsole()

	demonstrateCompletionWithImages(client, validator)
	printBreakToConsole()

	demonstrateCompletionWithServerSentEvents(client, validator)

}

func demonstrateBaseCompletion(client *perplexity.Client, validator *perplexity.RequestValidator) {
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

	// Print the response and citations
	fmt.Println("=== Response ===")
	fmt.Println(res.GetLastContent())

	if len(res.GetCitations()) > 0 {
		fmt.Println("\n=== Citations ===")
		for i, c := range res.GetCitations() {
			fmt.Printf("%d. %s\n", i+1, c)
		}
	}
	if len(res.GetSearchResults()) > 0 {
		fmt.Println("\n=== Search Results ===")
		for i, sr := range res.GetSearchResults() {
			fmt.Printf("%d. %s\n", i+1, sr.String())
		}
	}
	fmt.Println(res.GetImages())
}

func demonstrateCompletionWithStructuredOutput(client *perplexity.Client, validator *perplexity.RequestValidator) {

	// Example of structured output with JSON Schema
	fmt.Println("\n=== JSON Schema Structured Output Example ===")

	// Define a JSON schema for structured response
	personSchema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"name": map[string]interface{}{
				"type":        "string",
				"description": "The person's full name",
			},
			"age": map[string]interface{}{
				"type":        "integer",
				"description": "The person's age in years",
			},
			"profession": map[string]interface{}{
				"type":        "string",
				"description": "The person's job or profession",
			},
		},
		"required": []string{"name", "age", "profession"},
	}

	structuredMsg := []perplexity.Message{
		{
			Role:    "user",
			Content: "Tell me about Albert Einstein. Please format your response as a JSON object with name, age at death, and profession.",
		},
	}

	// Create a request with JSON schema structured output
	structuredReq := perplexity.NewCompletionRequest(
		perplexity.WithMessages(structuredMsg),
		perplexity.WithModel("sonar"), // Note: structured output only works with "sonar" model
		perplexity.WithJSONSchemaResponseFormat(personSchema),
	)

	if err := validator.ValidateRequest(structuredReq); err != nil {
		fmt.Printf("Structured request validation error: %v\n", err)
		os.Exit(1)
	}

	structuredRes, err := client.SendCompletionRequest(structuredReq)
	if err != nil {
		fmt.Printf("Structured API error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("JSON Schema Response:")
	fmt.Println(structuredRes.GetLastContent())

	// Example of structured output with Regex
	fmt.Println("\n=== Regex Structured Output Example ===")

	regexMsg := []perplexity.Message{
		{
			Role:    "user",
			Content: "What is the IP address of Google's primary DNS server? Please respond with just the IP address in the format x.x.x.x",
		},
	}

	// Create a request with regex structured output for IP addresses
	regexReq := perplexity.NewCompletionRequest(
		perplexity.WithMessages(regexMsg),
		perplexity.WithModel("sonar"),
		perplexity.WithRegexResponseFormat(`\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}`),
	)

	if err := validator.ValidateRequest(regexReq); err != nil {
		fmt.Printf("Regex request validation error: %v\n", err)
		os.Exit(1)
	}

	regexRes, err := client.SendCompletionRequest(regexReq)
	if err != nil {
		fmt.Printf("Regex API error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Regex Response:")
	fmt.Println(regexRes.GetLastContent())
}

func demonstrateCompletionWithImages(client *perplexity.Client, validator *perplexity.RequestValidator) {
	msg := []perplexity.Message{
		{
			Role:    "user",
			Content: "Find some photos of the Eiffel Tower",
		},
	}

	req := perplexity.NewCompletionRequest(
		perplexity.WithMessages(msg),
		perplexity.WithSearchRecencyFilter(""),
		perplexity.WithReturnImages(true),
	)

	if err := validator.ValidateRequest(req); err != nil {
		fmt.Printf("Images request validation error: %v\n", err)
		os.Exit(1)
	}

	imagesRes, err := client.SendCompletionRequest(req)
	if err != nil {
		fmt.Printf("Images API error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Images Response:")
	for _, img := range imagesRes.GetImages() {
		fmt.Println(img.String())
	}
}

// Support also server-sent events
func demonstrateCompletionWithServerSentEvents(client *perplexity.Client, validator *perplexity.RequestValidator) {

	msg := []perplexity.Message{
		{
			Role:    "user",
			Content: "What are the latest developments in AI?",
		},
	}
	req := perplexity.NewCompletionRequest(perplexity.WithMessages(msg), perplexity.WithStream(true))
	err := validator.ValidateRequest(req)
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

	wg.Wait()
	fmt.Println("----------------")
	fmt.Println(fullResponse.GetLastContent())
}

// printBreakToConsole prints a visual break in the console output for readability.
func printBreakToConsole() {
	fmt.Println()
	fmt.Println("*************")
	fmt.Println()
}
