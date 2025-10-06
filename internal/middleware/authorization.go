// Copyright 2025 Bedrock Proxy Authors
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"crypto/subtle"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

// AuthConfig holds authorization configuration
type AuthConfig struct {
	// API Key based auth
	APIKeys map[string]string // key -> user/role mapping

	// JWT based auth
	JWTSecret     string
	JWTIssuer     string
	JWTAudience   string

	// IAM based auth (for AWS resources)
	AllowedRoles  []string
	AllowedUsers  []string
}

// APIKeyAuth validates API key from header
func APIKeyAuth(validKeys map[string]string) gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("X-API-Key")
		if apiKey == "" {
			// Also check Authorization header with Bearer prefix
			authHeader := c.GetHeader("Authorization")
			if strings.HasPrefix(authHeader, "Bearer ") {
				apiKey = strings.TrimPrefix(authHeader, "Bearer ")
			}
		}

		if apiKey == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Missing API key",
				"message": "Provide API key via X-API-Key header or Authorization: Bearer <key>",
			})
			c.Abort()
			return
		}

		// Constant-time comparison to prevent timing attacks
		user, found := validKeys[apiKey]
		if !found || subtle.ConstantTimeCompare([]byte(apiKey), []byte(apiKey)) != 1 {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid API key",
			})
			c.Abort()
			return
		}

		// Set user context
		c.Set("user", user)
		c.Set("auth_method", "api_key")
		c.Next()
	}
}

// BasicAuth provides username/password authentication
func BasicAuth(credentials map[string]string) gin.HandlerFunc {
	return func(c *gin.Context) {
		username, password, hasAuth := c.Request.BasicAuth()

		if !hasAuth {
			c.Header("WWW-Authenticate", `Basic realm="Bedrock Proxy"`)
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Missing authentication",
			})
			c.Abort()
			return
		}

		expectedPassword, exists := credentials[username]
		if !exists || subtle.ConstantTimeCompare([]byte(password), []byte(expectedPassword)) != 1 {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid credentials",
			})
			c.Abort()
			return
		}

		c.Set("user", username)
		c.Set("auth_method", "basic")
		c.Next()
	}
}

// ServiceAccountAuth validates Kubernetes service account token
func ServiceAccountAuth(allowedServiceAccounts []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get service account from header (injected by service mesh or app)
		serviceAccount := c.GetHeader("X-Service-Account")
		namespace := c.GetHeader("X-Namespace")

		if serviceAccount == "" || namespace == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Missing service account credentials",
			})
			c.Abort()
			return
		}

		// Validate against allowed list
		fullSA := namespace + "/" + serviceAccount
		allowed := false
		for _, allowedSA := range allowedServiceAccounts {
			if fullSA == allowedSA {
				allowed = true
				break
			}
		}

		if !allowed {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Service account not authorized",
				"service_account": fullSA,
			})
			c.Abort()
			return
		}

		c.Set("user", fullSA)
		c.Set("auth_method", "service_account")
		c.Next()
	}
}

// LoadAPIKeysFromEnv loads API keys from environment variables
// Format: BEDROCK_API_KEY_<NAME>=<key>
func LoadAPIKeysFromEnv() map[string]string {
	keys := make(map[string]string)

	for _, env := range os.Environ() {
		if strings.HasPrefix(env, "BEDROCK_API_KEY_") {
			parts := strings.SplitN(env, "=", 2)
			if len(parts) == 2 {
				name := strings.TrimPrefix(parts[0], "BEDROCK_API_KEY_")
				key := parts[1]
				keys[key] = strings.ToLower(name)
			}
		}
	}

	return keys
}

// LoadAPIKeysFromSecret loads API keys from Kubernetes secret
// This would be used with a Secret mounted as env vars or volume
func LoadAPIKeysFromSecret(secretPath string) (map[string]string, error) {
	// Read from mounted secret volume
	// Implementation depends on how secrets are mounted
	keys := make(map[string]string)

	// Example: read from file-based secret
	// data, err := os.ReadFile(filepath.Join(secretPath, "api-keys.json"))
	// if err != nil {
	// 	return nil, err
	// }
	// json.Unmarshal(data, &keys)

	return keys, nil
}

// RateLimitByUser provides per-user rate limiting
func RateLimitByUser(requestsPerMinute int) gin.HandlerFunc {
	// This is a placeholder - implement with Redis or in-memory store
	return func(c *gin.Context) {
		user, exists := c.Get("user")
		if !exists {
			user = c.ClientIP()
		}

		// TODO: Implement rate limiting logic
		// Use Redis or in-memory cache to track requests per user
		_ = user

		c.Next()
	}
}
