package auth

import (
	"testing"
)

func TestNewAWSSigner(t *testing.T) {
	signer, err := NewAWSSigner("us-east-1", "bedrock")
	if err != nil {
		t.Fatalf("Failed to create AWS signer: %v", err)
	}

	if signer.region != "us-east-1" {
		t.Errorf("Expected region us-east-1, got %s", signer.region)
	}

	if signer.service != "bedrock" {
		t.Errorf("Expected service bedrock, got %s", signer.service)
	}
}
