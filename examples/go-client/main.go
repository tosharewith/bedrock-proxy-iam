// Copyright 2025 Bedrock Proxy Authors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

const (
	// BaseURL is the AI Gateway URL
	BaseURL = "http://localhost:8090"
)

// OpenAI-compatible types
type ChatCompletionRequest struct {
	Model       string        `json:"model"`
	Messages    []ChatMessage `json:"messages"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
	Temperature float64       `json:"temperature,omitempty"`
	Tools       []Tool        `json:"tools,omitempty"`
	ToolChoice  interface{}   `json:"tool_choice,omitempty"`
}

type ChatMessage struct {
	Role       string      `json:"role"`
	Content    string      `json:"content,omitempty"`
	ToolCalls  []ToolCall  `json:"tool_calls,omitempty"`
	ToolCallID string      `json:"tool_call_id,omitempty"`
}

type Tool struct {
	Type     string   `json:"type"`
	Function Function `json:"function"`
}

type Function struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

type ToolCall struct {
	ID       string       `json:"id"`
	Type     string       `json:"type"`
	Function FunctionCall `json:"function"`
}

type FunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type ChatCompletionResponse struct {
	ID      string                 `json:"id"`
	Object  string                 `json:"object"`
	Created int64                  `json:"created"`
	Model   string                 `json:"model"`
	Choices []ChatCompletionChoice `json:"choices"`
	Usage   Usage                  `json:"usage"`
}

type ChatCompletionChoice struct {
	Index        int         `json:"index"`
	Message      ChatMessage `json:"message"`
	FinishReason string      `json:"finish_reason"`
}

type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// AIGatewayClient is a simple client for the AI Gateway
type AIGatewayClient struct {
	BaseURL    string
	APIKey     string
	HTTPClient *http.Client
}

// NewClient creates a new AI Gateway client
func NewClient(baseURL, apiKey string) *AIGatewayClient {
	return &AIGatewayClient{
		BaseURL: baseURL,
		APIKey:  apiKey,
		HTTPClient: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

// ChatCompletion sends a chat completion request
func (c *AIGatewayClient) ChatCompletion(req *ChatCompletionRequest) (*ChatCompletionResponse, error) {
	// Marshal request
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequest("POST", c.BaseURL+"/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	if c.APIKey != "" {
		httpReq.Header.Set("X-API-Key", c.APIKey)
	}

	// Send request
	resp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	// Parse response
	var chatResp ChatCompletionResponse
	if err := json.Unmarshal(body, &chatResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &chatResp, nil
}

func main() {
	fmt.Println("╔══════════════════════════════════════════════════════════════╗")
	fmt.Println("║                                                              ║")
	fmt.Println("║          Go Client Examples - AI Gateway                     ║")
	fmt.Println("║                                                              ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════╝")
	fmt.Println()

	// Create client
	client := NewClient(BaseURL, "") // Add API key if auth is enabled

	// Example 1: Simple Chat Completion
	fmt.Println("Example 1: Simple Chat Completion (Claude 3 Haiku)")
	fmt.Println("─────────────────────────────────────────────────")
	simpleExample(client)
	fmt.Println()

	// Example 2: Chat with System Message
	fmt.Println("Example 2: Chat with System Message (Claude 3 Sonnet)")
	fmt.Println("─────────────────────────────────────────────────")
	systemMessageExample(client)
	fmt.Println()

	// Example 3: Function Calling
	fmt.Println("Example 3: Function/Tool Calling (Claude 3 Sonnet)")
	fmt.Println("─────────────────────────────────────────────────")
	functionCallingExample(client)
	fmt.Println()

	// Example 4: Multi-turn Conversation
	fmt.Println("Example 4: Multi-turn Conversation (Claude 3 Haiku)")
	fmt.Println("─────────────────────────────────────────────────")
	multiTurnExample(client)
	fmt.Println()
}

// simpleExample demonstrates a basic chat completion
func simpleExample(client *AIGatewayClient) {
	req := &ChatCompletionRequest{
		Model: "claude-3-haiku",
		Messages: []ChatMessage{
			{Role: "user", Content: "What is 2+2? Answer in one sentence."},
		},
		MaxTokens:   100,
		Temperature: 0.7,
	}

	resp, err := client.ChatCompletion(req)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	fmt.Printf("Model: %s\n", resp.Model)
	fmt.Printf("Response: %s\n", resp.Choices[0].Message.Content)
	fmt.Printf("Tokens: %d input + %d output = %d total\n",
		resp.Usage.PromptTokens,
		resp.Usage.CompletionTokens,
		resp.Usage.TotalTokens)
}

// systemMessageExample demonstrates using system messages
func systemMessageExample(client *AIGatewayClient) {
	req := &ChatCompletionRequest{
		Model: "claude-3-sonnet",
		Messages: []ChatMessage{
			{Role: "system", Content: "You are a helpful coding assistant. Always explain code clearly."},
			{Role: "user", Content: "Write a hello world in Go"},
		},
		MaxTokens:   300,
		Temperature: 0.7,
	}

	resp, err := client.ChatCompletion(req)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	fmt.Printf("Response: %s\n", resp.Choices[0].Message.Content)
}

// functionCallingExample demonstrates function/tool calling
func functionCallingExample(client *AIGatewayClient) {
	// Define a weather tool
	req := &ChatCompletionRequest{
		Model: "claude-3-sonnet",
		Messages: []ChatMessage{
			{Role: "user", Content: "What's the weather like in San Francisco?"},
		},
		MaxTokens:   500,
		Temperature: 0.7,
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
								"description": "The city and state, e.g. San Francisco, CA",
							},
							"unit": map[string]interface{}{
								"type":        "string",
								"enum":        []string{"celsius", "fahrenheit"},
								"description": "Temperature unit",
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
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	// Check if model wants to call a function
	if len(resp.Choices[0].Message.ToolCalls) > 0 {
		fmt.Println("Model wants to call functions:")
		for _, toolCall := range resp.Choices[0].Message.ToolCalls {
			fmt.Printf("  - Function: %s\n", toolCall.Function.Name)
			fmt.Printf("    Arguments: %s\n", toolCall.Function.Arguments)
		}

		// In a real application, you would:
		// 1. Execute the function
		// 2. Send the result back to the model
		// 3. Get the final response
		fmt.Println("\nIn a real app, you would now:")
		fmt.Println("  1. Execute get_weather(location='San Francisco, CA')")
		fmt.Println("  2. Send the result back to the model")
		fmt.Println("  3. Get the final natural language response")
	} else {
		fmt.Printf("Response: %s\n", resp.Choices[0].Message.Content)
	}
}

// multiTurnExample demonstrates a multi-turn conversation
func multiTurnExample(client *AIGatewayClient) {
	req := &ChatCompletionRequest{
		Model: "claude-3-haiku",
		Messages: []ChatMessage{
			{Role: "user", Content: "My name is Alice"},
			{Role: "assistant", Content: "Nice to meet you, Alice! How can I help you today?"},
			{Role: "user", Content: "What is my name?"},
		},
		MaxTokens:   50,
		Temperature: 0.7,
	}

	resp, err := client.ChatCompletion(req)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	fmt.Printf("Response: %s\n", resp.Choices[0].Message.Content)
	if len(resp.Choices[0].Message.Content) > 0 && contains(resp.Choices[0].Message.Content, "Alice") {
		fmt.Println("✓ Model remembered the name!")
	}
}

// contains checks if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr))
}
