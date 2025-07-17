# Perplexity Go Client Examples

This directory contains example programs demonstrating various features of the Perplexity Go client library.

## Examples

### 1. Basic Completion (`basic-completion/`)
Demonstrates basic completion requests with web search options, including:
- Setting search context size for comprehensive answers
- Using user location for localized results
- Handling citations and search results

```bash
go run cmd/basic-completion/main.go
```

### 2. Structured Output (`structured-output/`)
Shows how to use structured output features (requires "sonar" model):
- JSON Schema format for structured data responses
- Regex format for pattern-matching responses

```bash
go run cmd/structured-output/main.go
```

### 3. Image Search (`image-search/`)
Demonstrates how to search for and retrieve images:
- Enabling image returns in responses
- Parsing image metadata (URL, description, dimensions)

```bash
go run cmd/image-search/main.go
```

### 4. Streaming with SSE (`streaming-sse/`)
Shows Server-Sent Events (SSE) streaming implementation:
- Real-time response streaming
- Proper goroutine and channel management
- Synchronization with sync.WaitGroup

```bash
go run cmd/streaming-sse/main.go
```

## Prerequisites

Before running any example, ensure you have:

1. Set the `PPLX_API_KEY` environment variable:
   ```bash
   export PPLX_API_KEY="your-api-key-here"
   ```

2. Installed the required dependencies:
   ```bash
   go mod download
   ```

## Shared Utilities

The `shared/` directory contains common utilities used across examples, such as formatting helpers.

## Notes

- All examples include comprehensive error handling and validation
- Each example is self-contained and can be run independently
- Comments in the code explain key concepts and implementation details