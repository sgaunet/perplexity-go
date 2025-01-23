package main

import (
	"fmt"
	"os"

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
	req := perplexity.NewCompletionRequest(perplexity.WithMessages(msg), perplexity.WithReturnImages(true))
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
	// fmt.Printf("%+v\n", *req)
	// fmt.Println("*************")
	// fmt.Printf("%+v\n", res)
}
