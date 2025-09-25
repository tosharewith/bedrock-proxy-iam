package auth

import (
	"os"
	"testing"
)

func TestNewAWSSigner(t *testing.T) {
	// Set up minimal environment for testing
	os.Setenv("AWS_ACCESS_KEY_ID", "test-key")
	os.Setenv("AWS_SECRET_ACCESS_KEY", "test-secret")

	defer func() {
		os.Unsetenv("AWS_ACCESS_KEY_ID")
		os.Unsetenv("AWS_SECRET_ACCESS_KEY")
	}()

	signer, err := NewAWSSigner("us-east-1", "bedrock-runtime")
	if err != nil {
		t.Fatalf("Failed to create AWS signer: %v", err)
	}

	if signer.region != "us-east-1" {
		t.Errorf("Expected region us-east-1, got %s", signer.region)
	}

	if signer.service != "bedrock-runtime" {
		t.Errorf("Expected service bedrock-runtime, got %s", signer.service)
	}
}

func TestHmacSHA256(t *testing.T) {
	signer := &AWSSigner{}

	key := []byte("test-key")
	data := "test-data"

	result := signer.hmacSHA256(key, data)

	if len(result) == 0 {
		t.Error("HMAC result should not be empty")
	}

	// Test consistency
	result2 := signer.hmacSHA256(key, data)
	if string(result) != string(result2) {
		t.Error("HMAC should be consistent for same inputs")
	}
}

func TestGetSignatureKey(t *testing.T) {
	signer := &AWSSigner{}

	dateStamp := "20231025"
	region := "us-east-1"
	service := "bedrock-runtime"
	secretKey := "test-secret"

	key := signer.getSignatureKey(dateStamp, region, service, secretKey)

	if len(key) == 0 {
		t.Error("Signature key should not be empty")
	}

	// Test consistency
	key2 := signer.getSignatureKey(dateStamp, region, service, secretKey)
	if string(key) != string(key2) {
		t.Error("Signature key should be consistent for same inputs")
	}
}
