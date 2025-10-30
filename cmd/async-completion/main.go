// Package main demonstrates async completion requests with the Perplexity API.
// This example shows how to use the async API for long-running Sonar Deep Research tasks
// with proper polling, progress monitoring, and error handling.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/sgaunet/perplexity-go/v2"
)

func main() {
	// Initialize the client with API key from environment variable
	client := perplexity.NewClient(os.Getenv("PPLX_API_KEY"))
	validator := perplexity.NewAsyncJobRequestValidator()

	// Create messages for a complex research query
	messages := perplexity.NewMessages(
		perplexity.WithSystemMessage("You are a research assistant that provides comprehensive, well-researched answers with proper citations."),
	)

	err := messages.AddUserMessage("Conduct a comprehensive analysis of the latest developments in quantum computing, including recent breakthroughs, current challenges, commercial applications, and future prospects. Include information from academic papers, industry reports, and recent news.")
	if err != nil {
		fmt.Printf("Error creating messages: %v\n", err)
		os.Exit(1)
	}

	// Create web search options for comprehensive research
	webSearchOpts := &perplexity.WebSearchOptions{
		SearchContextSize: "high", // Use high context for deep research
		UserLocation: &perplexity.UserLocation{
			Country: "US", // Set to US for comprehensive coverage
		},
	}

	// Create async request with comprehensive settings
	req := perplexity.NewAsyncJobRequest(
		perplexity.WithAsyncMessagesObject(messages),
		perplexity.WithAsyncReasoningEffort("high"), // Use high reasoning for deep analysis
		perplexity.WithAsyncMaxTokens(4000),         // Allow for comprehensive response
		perplexity.WithAsyncWebSearchOptions(webSearchOpts),
		perplexity.WithAsyncSearchMode("academic"), // Prefer academic sources
		perplexity.WithAsyncTemperature(0.1),       // Low temperature for factual accuracy
	)

	// Debug: Print request structure
	fmt.Printf("Debug - Request Messages: %+v\n", req.Messages)
	fmt.Printf("Debug - Request Model: %s\n", req.Model)

	// Debug: Print JSON that will be sent
	jsonBytes, jsonErr := json.Marshal(req)
	if jsonErr != nil {
		fmt.Printf("JSON Marshal error: %v\n", jsonErr)
		os.Exit(1)
	}
	fmt.Printf("Debug - Generated JSON: %s\n", string(jsonBytes))

	// Validate the request
	if err := validator.Validate(req); err != nil {
		fmt.Printf("Validation error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("🚀 Creating async job for deep research...")

	// Create the async job
	job, err := client.CreateAsyncJob(req)
	if err != nil {
		fmt.Printf("Failed to create async job: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ Async job created successfully!\n")
	fmt.Printf("   Job ID: %s\n", job.ID)
	fmt.Printf("   Status: %s\n", job.Status)
	fmt.Printf("   Created: %s\n", job.CreatedAt.Format(time.RFC3339))
	fmt.Printf("   Expires: %s\n", job.ExpiresAt.Format(time.RFC3339))

	// Example 1: Simple polling with WaitForAsyncJob
	fmt.Println("\n📊 Waiting for job completion with simple polling...")

	// Configure custom polling options
	opts := &perplexity.AsyncPollingOptions{
		InitialInterval:   3 * time.Second,  // Start with 3-second intervals
		MaxInterval:       30 * time.Second, // Cap at 30 seconds
		BackoffMultiplier: 1.5,              // Gradual backoff
		MaxWaitTime:       10 * time.Minute, // Total timeout
		JitterEnabled:     true,             // Add randomness to avoid thundering herd
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer cancel()

	// Example 2: Advanced polling with progress monitoring
	fmt.Println("\n🔄 Using advanced polling with progress monitoring...")

	// Create progress channel
	progressChan := make(chan *perplexity.AsyncJobResponse, 10)

	// Start monitoring in a goroutine
	go func() {
		defer close(progressChan)

		monitor := client.NewAsyncJobMonitor(job.ID, opts)

		for {
			select {
			case <-ctx.Done():
				return
			case progress, ok := <-progressChan:
				if !ok {
					return
				}

				fmt.Printf("📈 Progress Update:\n")
				fmt.Printf("   Status: %s\n", progress.Status)
				fmt.Printf("   Elapsed: %s\n", monitor.GetElapsedTime().Round(time.Second))

				if progress.Progress != nil {
					if progress.Progress.Stage != "" {
						fmt.Printf("   Stage: %s\n", progress.Progress.Stage)
					}
					if progress.Progress.Percentage != nil {
						fmt.Printf("   Progress: %d%%\n", *progress.Progress.Percentage)

						// Estimate remaining time
						remaining := perplexity.EstimateRemainingTime(progress)
						if remaining > 0 {
							fmt.Printf("   Est. Remaining: %s\n", remaining.Round(time.Second))
						}
					}
				}
				fmt.Println()

				if progress.IsCompleted() || progress.IsExpired() {
					return
				}
			}
		}
	}()

	// Wait for completion with progress updates
	finalJob, err := client.WaitForAsyncJobWithProgress(ctx, job.ID, opts, progressChan)
	if err != nil {
		fmt.Printf("❌ Error waiting for job: %v\n", err)
		os.Exit(1)
	}

	// Handle final result
	fmt.Printf("🎉 Job completed!\n")
	fmt.Printf("   Final Status: %s\n", finalJob.Status)

	if finalJob.CompletedAt != nil {
		duration := finalJob.CompletedAt.Sub(finalJob.CreatedAt)
		fmt.Printf("   Total Duration: %s\n", duration.Round(time.Second))
	}

	// Get and display the result
	result, err := finalJob.GetResult()
	if err != nil {
		fmt.Printf("❌ Error getting result: %v\n", err)
		os.Exit(1)
	}

	// Print the comprehensive research response
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("📋 RESEARCH RESULTS")
	fmt.Println(strings.Repeat("=", 80))
	fmt.Println(result.GetLastContent())

	// Print citations if available
	if citations := result.GetCitations(); len(citations) > 0 {
		fmt.Println("\n" + strings.Repeat("=", 80))
		fmt.Println("📚 CITATIONS")
		fmt.Println(strings.Repeat("=", 80))
		for i, citation := range citations {
			fmt.Printf("%d. %s\n", i+1, citation)
		}
	}

	// Print search results if available
	if searchResults := result.GetSearchResults(); len(searchResults) > 0 {
		fmt.Println("\n" + strings.Repeat("=", 80))
		fmt.Println("🔍 SEARCH RESULTS")
		fmt.Println(strings.Repeat("=", 80))
		for i, sr := range searchResults {
			fmt.Printf("%d. %s\n", i+1, sr.String())
		}
	}

	// Print related questions if available
	if relatedQuestions := result.GetRelatedQuestions(); len(relatedQuestions) > 0 {
		fmt.Println("\n" + strings.Repeat("=", 80))
		fmt.Println("❓ RELATED QUESTIONS")
		fmt.Println(strings.Repeat("=", 80))
		for i, question := range relatedQuestions {
			fmt.Printf("%d. %s\n", i+1, question)
		}
	}

	// Example 3: Alternative approach using manual polling
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("📖 ALTERNATIVE: Manual Polling Example")
	fmt.Println(strings.Repeat("=", 80))

	// This example shows how you might implement your own polling logic
	exampleManualPolling(client, job.ID)
}

// exampleManualPolling demonstrates how to implement custom polling logic
func exampleManualPolling(client *perplexity.Client, jobID string) {
	fmt.Println("🔄 Demonstrating manual polling approach...")

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Check job status manually
	job, err := client.GetAsyncJobWithContext(ctx, jobID)
	if err != nil {
		fmt.Printf("Error getting job status: %v\n", err)
		return
	}

	fmt.Printf("Current job status: %s\n", job.Status)

	// If job is still processing, you could implement custom polling here
	if !job.IsCompleted() && !job.IsExpired() {
		fmt.Println("Job is still processing. In a real scenario, you would poll periodically.")
		fmt.Println("You can also list all your async jobs:")

		// List recent async jobs
		jobList, err := client.ListAsyncJobsWithContext(ctx, 5, 0)
		if err != nil {
			fmt.Printf("Error listing jobs: %v\n", err)
			return
		}

		fmt.Printf("Found %d recent jobs:\n", len(jobList.Jobs))
		for i, j := range jobList.Jobs {
			fmt.Printf("  %d. %s (%s) - %s\n", i+1, j.ID, j.Status, j.CreatedAt.Format("2006-01-02 15:04:05"))
		}
	}

	fmt.Println("✅ Manual polling example completed!")
}
