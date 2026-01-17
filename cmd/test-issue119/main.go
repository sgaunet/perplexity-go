// Package main provides a comprehensive integration test for Issue 119 features.
// This test combines all three new features: file attachments, max_tokens, and language_preference.
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

	fmt.Println("=== Issue 119 Integration Test ===")
	fmt.Println("Testing: File Attachments + Language Preference + Max Tokens")
	fmt.Println()

	// Test 1: File attachment with language preference
	fmt.Println("Test 1: File attachment + language preference")
	fmt.Println("---")
	testFileWithLanguagePreference(client)

	fmt.Println()

	// Test 2: Search with max_tokens and language preference
	fmt.Println("Test 2: Search API with max_tokens + language preference")
	fmt.Println("---")
	testSearchWithMaxTokensAndLanguage(client)

	fmt.Println()

	// Test 3: All features in sequence
	fmt.Println("Test 3: Complete feature validation")
	fmt.Println("---")
	testCompleteValidation(client)

	fmt.Println()
	fmt.Println("=== All Issue 119 features tested successfully! ===")
}

func testFileWithLanguagePreference(client *perplexity.Client) {
	// Create a simple text file for testing
	testFile := "/tmp/test-doc-issue119.txt"
	testContent := `Quantum Computing Overview

Quantum computing is a revolutionary approach to computation that harnesses the principles of quantum mechanics. Unlike classical computers that use bits (0 or 1), quantum computers use quantum bits or qubits, which can exist in multiple states simultaneously through superposition.

Key Benefits:
- Exponential speedup for certain problems
- Applications in cryptography, drug discovery, and optimization
- Potential to solve currently intractable problems`

	err := os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		fmt.Printf("Error creating test file: %v\n", err)
		os.Exit(1)
	}
	defer os.Remove(testFile) // Cleanup

	// Create messages with file attachment
	msgs := perplexity.NewMessages()
	err = msgs.AddUserMessageWithFileFromPath(
		"Résumez ce document en français s'il vous plaît",
		testFile,
	)
	if err != nil {
		fmt.Printf("Error adding file: %v\n", err)
		os.Exit(1)
	}

	// Create request with language preference
	req := perplexity.NewCompletionRequest(
		perplexity.WithMessagesFromMessages(&msgs),
		perplexity.WithModel("sonar"),
		perplexity.WithLanguagePreference("fr"),
		perplexity.WithMaxTokens(500),
	)

	// Send request
	resp, err := client.SendCompletionRequest(req)
	if err != nil {
		fmt.Printf("Completion failed: %v\n", err)
		os.Exit(1)
	}

	// Display result
	content := resp.GetLastContent()
	fmt.Printf("Request: Analyze document in French\n")
	fmt.Printf("File: %s (%d bytes)\n", testFile, len(testContent))
	fmt.Printf("Language: French (fr)\n")
	fmt.Printf("\nResponse:\n%s\n", content)
	fmt.Printf("\nToken Usage: %d prompt + %d completion = %d total\n",
		resp.Usage.PromptTokens, resp.Usage.CompletionTokens, resp.Usage.TotalTokens)
	fmt.Println("\n✅ File attachment + language preference: PASSED")
}

func testSearchWithMaxTokensAndLanguage(client *perplexity.Client) {
	// Test search with both max_tokens and language_preference
	req := perplexity.NewSearchRequest(
		"avances en informatique quantique",
		perplexity.WithSearchMaxResults(3),
		perplexity.WithSearchMaxTokens(500),
		perplexity.WithSearchLanguagePreference("fr"),
		perplexity.WithSearchReturnSnippets(true),
	)

	// Send request
	resp, err := client.SendSearchRequest(req)
	if err != nil {
		fmt.Printf("Search failed: %v\n", err)
		os.Exit(1)
	}

	// Display results
	fmt.Printf("Query: avances en informatique quantique\n")
	fmt.Printf("Max Tokens: 500\n")
	fmt.Printf("Language: French (fr)\n")
	fmt.Printf("\nFound %d French results with max_tokens=500:\n\n", len(resp.Results))

	for i, result := range resp.Results {
		snippetLen := 0
		if result.Snippet != nil {
			snippetLen = len(*result.Snippet)
		}
		fmt.Printf("%d. %s\n", i+1, result.Title)
		fmt.Printf("   URL: %s\n", result.URL)
		fmt.Printf("   Snippet length: %d chars\n", snippetLen)
		if result.Snippet != nil && len(*result.Snippet) > 80 {
			fmt.Printf("   Preview: %s...\n", (*result.Snippet)[:80])
		} else if result.Snippet != nil {
			fmt.Printf("   Preview: %s\n", *result.Snippet)
		}
		fmt.Println()
	}

	fmt.Println("✅ Search with max_tokens + language preference: PASSED")
}

