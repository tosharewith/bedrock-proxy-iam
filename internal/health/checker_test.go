package health

import (
	"testing"
)

func TestNewChecker(t *testing.T) {
	checker := NewChecker()

	if !checker.IsHealthy() {
		t.Error("New checker should be healthy")
	}

	if !checker.IsReady() {
		t.Error("New checker should be ready")
	}
}

func TestRecordError(t *testing.T) {
	checker := NewChecker()

	// Record some successes first to establish a baseline
	for i := 0; i < 10; i++ {
		checker.RecordSuccess()
	}

	// Record some errors
	for i := 0; i < 10; i++ {
		checker.RecordError()
	}

	// Should still be healthy with balanced errors/successes
	if !checker.IsHealthy() {
		t.Error("Checker should still be healthy with balanced errors/successes")
	}

	// Record many more errors to tip the balance
	for i := 0; i < 100; i++ {
		checker.RecordError()
	}

	// Now should be unhealthy due to high error rate
	if checker.IsHealthy() {
		t.Error("Checker should be unhealthy with high error rate")
	}
}

func TestRecordSuccess(t *testing.T) {
	checker := NewChecker()

	// Make unhealthy first
	for i := 0; i < 100; i++ {
		checker.RecordError()
	}

	if checker.IsHealthy() {
		t.Error("Checker should be unhealthy")
	}

	// Record success should make it healthy again
	checker.RecordSuccess()

	if !checker.IsHealthy() {
		t.Error("Checker should be healthy after success")
	}
}

func TestReadiness(t *testing.T) {
	checker := NewChecker()

	if !checker.IsReady() {
		t.Error("Checker should be ready by default")
	}

	checker.SetReady(false)
	if checker.IsReady() {
		t.Error("Checker should not be ready after SetReady(false)")
	}

	checker.SetReady(true)
	if !checker.IsReady() {
		t.Error("Checker should be ready after SetReady(true)")
	}
}
