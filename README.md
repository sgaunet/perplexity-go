# Perplexity API Go Client

[![Go Report Card](https://goreportcard.com/badge/github.com/sgaunet/perplexity-go)](https://goreportcard.com/report/github.com/sgaunet/perplexity-go)
![coverage](https://raw.githubusercontent.com/wiki/sgaunet/perplexity-go/coverage-badge.svg)

A lightweight Go library for interacting with the [Perplexity AI API](https://docs.perplexity.ai/reference/post_chat_completions), focusing on the chat completion endpoint.

Features

    Simple and easy-to-use interface for making chat completion requests
    Supports all Perplexity models, including online LLMs
    Handles authentication and API key management
    Provides convenient methods for common operations

If you need a **CLI tool** to interact with the API, check out the [pplx](https://github.com/sgaunet/pplx) project.

Due to AI models that change regulary, only the default model will be handled for version >=2.5.0. Using the `WithModel`, you're able to specify the model you want to use. The default model will always be maintained up to date.
Now the library should be stable.

**If you have access to the beta version of the API** I'm interesred to get some informations to hanle image generation. Please contact me.

## Installation

To install the library, use go get:

```sh
go get github.com/sgaunet/perplexity-go/v2
```

## Usage

Here's a quick example of how to use the library:

```go
package main

import (
  "fmt"
  "os"

  "github.com/sgaunet/perplexity-go/v2"
)

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
}
```

## Documentation

For detailed documentation and more examples, please refer to the GoDoc page.

## License

This project is licensed under the MIT License - see the LICENSE file for details.
