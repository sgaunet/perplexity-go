# Perplexity API Go Client

![coverage](https://raw.githubusercontent.com/wiki/sgaunet/perplexity-go/coverage-badge.svg)
[![GoDoc](https://godoc.org/github.com/sgaunet/perplexity-go/v2?status.svg)](https://godoc.org/github.com/sgaunet/perplexity-go/v2)
[![CI](https://github.com/sgaunet/perplexity-go/actions/workflows/coverage.yml/badge.svg)](https://github.com/sgaunet/perplexity-go/actions/workflows/coverage.yml)
[![Release](https://github.com/sgaunet/perplexity-go/actions/workflows/release.yml/badge.svg)](https://github.com/sgaunet/perplexity-go/actions/workflows/release.yml)
[![golangci-lint](https://github.com/sgaunet/perplexity-go/actions/workflows/linter.yml/badge.svg)](https://github.com/sgaunet/perplexity-go/actions/workflows/linter.yml)

A lightweight Go library for interacting with the [Perplexity AI API](https://docs.perplexity.ai/reference/post_chat_completions), supporting both chat completions and the Search API.

## Features

* Simple and easy-to-use interface for chat completion requests
* **Search API support** for direct access to Perplexity's real-time web index
* Supports all Perplexity models, including online LLMs
* Handles authentication and API key management
* Comprehensive request validation
* Concurrent request support with thread-safe operations
* Context-aware request handling with cancellation support

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

## Search API

The Perplexity Search API provides direct access to Perplexity's real-time web index without the generative LLM layer, returning raw ranked search results with structured snippets.

### Features

* Direct web search without AI generation layer
* Single query or multi-query search support
* Structured results with titles, URLs, snippets, and scores
* Optional image results
* Domain filtering
* Country-specific results

### Pricing

The Search API is priced at **$5 per 1,000 requests** (as of 2024), separate from chat completion pricing.

### Basic Search Usage

```go
package main

import (
    "fmt"
    "os"

    "github.com/sgaunet/perplexity-go/v2"
)

func main() {
    client := perplexity.NewClient(os.Getenv("PPLX_API_KEY"))

    // Simple search
    req := perplexity.NewSearchRequest("golang best practices")

    resp, err := client.SendSearchRequest(req)
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        os.Exit(1)
    }

    // Print results
    for i, result := range resp.Results {
        fmt.Printf("%d. %s\n", i+1, result.Title)
        fmt.Printf("   URL: %s\n", result.URL)
        if result.Snippet != nil {
            fmt.Printf("   Snippet: %s\n", *result.Snippet)
        }
    }
}
```

### Advanced Search Options

```go
// Search with all options
req := perplexity.NewSearchRequest(
    "machine learning papers",
    perplexity.WithSearchMaxResults(10),
    perplexity.WithSearchReturnImages(true),
    perplexity.WithSearchReturnSnippets(true),
    perplexity.WithSearchCountry("US"),
    perplexity.WithSearchDomains([]string{"arxiv.org", "github.com"}),
)

resp, err := client.SendSearchRequest(req)
```

### Multi-Query Search

```go
// Search multiple queries at once
queries := []string{
    "Go programming language",
    "Rust programming language",
    "Python programming language",
}

req := perplexity.NewSearchRequest(queries)
resp, err := client.SendSearchRequest(req)
```

### Request Validation

```go
// Validate search request before sending
validator := perplexity.NewSearchRequestValidator()
if err := validator.ValidateSearchRequest(req); err != nil {
    fmt.Printf("Validation error: %v\n", err)
    return
}
```

### Context Support

```go
import "context"

// With timeout
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

resp, err := client.SendSearchRequestWithContext(ctx, req)
```

### Search vs Chat Completions

| Feature | Search API | Chat Completions |
|---------|------------|------------------|
| **Purpose** | Raw web search results | AI-generated responses |
| **Response** | Ranked list of URLs with snippets | Natural language text |
| **Use Case** | Research, data gathering | Q&A, summarization, analysis |
| **Pricing** | $5 per 1K requests | Variable by model |
| **Latency** | Lower (no generation) | Higher (includes generation) |

## Documentation

For detailed documentation and more examples, please refer to the [GoDoc page](https://godoc.org/github.com/sgaunet/perplexity-go/v2).

## Max Tokens

* **General Use Cases**: For most general-purpose applications, setting max_tokens to 4000 is a good starting point. This is because many Perplexity models, like the default sonar model, can generate responses up to this limit (It's the default value in this library).

* **Long-Form Content**: If you are working with long-form content or need more extensive responses, you might consider models like sonar-pro, which can handle larger outputs. However, the maximum output tokens for such models might still be capped at 8,000 tokens.

* **Model-Specific Limits**: Ensure that your chosen model supports the max_tokens value you set. For example, some models might have a maximum context window or output limit that you should not exceed.

* **Performance Considerations**: Higher max_tokens values can increase response time and computational resources. Therefore, balance between the need for detailed responses and performance efficiency.

## License

This project is licensed under the MIT License - see the LICENSE file for details.
