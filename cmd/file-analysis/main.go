// Package main demonstrates file attachment capabilities of the Perplexity API.
// This example shows how to analyze documents (PDF, DOC, DOCX, TXT, RTF) using both
// public URLs and local files.
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
	validator := perplexity.NewRequestValidator()

	// Example 1: Analyze document from URL
	fmt.Println("=== Example 1: Analyzing document from URL ===")
	analyzeDocumentFromURL(client, validator)

	fmt.Println() // Separator

	// Example 2: Analyze local document file
	// Uncomment and update the path to analyze a local file
	// fmt.Println("=== Example 2: Analyzing local document ===\n")
	// analyzeLocalDocument(client, validator, "./sample.pdf")
}

// analyzeDocumentFromURL demonstrates analyzing a document from a public URL.
func analyzeDocumentFromURL(client *perplexity.Client, validator *perplexity.RequestValidator) {
	// Create a Messages object with file attachment
	msgs := perplexity.NewMessages()

	// Add a user message with a file URL
	// Example using a publicly accessible PDF document
	documentURL := "https://arxiv.org/pdf/1706.03762.pdf" // "Attention Is All You Need" paper
	err := msgs.AddUserMessageWithFile(
		"Please provide a brief summary of this paper's main contributions",
		documentURL,
		"attention-paper.pdf",
	)
	if err != nil {
		fmt.Printf("Error adding file message: %v\n", err)
		os.Exit(1)
	}

	// Create completion request
	req := perplexity.NewCompletionRequest(
		perplexity.WithMessagesFromMessages(&msgs),
		perplexity.WithModel("sonar-pro"), // Use sonar-pro for document analysis
		perplexity.WithMaxTokens(2000),
	)

	// Validate the request
	if err := validator.ValidateRequest(req); err != nil {
		fmt.Printf("Validation error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Document URL:", documentURL)
	fmt.Println("Query: Please provide a brief summary of this paper's main contributions")
	fmt.Println("\nSending request to Perplexity API...")
	fmt.Println()

	// Send the request
	res, err := client.SendCompletionRequest(req)
	if err != nil {
		fmt.Printf("API error: %v\n", err)
		os.Exit(1)
	}

	// Print the analysis
	fmt.Println("=== Analysis ===")
	fmt.Println(res.GetLastContent())

	// Show citations if available
	if len(res.GetCitations()) > 0 {
		fmt.Println("\n=== Citations ===")
		for i, citation := range res.GetCitations() {
			fmt.Printf("%d. %s\n", i+1, citation)
		}
	}

	// Show token usage
	fmt.Printf("\n=== Token Usage ===\n")
	fmt.Printf("Prompt tokens: %d\n", res.Usage.PromptTokens)
	fmt.Printf("Completion tokens: %d\n", res.Usage.CompletionTokens)
	fmt.Printf("Total tokens: %d\n", res.Usage.TotalTokens)
}

// analyzeLocalDocument demonstrates analyzing a local document file.
// Uncomment the call in main() to use this function.
//
//nolint:unused // This function is intentionally included for documentation purposes
func analyzeLocalDocument(client *perplexity.Client, validator *perplexity.RequestValidator, filePath string) {
	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		fmt.Printf("Error: File not found: %s\n", filePath)
		os.Exit(1)
	}

	// Create a Messages object
	msgs := perplexity.NewMessages()

	// Add a user message with local file
	err := msgs.AddUserMessageWithFileFromPath(
		"What are the key points discussed in this document?",
		filePath,
	)
	if err != nil {
		fmt.Printf("Error adding file message: %v\n", err)
		os.Exit(1)
	}

	// Create completion request
	req := perplexity.NewCompletionRequest(
		perplexity.WithMessagesFromMessages(&msgs),
		perplexity.WithModel("sonar-pro"),
		perplexity.WithMaxTokens(2000),
	)

	// Validate the request
	if err := validator.ValidateRequest(req); err != nil {
		fmt.Printf("Validation error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Local file:", filePath)
	fmt.Println("Query: What are the key points discussed in this document?")
	fmt.Println("\nSending request to Perplexity API...")
	fmt.Println()

	// Send the request
	res, err := client.SendCompletionRequest(req)
	if err != nil {
		fmt.Printf("API error: %v\n", err)
		os.Exit(1)
	}

	// Print the analysis
	fmt.Println("=== Analysis ===")
	fmt.Println(res.GetLastContent())

	// Show citations if available
	if len(res.GetCitations()) > 0 {
		fmt.Println("\n=== Citations ===")
		for i, citation := range res.GetCitations() {
			fmt.Printf("%d. %s\n", i+1, citation)
		}
	}

	// Show token usage
	fmt.Printf("\n=== Token Usage ===\n")
	fmt.Printf("Prompt tokens: %d\n", res.Usage.PromptTokens)
	fmt.Printf("Completion tokens: %d\n", res.Usage.CompletionTokens)
	fmt.Printf("Total tokens: %d\n", res.Usage.TotalTokens)
}

/*
Example usage:

1. Set your API key:
   export PPLX_API_KEY=your_api_key_here

2. Build the example:
   go build -o bin/file-analysis ./cmd/file-analysis

3. Run the example:
   ./bin/file-analysis

4. To analyze a local file, uncomment the analyzeLocalDocument call in main()
   and update the file path.

Supported file formats:
- PDF (.pdf)
- Microsoft Word (.doc, .docx)
- Text files (.txt)
- Rich Text Format (.rtf)

Note: Files must be under 50MB in size.
*/
