package main

import (
	"fmt"
	"os"
	"sync"

	"github.com/sgaunet/perplexity-go/v2"
)

// This example demonstrates how to create a completion request with a message
// It then sends the request to the API and prints the last completion content.
func main() {
	client := perplexity.NewClient(os.Getenv("PPLX_API_KEY"))
	msg := []perplexity.Message{
		{
			Role:    "user",
			Content: "Wat's the capital of France?",
		},
	}
	req := perplexity.NewCompletionRequest(perplexity.WithMessages(msg))
	err := req.Validate()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	res, err := client.SendCompletionRequest(req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(res.GetLastContent())
	for i, c := range res.GetCitations() {
		fmt.Printf("Citation %d: %s", i+1, c)
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
