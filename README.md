# Perplexity API Go Client

[![Go Report Card](https://goreportcard.com/badge/github.com/sgaunet/perplexity-go)](https://goreportcard.com/report/github.com/sgaunet/perplexity-go)
![coverage](https://raw.githubusercontent.com/wiki/sgaunet/perplexity-go/coverage-badge.svg)
[![GoDoc](https://godoc.org/github.com/sgaunet/perplexity-go/v2?status.svg)](https://godoc.org/github.com/sgaunet/perplexity-go/v2)

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

## Max Tokens

* **General Use Cases**: For most general-purpose applications, setting max_tokens to 4000 is a good starting point. This is because many Perplexity models, like the default sonar model, can generate responses up to this limit (It's the default value in this library).

* **Long-Form Content**: If you are working with long-form content or need more extensive responses, you might consider models like sonar-pro, which can handle larger outputs. However, the maximum output tokens for such models might still be capped at 8,000 tokens.

* **Model-Specific Limits**: Ensure that your chosen model supports the max_tokens value you set. For example, some models might have a maximum context window or output limit that you should not exceed.

* **Performance Considerations**: Higher max_tokens values can increase response time and computational resources. Therefore, balance between the need for detailed responses and performance efficiency.

## License

This project is licensed under the MIT License - see the LICENSE file for details.
