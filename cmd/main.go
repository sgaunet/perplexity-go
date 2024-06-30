package main

import (
	"fmt"
	"os"

	"github.com/sgaunet/perplexity-go"
)

func main() {
	client := perplexity.NewClient(os.Getenv("PPLX_API_KEY"))
	res, err := client.CreateCompletion([]perplexity.Message{
		{
			Role:    "user",
			Content: "What's the capital of France?",
		},
	})

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(res.GetLastContent())
}
