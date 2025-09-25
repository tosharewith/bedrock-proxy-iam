package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// RequestDuration tracks request duration for Bedrock API calls
	RequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "bedrock_proxy_request_duration_seconds",
			Help: "Duration of Bedrock proxy requests in seconds",
			Buckets: prometheus.ExponentialBuckets(0.001, 2, 15), // 1ms to ~32s
		},
		[]string{"method", "status"},
	)

	// RequestsTotal tracks total number of requests
	RequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "bedrock_proxy_requests_total",
			Help: "Total number of Bedrock proxy requests",
		},
		[]string{"method", "status"},
	)

	// HTTPRequestDuration tracks HTTP request duration
	HTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "http_request_duration_seconds",
			Help: "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	// HTTPRequestsTotal tracks total HTTP requests
	HTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	// HTTPRequestErrors tracks HTTP errors
	HTTPRequestErrors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_request_errors_total",
			Help: "Total number of HTTP request errors",
		},
		[]string{"method", "path"},
	)

	// AWSCredentialRetrievals tracks AWS credential retrievals
	AWSCredentialRetrievals = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "aws_credential_retrievals_total",
			Help: "Total number of AWS credential retrievals",
		},
		[]string{"method", "status"},
	)

	// BedrockModelInvocations tracks Bedrock model invocations
	BedrockModelInvocations = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "bedrock_model_invocations_total",
			Help: "Total number of Bedrock model invocations",
		},
		[]string{"model", "status"},
	)

	// BedrockTokensProcessed tracks tokens processed
	BedrockTokensProcessed = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "bedrock_tokens_processed_total",
			Help: "Total number of tokens processed by Bedrock",
		},
		[]string{"model", "type"}, // type: input/output
	)

	// ConnectedClients tracks number of connected clients
	ConnectedClients = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "connected_clients",
			Help: "Number of currently connected clients",
		},
	)

	// HealthCheckStatus tracks health check results
	HealthCheckStatus = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "health_check_status",
			Help: "Health check status (1 = healthy, 0 = unhealthy)",
		},
		[]string{"check_type"}, // health, readiness
	)
)

// Init initializes metrics (can be used for custom setup if needed)
func Init() {
	// Register custom metrics or perform initialization if needed
	// For now, promauto handles registration automatically
}

// RecordModelInvocation records a Bedrock model invocation
func RecordModelInvocation(modelID, status string) {
	BedrockModelInvocations.WithLabelValues(modelID, status).Inc()
}

// RecordTokensProcessed records tokens processed by a model
func RecordTokensProcessed(modelID, tokenType string, count int) {
	BedrockTokensProcessed.WithLabelValues(modelID, tokenType).Add(float64(count))
}

// RecordCredentialRetrieval records AWS credential retrieval
func RecordCredentialRetrieval(method, status string) {
	AWSCredentialRetrievals.WithLabelValues(method, status).Inc()
}

// SetHealthStatus sets health check status
func SetHealthStatus(checkType string, healthy bool) {
	var value float64
	if healthy {
		value = 1
	}
	HealthCheckStatus.WithLabelValues(checkType).Set(value)
}

// SetConnectedClients sets the number of connected clients
func SetConnectedClients(count int) {
	ConnectedClients.Set(float64(count))
}