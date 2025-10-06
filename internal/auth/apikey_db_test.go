// Copyright 2025 Bedrock Proxy Authors
// SPDX-License-Identifier: Apache-2.0

package auth

import (
	"os"
	"testing"
	"time"

	"github.com/pquerna/otp/totp"
)

func TestAPIKeyDB(t *testing.T) {
	// Create temporary database
	dbPath := "/tmp/test_apikeys.db"
	defer os.Remove(dbPath)

	db, err := NewAPIKeyDB(dbPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	t.Run("GenerateAPIKey", func(t *testing.T) {
		apiKey, err := db.GenerateAPIKey("Test User", "test@example.com", "Test key", nil)
		if err != nil {
			t.Fatalf("Failed to generate API key: %v", err)
		}

		if apiKey[:5] != "bdrk_" {
			t.Errorf("API key should start with 'bdrk_', got: %s", apiKey[:5])
		}

		if len(apiKey) != 69 { // bdrk_ (5) + 64 hex chars
			t.Errorf("API key should be 69 chars, got: %d", len(apiKey))
		}
	})

	t.Run("ValidateAPIKey", func(t *testing.T) {
		apiKey, err := db.GenerateAPIKey("Validate User", "validate@example.com", "Validation test", nil)
		if err != nil {
			t.Fatalf("Failed to generate API key: %v", err)
		}

		keyInfo, err := db.ValidateAPIKey(apiKey)
		if err != nil {
			t.Fatalf("Failed to validate API key: %v", err)
		}

		if keyInfo.Name != "Validate User" {
			t.Errorf("Expected name 'Validate User', got: %s", keyInfo.Name)
		}

		if keyInfo.Email != "validate@example.com" {
			t.Errorf("Expected email 'validate@example.com', got: %s", keyInfo.Email)
		}

		// Test invalid key
		_, err = db.ValidateAPIKey("bdrk_invalid_key_that_does_not_exist")
		if err == nil {
			t.Error("Expected error for invalid API key, got nil")
		}
	})

	t.Run("ExpiredAPIKey", func(t *testing.T) {
		expiration := time.Duration(-1 * time.Hour) // Already expired
		apiKey, err := db.GenerateAPIKey("Expired User", "expired@example.com", "Expired key", &expiration)
		if err != nil {
			t.Fatalf("Failed to generate API key: %v", err)
		}

		_, err = db.ValidateAPIKey(apiKey)
		if err == nil {
			t.Error("Expected error for expired API key, got nil")
		}
	})

	t.Run("RevokeAPIKey", func(t *testing.T) {
		apiKey, err := db.GenerateAPIKey("Revoke User", "revoke@example.com", "Revoke test", nil)
		if err != nil {
			t.Fatalf("Failed to generate API key: %v", err)
		}

		keyInfo, err := db.ValidateAPIKey(apiKey)
		if err != nil {
			t.Fatalf("Failed to validate API key: %v", err)
		}

		// Revoke the key
		err = db.RevokeAPIKey(keyInfo.ID)
		if err != nil {
			t.Fatalf("Failed to revoke API key: %v", err)
		}

		// Try to validate revoked key
		_, err = db.ValidateAPIKey(apiKey)
		if err == nil {
			t.Error("Expected error for revoked API key, got nil")
		}
	})

	t.Run("ListAPIKeys", func(t *testing.T) {
		keys, err := db.ListAPIKeys()
		if err != nil {
			t.Fatalf("Failed to list API keys: %v", err)
		}

		if len(keys) < 1 {
			t.Error("Expected at least 1 API key in the list")
		}
	})

	t.Run("GetAPIKeyByEmail", func(t *testing.T) {
		email := "unique@example.com"
		_, err := db.GenerateAPIKey("Email User", email, "Email test", nil)
		if err != nil {
			t.Fatalf("Failed to generate API key: %v", err)
		}

		keyInfo, err := db.GetAPIKeyByEmail(email)
		if err != nil {
			t.Fatalf("Failed to get API key by email: %v", err)
		}

		if keyInfo.Email != email {
			t.Errorf("Expected email %s, got: %s", email, keyInfo.Email)
		}
	})

	t.Run("AuditLog", func(t *testing.T) {
		apiKey, err := db.GenerateAPIKey("Audit User", "audit@example.com", "Audit test", nil)
		if err != nil {
			t.Fatalf("Failed to generate API key: %v", err)
		}

		keyInfo, err := db.ValidateAPIKey(apiKey)
		if err != nil {
			t.Fatalf("Failed to validate API key: %v", err)
		}

		err = db.LogAPIKeyUsage(
			keyInfo.ID,
			"test_action",
			"192.168.1.1",
			"test-agent",
			"/test/path",
			200,
			`{"test":"data"}`,
		)

		if err != nil {
			t.Fatalf("Failed to log API key usage: %v", err)
		}
	})
}

func TestTOTP(t *testing.T) {
	// Create temporary database
	dbPath := "/tmp/test_totp.db"
	defer os.Remove(dbPath)

	apiKeyDB, err := NewAPIKeyDB(dbPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer apiKeyDB.Close()

	totpManager := NewTOTPManager(apiKeyDB.db)

	// Generate API key for testing
	apiKey, err := apiKeyDB.GenerateAPIKey("TOTP User", "totp@example.com", "TOTP test", nil)
	if err != nil {
		t.Fatalf("Failed to generate API key: %v", err)
	}

	keyInfo, err := apiKeyDB.ValidateAPIKey(apiKey)
	if err != nil {
		t.Fatalf("Failed to validate API key: %v", err)
	}

	t.Run("GenerateTOTP", func(t *testing.T) {
		key, backupCodes, err := totpManager.GenerateTOTP(keyInfo.ID, "totp@example.com", "Bedrock Proxy")
		if err != nil {
			t.Fatalf("Failed to generate TOTP: %v", err)
		}

		if key.Secret() == "" {
			t.Error("TOTP secret should not be empty")
		}

		if len(backupCodes) != 10 {
			t.Errorf("Expected 10 backup codes, got: %d", len(backupCodes))
		}
	})

	t.Run("ValidateTOTP", func(t *testing.T) {
		// Generate new TOTP
		key, _, err := totpManager.GenerateTOTP(keyInfo.ID, "totp@example.com", "Bedrock Proxy")
		if err != nil {
			t.Fatalf("Failed to generate TOTP: %v", err)
		}

		// Generate current code
		code, err := totp.GenerateCode(key.Secret(), time.Now())
		if err != nil {
			t.Fatalf("Failed to generate TOTP code: %v", err)
		}

		// Validate code
		valid, err := totpManager.ValidateTOTP(keyInfo.ID, code)
		if err != nil {
			t.Fatalf("Failed to validate TOTP: %v", err)
		}

		if !valid {
			t.Error("TOTP code should be valid")
		}

		// Test invalid code
		valid, err = totpManager.ValidateTOTP(keyInfo.ID, "000000")
		if valid {
			t.Error("Invalid TOTP code should not be valid")
		}
	})

	t.Run("IsTOTPEnabled", func(t *testing.T) {
		enabled, err := totpManager.IsTOTPEnabled(keyInfo.ID)
		if err != nil {
			t.Fatalf("Failed to check TOTP status: %v", err)
		}

		if !enabled {
			t.Error("TOTP should be enabled")
		}
	})

	t.Run("DisableTOTP", func(t *testing.T) {
		err := totpManager.DisableTOTP(keyInfo.ID)
		if err != nil {
			t.Fatalf("Failed to disable TOTP: %v", err)
		}

		enabled, err := totpManager.IsTOTPEnabled(keyInfo.ID)
		if err != nil {
			t.Fatalf("Failed to check TOTP status: %v", err)
		}

		if enabled {
			t.Error("TOTP should be disabled")
		}
	})
}
