// Copyright 2025 Bedrock Proxy Authors
// SPDX-License-Identifier: Apache-2.0

package router

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Config represents the router configuration loaded from YAML
type Config struct {
	ModelMappings map[string]ModelMapping `yaml:"model_mappings"`
	Routing       RoutingConfig           `yaml:"routing"`
	Providers     map[string]ProviderConfig `yaml:"providers"`
	Features      FeatureFlags            `yaml:"features"`
}

// ModelMapping defines how a model name maps to different providers
type ModelMapping struct {
	DefaultProvider string                       `yaml:"default_provider"`
	Providers       map[string]ProviderModelInfo `yaml:"providers"`
}

// ProviderModelInfo contains provider-specific model information
type ProviderModelInfo struct {
	Model      string            `yaml:"model"`
	Region     string            `yaml:"region,omitempty"`
	Location   string            `yaml:"location,omitempty"`
	Deployment string            `yaml:"deployment,omitempty"`
	APIVersion string            `yaml:"api_version,omitempty"`
	Metadata   map[string]string `yaml:"metadata,omitempty"`
}

// RoutingConfig defines routing rules and fallback behavior
type RoutingConfig struct {
	Patterns       []RoutingPattern `yaml:"patterns"`
	Fallback       FallbackConfig   `yaml:"fallback"`
	LoadBalancing  LoadBalancingConfig `yaml:"load_balancing"`
}

// RoutingPattern defines a regex pattern for routing
type RoutingPattern struct {
	Pattern         string `yaml:"pattern"`
	DefaultProvider string `yaml:"default_provider"`
	Description     string `yaml:"description"`
	compiledPattern *regexp.Regexp
}

// FallbackConfig defines fallback behavior
type FallbackConfig struct {
	Enabled     bool     `yaml:"enabled"`
	Providers   []string `yaml:"providers"`
	MaxAttempts int      `yaml:"max_attempts"`
}

// LoadBalancingConfig defines load balancing strategy
type LoadBalancingConfig struct {
	Enabled  bool   `yaml:"enabled"`
	Strategy string `yaml:"strategy"` // round_robin, least_latency, random, cost_optimized
}

// ProviderConfig contains provider-specific configuration
type ProviderConfig struct {
	Enabled     bool          `yaml:"enabled"`
	Region      string        `yaml:"region,omitempty"`
	Location    string        `yaml:"location,omitempty"`
	ProjectID   string        `yaml:"project_id,omitempty"`
	Endpoint    string        `yaml:"endpoint,omitempty"`
	BaseURL     string        `yaml:"base_url,omitempty"`
	APIVersion  string        `yaml:"api_version,omitempty"`
	Timeout     time.Duration `yaml:"timeout"`
	MaxRetries  int           `yaml:"max_retries"`
	RetryDelay  time.Duration `yaml:"retry_delay,omitempty"`
}

// FeatureFlags contains feature flag settings
type FeatureFlags struct {
	OpenAICompatibility bool `yaml:"openai_compatibility"`
	Streaming           bool `yaml:"streaming"`
	CostTracking        bool `yaml:"cost_tracking"`
	AutoFallback        bool `yaml:"auto_fallback"`
	ResponseCaching     bool `yaml:"response_caching"`
}

// LoadConfig loads the router configuration from a YAML file
func LoadConfig(path string) (*Config, error) {
	// Read file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Expand environment variables
	expanded := os.ExpandEnv(string(data))

	// Parse YAML
	var config Config
	if err := yaml.Unmarshal([]byte(expanded), &config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Compile regex patterns
	for i := range config.Routing.Patterns {
		pattern := &config.Routing.Patterns[i]
		compiled, err := regexp.Compile(pattern.Pattern)
		if err != nil {
			return nil, fmt.Errorf("failed to compile pattern %q: %w", pattern.Pattern, err)
		}
		pattern.compiledPattern = compiled
	}

	// Set defaults
	if config.Routing.Fallback.MaxAttempts == 0 {
		config.Routing.Fallback.MaxAttempts = 2
	}

	return &config, nil
}

// GetModelMapping returns the mapping for a given model name
func (c *Config) GetModelMapping(modelName string) (*ModelMapping, bool) {
	mapping, exists := c.ModelMappings[modelName]
	return &mapping, exists
}

// GetDefaultProvider returns the default provider for a model
// First checks exact model mapping, then pattern matching, then returns empty string
func (c *Config) GetDefaultProvider(modelName string) string {
	// Check exact mapping
	if mapping, exists := c.ModelMappings[modelName]; exists {
		return mapping.DefaultProvider
	}

	// Check pattern matching
	for _, pattern := range c.Routing.Patterns {
		if pattern.compiledPattern.MatchString(modelName) {
			return pattern.DefaultProvider
		}
	}

	return ""
}

// GetProviderModelInfo returns provider-specific model info
func (c *Config) GetProviderModelInfo(modelName, providerName string) (*ProviderModelInfo, error) {
	mapping, exists := c.ModelMappings[modelName]
	if !exists {
		return nil, fmt.Errorf("model %q not found in mappings", modelName)
	}

	providerInfo, exists := mapping.Providers[providerName]
	if !exists {
		return nil, fmt.Errorf("provider %q not found for model %q", providerName, modelName)
	}

	return &providerInfo, nil
}

// GetProviderConfig returns configuration for a provider
func (c *Config) GetProviderConfig(providerName string) (*ProviderConfig, bool) {
	config, exists := c.Providers[providerName]
	return &config, exists
}

// IsProviderEnabled checks if a provider is enabled
func (c *Config) IsProviderEnabled(providerName string) bool {
	config, exists := c.Providers[providerName]
	if !exists {
		return false
	}
	return config.Enabled
}

// GetFallbackProviders returns the list of fallback providers
func (c *Config) GetFallbackProviders() []string {
	if !c.Routing.Fallback.Enabled {
		return nil
	}
	return c.Routing.Fallback.Providers
}

// ListEnabledProviders returns all enabled providers
func (c *Config) ListEnabledProviders() []string {
	var enabled []string
	for name, config := range c.Providers {
		if config.Enabled {
			enabled = append(enabled, name)
		}
	}
	return enabled
}

// ListModelsForProvider returns all models that can use a specific provider
func (c *Config) ListModelsForProvider(providerName string) []string {
	var models []string
	for modelName, mapping := range c.ModelMappings {
		if _, exists := mapping.Providers[providerName]; exists {
			models = append(models, modelName)
		}
	}
	return models
}

// ValidateConfig performs validation on the loaded configuration
func (c *Config) ValidateConfig() error {
	var errors []string

	// Check that default providers exist and are enabled
	for modelName, mapping := range c.ModelMappings {
		if mapping.DefaultProvider == "" {
			errors = append(errors, fmt.Sprintf("model %q has no default provider", modelName))
			continue
		}

		providerConfig, exists := c.Providers[mapping.DefaultProvider]
		if !exists {
			errors = append(errors, fmt.Sprintf("model %q default provider %q not found in provider configs",
				modelName, mapping.DefaultProvider))
			continue
		}

		if !providerConfig.Enabled {
			errors = append(errors, fmt.Sprintf("model %q default provider %q is disabled",
				modelName, mapping.DefaultProvider))
		}
	}

	// Check fallback providers exist
	if c.Routing.Fallback.Enabled {
		for _, providerName := range c.Routing.Fallback.Providers {
			if _, exists := c.Providers[providerName]; !exists {
				errors = append(errors, fmt.Sprintf("fallback provider %q not found in provider configs", providerName))
			}
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("configuration validation failed:\n  - %s", strings.Join(errors, "\n  - "))
	}

	return nil
}
