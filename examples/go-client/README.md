# Go Client Examples for AI Gateway

This directory contains Go examples for using the AI Gateway with Claude models through an OpenAI-compatible API.

## Features Demonstrated

1. ✅ **Simple Chat Completions** - Basic text generation
2. ✅ **System Messages** - Configuring model behavior
3. ✅ **Function/Tool Calling** - Using Claude with tools
4. ✅ **Multi-turn Conversations** - Context-aware dialogues

## Running the Examples

### Prerequisites

1. **Start the AI Gateway server:**
   ```bash
   # From the project root
   export AWS_REGION=us-east-1
   export PORT=8090
   ./server
   ```

2. **Configure AWS credentials** (if testing with real Bedrock):
   ```bash
   aws configure
   # or set environment variables:
   export AWS_ACCESS_KEY_ID=your-key
   export AWS_SECRET_ACCESS_KEY=your-secret
   ```

### Run the Examples

```bash
cd examples/go-client
go run main.go
```

## Example Output

```
╔══════════════════════════════════════════════════════════════╗
║                                                              ║
║          Go Client Examples - AI Gateway                     ║
║                                                              ║
╚══════════════════════════════════════════════════════════════╝

Example 1: Simple Chat Completion (Claude 3 Haiku)
─────────────────────────────────────────────────
Model: claude-3-haiku
Response: 2 + 2 equals 4.
Tokens: 15 input + 8 output = 23 total

Example 2: Chat with System Message (Claude 3 Sonnet)
─────────────────────────────────────────────────
Response: Here's a simple "Hello, World!" program in Go:

```go
package main

import "fmt"

func main() {
    fmt.Println("Hello, World!")
}
```

...
```

## Code Examples

### 1. Simple Chat Completion

```go
client := NewClient("http://localhost:8090", "")

req := &ChatCompletionRequest{
    Model: "claude-3-haiku",
    Messages: []ChatMessage{
        {Role: "user", Content: "What is 2+2?"},
    },
    MaxTokens:   100,
    Temperature: 0.7,
}

resp, err := client.ChatCompletion(req)
if err != nil {
    log.Fatal(err)
}

fmt.Println(resp.Choices[0].Message.Content)
```

### 2. System Message

```go
req := &ChatCompletionRequest{
    Model: "claude-3-sonnet",
    Messages: []ChatMessage{
        {Role: "system", Content: "You are a helpful coding assistant."},
        {Role: "user", Content: "Write a hello world in Go"},
    },
    MaxTokens: 300,
}
```

### 3. Function/Tool Calling

```go
req := &ChatCompletionRequest{
    Model: "claude-3-sonnet",
    Messages: []ChatMessage{
        {Role: "user", Content: "What's the weather in SF?"},
    },
    Tools: []Tool{
        {
            Type: "function",
            Function: Function{
                Name:        "get_weather",
                Description: "Get the current weather for a location",
                Parameters: map[string]interface{}{
                    "type": "object",
                    "properties": map[string]interface{}{
                        "location": map[string]interface{}{
                            "type":        "string",
                            "description": "The city and state",
                        },
                    },
                    "required": []string{"location"},
                },
            },
        },
    },
    ToolChoice: "auto",
}

resp, err := client.ChatCompletion(req)

// Check if model wants to call a function
if len(resp.Choices[0].Message.ToolCalls) > 0 {
    for _, toolCall := range resp.Choices[0].Message.ToolCalls {
        fmt.Printf("Function: %s\n", toolCall.Function.Name)
        fmt.Printf("Arguments: %s\n", toolCall.Function.Arguments)

        // Execute the function and send result back
        // ...
    }
}
```

### 4. Multi-turn Conversation

```go
req := &ChatCompletionRequest{
    Model: "claude-3-haiku",
    Messages: []ChatMessage{
        {Role: "user", Content: "My name is Alice"},
        {Role: "assistant", Content: "Nice to meet you, Alice!"},
        {Role: "user", Content: "What is my name?"},
    },
}

resp, err := client.ChatCompletion(req)
// Response: "Your name is Alice."
```

## Available Models

| Model ID | Description | Speed | Cost |
|----------|-------------|-------|------|
| `claude-3-haiku` | Fast, cost-effective | ⚡⚡⚡ | $ |
| `claude-3-sonnet` | Balanced performance | ⚡⚡ | $$ |
| `claude-3-opus` | Most capable | ⚡ | $$$ |
| `claude-3-5-sonnet` | Latest, enhanced | ⚡⚡ | $$ |

## Client Configuration

### With Authentication

If the gateway has authentication enabled:

```go
client := NewClient("http://localhost:8090", "your-api-key-here")
```

### Custom Timeout

```go
client := &AIGatewayClient{
    BaseURL: "http://localhost:8090",
    APIKey:  "",
    HTTPClient: &http.Client{
        Timeout: 300 * time.Second, // 5 minutes
    },
}
```

## Error Handling

```go
resp, err := client.ChatCompletion(req)
if err != nil {
    // Handle different error types
    if strings.Contains(err.Error(), "API error") {
        log.Printf("API returned error: %v", err)
    } else if strings.Contains(err.Error(), "timeout") {
        log.Printf("Request timed out: %v", err)
    } else {
        log.Printf("Unknown error: %v", err)
    }
    return
}
```

## Best Practices

1. **Reuse HTTP Client**: Create one `AIGatewayClient` and reuse it
2. **Set Reasonable Timeouts**: Default is 120s, adjust based on your needs
3. **Handle Tool Calls**: When using functions, implement the full loop
4. **Check Token Usage**: Monitor `resp.Usage` to track costs
5. **Use Appropriate Models**:
   - Haiku for simple tasks
   - Sonnet for most use cases
   - Opus for complex reasoning

## Integration with Your App

```go
package main

import (
    "log"
    "yourapp/aigw" // Your wrapper around the client
)

type AIService struct {
    client *AIGatewayClient
}

func NewAIService(baseURL, apiKey string) *AIService {
    return &AIService{
        client: NewClient(baseURL, apiKey),
    }
}

func (s *AIService) GenerateResponse(userMessage string) (string, error) {
    req := &ChatCompletionRequest{
        Model: "claude-3-sonnet",
        Messages: []ChatMessage{
            {Role: "user", Content: userMessage},
        },
        MaxTokens: 1000,
    }

    resp, err := s.client.ChatCompletion(req)
    if err != nil {
        return "", err
    }

    return resp.Choices[0].Message.Content, nil
}

func main() {
    service := NewAIService("http://localhost:8090", "")

    response, err := service.GenerateResponse("Hello!")
    if err != nil {
        log.Fatal(err)
    }

    log.Println(response)
}
```

## Next Steps

1. Implement streaming support (coming soon)
2. Add response caching
3. Implement retry logic with exponential backoff
4. Add structured logging
5. Create helper methods for common patterns

## Troubleshooting

### "Connection refused"
- Ensure the server is running on port 8090
- Check: `curl http://localhost:8090/health`

### "API error (status 401)"
- Add API key if authentication is enabled
- Check: `export AUTH_ENABLED` on server

### "API error (status 400): Model not found"
- Verify model ID is correct
- Check available models: `curl http://localhost:8090/v1/models`

### AWS Credential Errors
- Server needs valid AWS credentials to invoke Bedrock
- Configure on the server side, not in the client

## License

Apache 2.0 - See LICENSE file
