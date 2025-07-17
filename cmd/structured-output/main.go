// Package main demonstrates structured output capabilities of the Perplexity API.
// This example shows two types of structured output:
// 1. JSON Schema format - for structured data responses
// 2. Regex format - for pattern-matching responses
// Note: Structured output requires the "sonar" model.
package main

import (
	"fmt"
	"os"

	"github.com/sgaunet/perplexity-go/v2"
	"github.com/sgaunet/perplexity-go/v2/cmd/shared"
)

func main() {
	// Initialize the client with API key from environment variable
	client := perplexity.NewClient(os.Getenv("PPLX_API_KEY"))
	validator := perplexity.NewRequestValidator()

	// Demonstrate JSON Schema structured output
	demonstrateJSONSchemaOutput(client, validator)
	shared.PrintBreakToConsole()

	// Demonstrate Regex structured output
	demonstrateRegexOutput(client, validator)
}

func demonstrateJSONSchemaOutput(client *perplexity.Client, validator *perplexity.RequestValidator) {
	fmt.Println("=== JSON Schema Structured Output Example ===")

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

	// Create a message requesting structured information
	msg := []perplexity.Message{
		{
			Role:    "user",
			Content: "Tell me about Albert Einstein. Please format your response as a JSON object with name, age at death, and profession.",
		},
	}

	// Create a request with JSON schema structured output
	req := perplexity.NewCompletionRequest(
		perplexity.WithMessages(msg),
		perplexity.WithModel("sonar"), // Required: structured output only works with "sonar" model
		perplexity.WithJSONSchemaResponseFormat(personSchema),
	)

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

	// Print the structured JSON response
	fmt.Println("JSON Schema Response:")
	fmt.Println(res.GetLastContent())
}

func demonstrateRegexOutput(client *perplexity.Client, validator *perplexity.RequestValidator) {
	fmt.Println("=== Regex Structured Output Example ===")

	// Create a message requesting a specific pattern (IP address)
	msg := []perplexity.Message{
		{
			Role:    "user",
			Content: "What is the IP address of Google's primary DNS server? Please respond with just the IP address in the format x.x.x.x",
		},
	}

	// Create a request with regex structured output for IP addresses
	// This regex matches IPv4 addresses in the format xxx.xxx.xxx.xxx
	req := perplexity.NewCompletionRequest(
		perplexity.WithMessages(msg),
		perplexity.WithModel("sonar"), // Required: structured output only works with "sonar" model
		perplexity.WithRegexResponseFormat(`\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}`),
	)

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

	// Print the regex-validated response
	fmt.Println("Regex Response:")
	fmt.Println(res.GetLastContent())
	fmt.Println("\nNote: The response is guaranteed to match the specified regex pattern.")
}