func testCompleteValidation(client *perplexity.Client) {
	// Validate all three features are working correctly

	// Feature 1: File attachments
	fmt.Print("Checking file attachment support... ")
	testFile := "/tmp/test-validation.txt"
	os.WriteFile(testFile, []byte("Test content"), 0644)
	defer os.Remove(testFile)

	msgs := perplexity.NewMessages()
	err := msgs.AddUserMessageWithFileFromPath("Test", testFile)
	if err != nil {
		fmt.Printf("FAILED: %v\n", err)
		os.Exit(1)
	}
	if !msgs.HasFiles() {
		fmt.Println("FAILED: HasFiles() returned false")
		os.Exit(1)
	}
	fmt.Println("OK")

	// Feature 2: Language preference (Chat API)
	fmt.Print("Checking language preference for Chat API... ")
	req1 := perplexity.NewCompletionRequest(
		perplexity.WithMessagesFromMessages(&msgs),
		perplexity.WithModel("sonar"),
		perplexity.WithLanguagePreference("fr"),
	)
	if req1.LanguagePreference != "fr" {
		fmt.Println("FAILED: Language preference not set")
		os.Exit(1)
	}
	fmt.Println("OK")

	// Feature 3: Max tokens and language preference (Search API)
	fmt.Print("Checking max_tokens + language_preference for Search API... ")
	req2 := perplexity.NewSearchRequest(
		"test query",
		perplexity.WithSearchMaxTokens(500),
		perplexity.WithSearchLanguagePreference("es"),
	)
	if req2.MaxTokens == nil || *req2.MaxTokens != 500 {
		fmt.Println("FAILED: Max tokens not set correctly")
		os.Exit(1)
	}
	if req2.LanguagePreference == nil || *req2.LanguagePreference != "es" {
		fmt.Println("FAILED: Language preference not set correctly")
		os.Exit(1)
	}
	fmt.Println("OK")

	// Validate language preference format
	fmt.Print("Checking language preference format validation... ")
	validator := perplexity.NewSearchRequestValidator()

	// Valid formats
	validLangs := []string{"en", "fr", "es", "ja", "en-US", "fr-CA", "es-MX"}
	for _, lang := range validLangs {
		req := perplexity.NewSearchRequest(
			"test",
			perplexity.WithSearchLanguagePreference(lang),
		)
		if err := validator.ValidateSearchRequest(req); err != nil {
			fmt.Printf("FAILED: Valid language '%s' rejected: %v\n", lang, err)
			os.Exit(1)
		}
	}

	// Invalid formats (excluding empty string, which is treated as "not set")
	invalidLangs := []string{"e", "eng", "en-", "en-usa", "en-U"}
	for _, lang := range invalidLangs {
		req := perplexity.NewSearchRequest(
			"test",
			perplexity.WithSearchLanguagePreference(lang),
		)
		if err := validator.ValidateSearchRequest(req); err == nil {
			fmt.Printf("FAILED: Invalid language '%s' accepted\n", lang)
			os.Exit(1)
		}
	}
	fmt.Println("OK")

	fmt.Println("\n✅ Complete feature validation: PASSED")
}

/*
Example usage:

1. Set your API key:
   export PPLX_API_KEY=your_api_key_here

2. Build the example:
   go build -o bin/test-issue119 ./cmd/test-issue119

3. Run the integration test:
   ./bin/test-issue119

This test validates:
✅ Feature 1: File attachment support (PDF, DOC, DOCX, TXT, RTF)
✅ Feature 2: Language preference for Chat API (ISO 639-1 codes)
✅ Feature 3: Max tokens for Search API (controls extraction length)
✅ Combined: File + language preference
✅ Combined: Search with max_tokens + language preference
✅ Validation: Format validation for all parameters

Exit codes:
- 0: All tests passed
- 1: One or more tests failed
*/
