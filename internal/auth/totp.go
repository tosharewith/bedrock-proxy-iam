// Copyright 2025 Bedrock Proxy Authors
// SPDX-License-Identifier: Apache-2.0

package auth

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

// TOTPManager manages TOTP (Time-based One-Time Passwords) for 2FA
type TOTPManager struct {
	db *sql.DB
}

// NewTOTPManager creates a new TOTP manager
func NewTOTPManager(db *sql.DB) *TOTPManager {
	return &TOTPManager{db: db}
}

// GenerateTOTP creates a new TOTP secret for a user
func (m *TOTPManager) GenerateTOTP(apiKeyID int64, accountName, issuer string) (*otp.Key, []string, error) {
	// Generate TOTP key
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      issuer,
		AccountName: accountName,
		Period:      30,
		Digits:      otp.DigitsSix,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate TOTP: %w", err)
	}

	// Generate backup codes
	backupCodes := make([]string, 10)
	for i := 0; i < 10; i++ {
		code, err := generateBackupCode()
		if err != nil {
			return nil, nil, fmt.Errorf("failed to generate backup code: %w", err)
		}
		backupCodes[i] = code
	}

	// Store in database
	backupCodesStr := strings.Join(backupCodes, ",")
	_, err = m.db.Exec(`
		INSERT INTO api_key_2fa (api_key_id, totp_secret, backup_codes, is_enabled)
		VALUES (?, ?, ?, 1)
		ON CONFLICT(api_key_id) DO UPDATE SET
			totp_secret = excluded.totp_secret,
			backup_codes = excluded.backup_codes,
			is_enabled = 1
	`, apiKeyID, key.Secret(), backupCodesStr)

	if err != nil {
		return nil, nil, fmt.Errorf("failed to store TOTP: %w", err)
	}

	return key, backupCodes, nil
}

// ValidateTOTP validates a TOTP code for an API key
func (m *TOTPManager) ValidateTOTP(apiKeyID int64, code string) (bool, error) {
	var secret string
	var backupCodes string
	var isEnabled bool

	err := m.db.QueryRow(`
		SELECT totp_secret, backup_codes, is_enabled
		FROM api_key_2fa
		WHERE api_key_id = ?
	`, apiKeyID).Scan(&secret, &backupCodes, &isEnabled)

	if err == sql.ErrNoRows {
		return false, fmt.Errorf("2FA not configured for this API key")
	}
	if err != nil {
		return false, fmt.Errorf("failed to get TOTP: %w", err)
	}

	if !isEnabled {
		return false, fmt.Errorf("2FA is disabled for this API key")
	}

	// Try TOTP code first
	valid := totp.Validate(code, secret)
	if valid {
		return true, nil
	}

	// Try backup codes
	codes := strings.Split(backupCodes, ",")
	for i, backupCode := range codes {
		if backupCode == code {
			// Remove used backup code
			codes = append(codes[:i], codes[i+1:]...)
			newBackupCodes := strings.Join(codes, ",")

			_, err := m.db.Exec(`
				UPDATE api_key_2fa
				SET backup_codes = ?
				WHERE api_key_id = ?
			`, newBackupCodes, apiKeyID)

			if err != nil {
				return false, fmt.Errorf("failed to update backup codes: %w", err)
			}

			return true, nil
		}
	}

	return false, fmt.Errorf("invalid TOTP code")
}

// DisableTOTP disables 2FA for an API key
func (m *TOTPManager) DisableTOTP(apiKeyID int64) error {
	_, err := m.db.Exec(`
		UPDATE api_key_2fa
		SET is_enabled = 0
		WHERE api_key_id = ?
	`, apiKeyID)

	return err
}

// IsTOTPEnabled checks if 2FA is enabled for an API key
func (m *TOTPManager) IsTOTPEnabled(apiKeyID int64) (bool, error) {
	var isEnabled bool

	err := m.db.QueryRow(`
		SELECT is_enabled
		FROM api_key_2fa
		WHERE api_key_id = ?
	`, apiKeyID).Scan(&isEnabled)

	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return isEnabled, nil
}

// generateBackupCode creates a random backup code
func generateBackupCode() (string, error) {
	bytes := make([]byte, 6)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	// Convert to base32-like string (uppercase letters + numbers)
	code := base64.RawStdEncoding.EncodeToString(bytes)
	code = strings.ToUpper(strings.ReplaceAll(code, "=", ""))

	// Format as XXXX-XXXX
	if len(code) >= 8 {
		code = code[:4] + "-" + code[4:8]
	}

	return code, nil
}
