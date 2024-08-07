[![Go Report Card](https://goreportcard.com/badge/github.com/sgaunet/perplexity-go)](https://goreportcard.com/report/github.com/sgaunet/perplexity-go)
[![Maintainability](https://api.codeclimate.com/v1/badges/f01b49c0008ff9ad59cb/maintainability)](https://codeclimate.com/github/sgaunet/perplexity-go/maintainability)
[![Test Coverage](https://api.codeclimate.com/v1/badges/f01b49c0008ff9ad59cb/test_coverage)](https://codeclimate.com/github/sgaunet/perplexity-go/test_coverage)

# Perplexity API Go Client

A lightweight Go library for interacting with the [Perplexity AI API](https://docs.perplexity.ai/reference/post_chat_completions), focusing on the chat completion endpoint.

Features

    Simple and easy-to-use interface for making chat completion requests
    Supports all Perplexity models, including online LLMs
    Handles authentication and API key management
    Provides convenient methods for common operations

## Installation

To install the library, use go get:

```bash
go get github.com/sgaunet/perplexity-go
```

## Usage

Here's a quick example of how to use the library:

```go
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
```

## Documentation

For detailed documentation and more examples, please refer to the GoDoc page.

## License

This project is licensed under the MIT License - see the LICENSE file for details.
