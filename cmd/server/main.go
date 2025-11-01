// Copyright 2025 Bedrock Proxy Authors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/bedrock-proxy/bedrock-iam-proxy/internal/health"
	"github.com/bedrock-proxy/bedrock-iam-proxy/internal/middleware"
	"github.com/bedrock-proxy/bedrock-iam-proxy/internal/proxy"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	// Configuration from environment
	port := getEnv("PORT", "8080")
	tlsPort := getEnv("TLS_PORT", "8443")
	region := getEnv("AWS_REGION", "sa-east-1")
	ginMode := getEnv("GIN_MODE", "release")
	authEnabled := getEnv("AUTH_ENABLED", "false") == "true"
	authMode := getEnv("AUTH_MODE", "api_key")
	tlsCertFile := getEnv("TLS_CERT_FILE", "/etc/tls/tls.crt")
	tlsKeyFile := getEnv("TLS_KEY_FILE", "/etc/tls/tls.key")
	tlsEnabled := getEnv("TLS_ENABLED", "false") == "true"

	// Set Gin mode
	gin.SetMode(ginMode)

	// Initialize components
	healthChecker := health.NewChecker()
	bedrockProxy, err := proxy.NewBedrockProxy(region, healthChecker)
	if err != nil {
		log.Fatalf("Failed to create Bedrock proxy: %v", err)
	}

	// Initialize Gin router
	router := gin.New()

	// Global middleware
	router.Use(middleware.Recovery())
	router.Use(middleware.RequestID())
	router.Use(middleware.Logger())
	router.Use(middleware.Security())
	router.Use(middleware.Metrics())

	// Health endpoints (no auth required)
	router.GET("/health", healthHandler(healthChecker))
	router.GET("/ready", readyHandler(healthChecker))
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Proxy routes with authentication
	proxyGroup := router.Group("/")

	if authEnabled {
		log.Printf("Authentication enabled: mode=%s", authMode)

		switch authMode {
		case "api_key":
			apiKeys := middleware.LoadAPIKeysFromEnv()
			if len(apiKeys) == 0 {
				log.Fatal("API key auth enabled but no keys found. Set BEDROCK_API_KEY_<NAME> env vars")
			}
			log.Printf("Loaded %d API keys", len(apiKeys))
			proxyGroup.Use(middleware.APIKeyAuth(apiKeys))

		case "basic":
			credentials := loadBasicAuthCredentials()
			if len(credentials) == 0 {
				log.Fatal("Basic auth enabled but no credentials found")
			}
			proxyGroup.Use(middleware.BasicAuth(credentials))

		case "service_account":
			allowedSAs := loadAllowedServiceAccounts()
			if len(allowedSAs) == 0 {
				log.Fatal("Service account auth enabled but no allowed accounts found")
			}
			proxyGroup.Use(middleware.ServiceAccountAuth(allowedSAs))

		default:
			log.Printf("Unknown auth mode: %s, running without auth", authMode)
		}

		// Optional rate limiting
		if getEnv("RATE_LIMIT_ENABLED", "false") == "true" {
			// proxyGroup.Use(middleware.RateLimitByUser(100))
			log.Println("Rate limiting enabled")
		}
	} else {
		log.Println("WARNING: Authentication is DISABLED")
	}

	// Bedrock proxy routes
	proxyGroup.Any("/v1/bedrock/*path", bedrockProxy.Handler())
	proxyGroup.Any("/bedrock/*path", bedrockProxy.Handler())
	proxyGroup.Any("/model/*path", bedrockProxy.Handler())

	// Start server(s)
	if tlsEnabled {
		// Start HTTP server in goroutine
		go func() {
			addr := fmt.Sprintf(":%s", port)
			log.Printf("Starting HTTP server on %s (region: %s)", addr, region)
			if err := router.Run(addr); err != nil {
				log.Fatalf("Failed to start HTTP server: %v", err)
			}
		}()

		// Start HTTPS/TLS server (blocking)
		addrTLS := fmt.Sprintf(":%s", tlsPort)
		log.Printf("Starting HTTPS/TLS server on %s (region: %s)", addrTLS, region)
		if err := router.RunTLS(addrTLS, tlsCertFile, tlsKeyFile); err != nil {
			log.Fatalf("Failed to start HTTPS/TLS server: %v", err)
		}
	} else {
		// Start HTTP server only
		addr := fmt.Sprintf(":%s", port)
		log.Printf("Starting HTTP server on %s (region: %s)", addr, region)
		if err := router.Run(addr); err != nil {
			log.Fatalf("Failed to start HTTP server: %v", err)
		}
	}
}

func healthHandler(checker *health.Checker) gin.HandlerFunc {
	return func(c *gin.Context) {
		if checker.IsHealthy() {
			c.JSON(200, gin.H{
				"status": "healthy",
				"service": "bedrock-proxy",
			})
		} else {
			c.JSON(503, gin.H{
				"status": "unhealthy",
				"service": "bedrock-proxy",
			})
		}
	}
}

func readyHandler(checker *health.Checker) gin.HandlerFunc {
	return func(c *gin.Context) {
		if checker.IsHealthy() {
			c.JSON(200, gin.H{
				"status": "ready",
			})
		} else {
			c.JSON(503, gin.H{
				"status": "not ready",
			})
		}
	}
}

func loadBasicAuthCredentials() map[string]string {
	creds := make(map[string]string)

	// Load from BASIC_AUTH_CREDENTIALS env var (format: user1:pass1,user2:pass2)
	if credsEnv := os.Getenv("BASIC_AUTH_CREDENTIALS"); credsEnv != "" {
		for _, pair := range strings.Split(credsEnv, ",") {
			parts := strings.SplitN(pair, ":", 2)
			if len(parts) == 2 {
				creds[parts[0]] = parts[1]
			}
		}
	}

	return creds
}

func loadAllowedServiceAccounts() []string {
	var accounts []string

	// Load from ALLOWED_SERVICE_ACCOUNTS env var (format: ns1/sa1,ns2/sa2)
	if sasEnv := os.Getenv("ALLOWED_SERVICE_ACCOUNTS"); sasEnv != "" {
		for _, sa := range strings.Split(sasEnv, ",") {
			sa = strings.TrimSpace(sa)
			if sa != "" {
				accounts = append(accounts, sa)
			}
		}
	}

	return accounts
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
