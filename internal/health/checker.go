package health

import (
	"net/http"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
)

// Checker provides health and readiness checking functionality
type Checker struct {
	healthy     int32
	ready       int32
	errors      int64
	successes   int64
	lastError   time.Time
	lastSuccess time.Time
	startTime   time.Time
}

// NewChecker creates a new health checker
func NewChecker() *Checker {
	checker := &Checker{
		healthy:     1,
		ready:       1,
		startTime:   time.Now(),
		lastSuccess: time.Now(),
	}
	return checker
}

// IsHealthy returns true if the service is healthy
func (c *Checker) IsHealthy() bool {
	return atomic.LoadInt32(&c.healthy) == 1
}

// IsReady returns true if the service is ready to serve traffic
func (c *Checker) IsReady() bool {
	return atomic.LoadInt32(&c.ready) == 1
}

// RecordError records a service error
func (c *Checker) RecordError() {
	atomic.AddInt64(&c.errors, 1)
	c.lastError = time.Now()

	// Mark as unhealthy if error rate is too high
	errorRate := c.getErrorRate()
	if errorRate > 0.5 { // More than 50% errors
		atomic.StoreInt32(&c.healthy, 0)
	}
}

// RecordSuccess records a successful operation
func (c *Checker) RecordSuccess() {
	atomic.AddInt64(&c.successes, 1)
	c.lastSuccess = time.Now()

	// Mark as healthy if we have recent success
	atomic.StoreInt32(&c.healthy, 1)
}

// SetReady sets the readiness state
func (c *Checker) SetReady(ready bool) {
	if ready {
		atomic.StoreInt32(&c.ready, 1)
	} else {
		atomic.StoreInt32(&c.ready, 0)
	}
}

// getErrorRate calculates the current error rate
func (c *Checker) getErrorRate() float64 {
	errors := atomic.LoadInt64(&c.errors)
	successes := atomic.LoadInt64(&c.successes)
	total := errors + successes

	if total == 0 {
		return 0.0
	}

	return float64(errors) / float64(total)
}

// GetStats returns health statistics
func (c *Checker) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"healthy":      c.IsHealthy(),
		"ready":        c.IsReady(),
		"errors":       atomic.LoadInt64(&c.errors),
		"successes":    atomic.LoadInt64(&c.successes),
		"error_rate":   c.getErrorRate(),
		"uptime":       time.Since(c.startTime).String(),
		"last_error":   c.lastError.Format(time.RFC3339),
		"last_success": c.lastSuccess.Format(time.RFC3339),
	}
}

// HealthHandler returns a health check handler
func HealthHandler(checker *Checker) gin.HandlerFunc {
	return func(c *gin.Context) {
		if checker.IsHealthy() {
			c.JSON(http.StatusOK, gin.H{
				"status":    "healthy",
				"service":   "bedrock-proxy",
				"timestamp": time.Now().UTC().Format(time.RFC3339),
				"stats":     checker.GetStats(),
			})
		} else {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status":    "unhealthy",
				"service":   "bedrock-proxy",
				"timestamp": time.Now().UTC().Format(time.RFC3339),
				"stats":     checker.GetStats(),
			})
		}
	}
}

// ReadinessHandler returns a readiness check handler
func ReadinessHandler(checker *Checker) gin.HandlerFunc {
	return func(c *gin.Context) {
		if checker.IsReady() {
			c.JSON(http.StatusOK, gin.H{
				"status":    "ready",
				"service":   "bedrock-proxy",
				"timestamp": time.Now().UTC().Format(time.RFC3339),
			})
		} else {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status":    "not ready",
				"service":   "bedrock-proxy",
				"timestamp": time.Now().UTC().Format(time.RFC3339),
			})
		}
	}
}
