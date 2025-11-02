// Copyright 2025 Bedrock Proxy Authors
// SPDX-License-Identifier: Apache-2.0

package providers

import (
	"context"
	"io"
	"time"
)

// Provider defines the interface that all AI providers must implement
type Provider interface {
	// Name returns the provider identifier (bedrock, azure, openai, anthropic, vertex)
	Name() string

	// HealthCheck verifies the provider is accessible
	HealthCheck(ctx context.Context) error

	// Invoke sends a request to the provider's native API
	// Returns response, error
	Invoke(ctx context.Context, request *ProviderRequest) (*ProviderResponse, error)

	// InvokeStreaming handles streaming responses (for chat completions)
	InvokeStreaming(ctx context.Context, request *ProviderRequest) (io.ReadCloser, error)

	// ListModels returns available models for this provider
	ListModels(ctx context.Context) ([]Model, error)

	// GetModelInfo returns details about a specific model
	GetModelInfo(ctx context.Context, modelID string) (*Model, error)
}

// ProviderRequest wraps the provider-specific request
type ProviderRequest struct {
	// HTTP method (POST, GET, etc.)
	Method string

	// API endpoint path (relative to provider base URL)
	Path string

	// HTTP headers
	Headers map[string]string

	// Request body (usually JSON bytes)
	Body []byte

	// URL query parameters
	QueryParams map[string]string

	// Additional metadata (user info, tracing, etc.)
	Metadata map[string]any

	// Original request context
	Context context.Context
}

// ProviderResponse wraps the provider's response
type ProviderResponse struct {
	// HTTP status code
	StatusCode int

	// Response headers
	Headers map[string]string

	// Response body (usually JSON bytes)
	Body []byte

	// Additional metadata (latency, tokens, cost, etc.)
	Metadata ResponseMetadata
}

// ResponseMetadata contains additional information about the response
type ResponseMetadata struct {
	// Latency of the request
	Latency time.Duration

	// Token usage
	InputTokens  int
	OutputTokens int
	TotalTokens  int

	// Cost in USD
	InputCost  float64
	OutputCost float64
	TotalCost  float64

	// Model used (may differ from requested model)
	ModelUsed string

	// Provider-specific metadata
	ProviderMetadata map[string]any
}

// Model represents an AI model
type Model struct {
	// Model identifier (e.g., "gpt-4", "claude-3-sonnet")
	ID string

	// Provider name
	Provider string

	// Human-readable name
	Name string

	// Model description
	Description string

	// Capabilities (chat, completion, embeddings, streaming, vision, function_calling)
	Capabilities []string

	// Max context window length in tokens
	ContextWindow int

	// Input price per 1M tokens (USD)
	InputPrice float64

	// Output price per 1M tokens (USD)
	OutputPrice float64

	// Whether the model is currently available
	Available bool

	// Additional provider-specific metadata
	Metadata map[string]any
}

// StreamEvent represents a single event in a streaming response
type StreamEvent struct {
	// Event type (e.g., "message_start", "content_block_delta", "message_stop")
	Type string

	// Event data (provider-specific format)
	Data []byte

	// Error if the event represents an error
	Error error
}

// HasCapability checks if a model has a specific capability
func (m *Model) HasCapability(capability string) bool {
	for _, cap := range m.Capabilities {
		if cap == capability {
			return true
		}
	}
	return false
}

// CalculateCost calculates the total cost based on token usage
func (m *Model) CalculateCost(inputTokens, outputTokens int) float64 {
	inputCost := (float64(inputTokens) / 1_000_000) * m.InputPrice
	outputCost := (float64(outputTokens) / 1_000_000) * m.OutputPrice
	return inputCost + outputCost
}

// ProviderError represents a provider-specific error
type ProviderError struct {
	// Provider name
	Provider string

	// HTTP status code (if applicable)
	StatusCode int

	// Error code (provider-specific)
	Code string

	// Human-readable error message
	Message string

	// Original error
	Err error
}

func (e *ProviderError) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Message
}

func (e *ProviderError) Unwrap() error {
	return e.Err
}

// Common capability constants
const (
	CapabilityChat            = "chat"
	CapabilityCompletion      = "completion"
	CapabilityEmbeddings      = "embeddings"
	CapabilityStreaming       = "streaming"
	CapabilityVision          = "vision"
	CapabilityFunctionCalling = "function_calling"
	CapabilityJSON            = "json_mode"
)

// Common error codes
const (
	ErrCodeInvalidRequest     = "invalid_request"
	ErrCodeAuthenticationFail = "authentication_failed"
	ErrCodeRateLimitExceeded  = "rate_limit_exceeded"
	ErrCodeModelNotFound      = "model_not_found"
	ErrCodeServiceUnavailable = "service_unavailable"
	ErrCodeInternalError      = "internal_error"
)
