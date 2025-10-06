// Copyright 2025 Bedrock Proxy Authors
// SPDX-License-Identifier: Apache-2.0

package auth

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

// APIKey represents an API key in the database
type APIKey struct {
	ID          int64
	KeyHash     string
	Name        string
	Email       string
	Description string
	IsActive    bool
	CreatedAt   time.Time
	LastUsedAt  *time.Time
	ExpiresAt   *time.Time
	Permissions string // JSON array of permissions
	Metadata    string // JSON metadata
}

// APIKeyDB manages API keys in SQLite
type APIKeyDB struct {
	db *sql.DB
}

// NewAPIKeyDB creates a new API key database
func NewAPIKeyDB(dbPath string) (*APIKeyDB, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Create schema
	schema := `
	CREATE TABLE IF NOT EXISTS api_keys (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		key_hash TEXT NOT NULL UNIQUE,
		name TEXT NOT NULL,
		email TEXT,
		description TEXT,
		is_active BOOLEAN DEFAULT 1,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		last_used_at TIMESTAMP,
		expires_at TIMESTAMP,
		permissions TEXT DEFAULT '[]',
		metadata TEXT DEFAULT '{}'
	);

	CREATE INDEX IF NOT EXISTS idx_key_hash ON api_keys(key_hash);
	CREATE INDEX IF NOT EXISTS idx_email ON api_keys(email);
	CREATE INDEX IF NOT EXISTS idx_is_active ON api_keys(is_active);

	-- Audit log table
	CREATE TABLE IF NOT EXISTS api_key_audit (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		api_key_id INTEGER,
		action TEXT NOT NULL,
		ip_address TEXT,
		user_agent TEXT,
		request_path TEXT,
		status_code INTEGER,
		timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		metadata TEXT DEFAULT '{}'
	);

	CREATE INDEX IF NOT EXISTS idx_audit_key_id ON api_key_audit(api_key_id);
	CREATE INDEX IF NOT EXISTS idx_audit_timestamp ON api_key_audit(timestamp);

	-- 2FA table
	CREATE TABLE IF NOT EXISTS api_key_2fa (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		api_key_id INTEGER NOT NULL UNIQUE,
		totp_secret TEXT NOT NULL,
		backup_codes TEXT,
		is_enabled BOOLEAN DEFAULT 0,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (api_key_id) REFERENCES api_keys(id) ON DELETE CASCADE
	);
	`

	if _, err := db.Exec(schema); err != nil {
		return nil, fmt.Errorf("failed to create schema: %w", err)
	}

	return &APIKeyDB{db: db}, nil
}

// GenerateAPIKey creates a new secure API key
func (db *APIKeyDB) GenerateAPIKey(name, email, description string, expiresIn *time.Duration) (string, error) {
	// Generate secure random key
	keyBytes := make([]byte, 32)
	if _, err := rand.Read(keyBytes); err != nil {
		return "", fmt.Errorf("failed to generate random key: %w", err)
	}
	apiKey := "bdrk_" + hex.EncodeToString(keyBytes)

	// Hash the key for storage (bcrypt)
	hash, err := bcrypt.GenerateFromPassword([]byte(apiKey), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash key: %w", err)
	}

	// Calculate expiration
	var expiresAt *time.Time
	if expiresIn != nil {
		exp := time.Now().Add(*expiresIn)
		expiresAt = &exp
	}

	// Insert into database
	_, err = db.db.Exec(`
		INSERT INTO api_keys (key_hash, name, email, description, expires_at)
		VALUES (?, ?, ?, ?, ?)
	`, string(hash), name, email, description, expiresAt)

	if err != nil {
		return "", fmt.Errorf("failed to insert API key: %w", err)
	}

	return apiKey, nil
}

// ValidateAPIKey checks if an API key is valid and returns the key info
func (db *APIKeyDB) ValidateAPIKey(apiKey string) (*APIKey, error) {
	// Get all active keys
	rows, err := db.db.Query(`
		SELECT id, key_hash, name, email, description, is_active, created_at, last_used_at, expires_at, permissions, metadata
		FROM api_keys
		WHERE is_active = 1
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to query keys: %w", err)
	}
	defer rows.Close()

	// Check each key with constant-time comparison
	for rows.Next() {
		var key APIKey
		var lastUsed, expires sql.NullTime

		err := rows.Scan(
			&key.ID, &key.KeyHash, &key.Name, &key.Email, &key.Description,
			&key.IsActive, &key.CreatedAt, &lastUsed, &expires,
			&key.Permissions, &key.Metadata,
		)
		if err != nil {
			continue
		}

		if lastUsed.Valid {
			key.LastUsedAt = &lastUsed.Time
		}
		if expires.Valid {
			key.ExpiresAt = &expires.Time
		}

		// Check if key matches (constant-time comparison via bcrypt)
		if err := bcrypt.CompareHashAndPassword([]byte(key.KeyHash), []byte(apiKey)); err == nil {
			// Check expiration
			if key.ExpiresAt != nil && time.Now().After(*key.ExpiresAt) {
				return nil, fmt.Errorf("API key expired")
			}

			// Update last used timestamp
			db.db.Exec("UPDATE api_keys SET last_used_at = ? WHERE id = ?", time.Now(), key.ID)

			return &key, nil
		}
	}

	return nil, fmt.Errorf("invalid API key")
}

// RevokeAPIKey deactivates an API key
func (db *APIKeyDB) RevokeAPIKey(keyID int64) error {
	_, err := db.db.Exec("UPDATE api_keys SET is_active = 0 WHERE id = ?", keyID)
	if err != nil {
		return fmt.Errorf("failed to revoke key: %w", err)
	}
	return nil
}

// ListAPIKeys returns all API keys (for admin)
func (db *APIKeyDB) ListAPIKeys() ([]APIKey, error) {
	rows, err := db.db.Query(`
		SELECT id, key_hash, name, email, description, is_active, created_at, last_used_at, expires_at, permissions, metadata
		FROM api_keys
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to query keys: %w", err)
	}
	defer rows.Close()

	var keys []APIKey
	for rows.Next() {
		var key APIKey
		var lastUsed, expires sql.NullTime

		err := rows.Scan(
			&key.ID, &key.KeyHash, &key.Name, &key.Email, &key.Description,
			&key.IsActive, &key.CreatedAt, &lastUsed, &expires,
			&key.Permissions, &key.Metadata,
		)
		if err != nil {
			continue
		}

		if lastUsed.Valid {
			key.LastUsedAt = &lastUsed.Time
		}
		if expires.Valid {
			key.ExpiresAt = &expires.Time
		}

		keys = append(keys, key)
	}

	return keys, nil
}

// LogAPIKeyUsage records API key usage for audit
func (db *APIKeyDB) LogAPIKeyUsage(keyID int64, action, ip, userAgent, path string, statusCode int, metadata string) error {
	_, err := db.db.Exec(`
		INSERT INTO api_key_audit (api_key_id, action, ip_address, user_agent, request_path, status_code, metadata)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, keyID, action, ip, userAgent, path, statusCode, metadata)

	return err
}

// GetAPIKeyByEmail returns API key info by email
func (db *APIKeyDB) GetAPIKeyByEmail(email string) (*APIKey, error) {
	var key APIKey
	var lastUsed, expires sql.NullTime

	err := db.db.QueryRow(`
		SELECT id, key_hash, name, email, description, is_active, created_at, last_used_at, expires_at, permissions, metadata
		FROM api_keys
		WHERE email = ? AND is_active = 1
		LIMIT 1
	`, email).Scan(
		&key.ID, &key.KeyHash, &key.Name, &key.Email, &key.Description,
		&key.IsActive, &key.CreatedAt, &lastUsed, &expires,
		&key.Permissions, &key.Metadata,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("no active API key found for email: %s", email)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get API key: %w", err)
	}

	if lastUsed.Valid {
		key.LastUsedAt = &lastUsed.Time
	}
	if expires.Valid {
		key.ExpiresAt = &expires.Time
	}

	return &key, nil
}

// GetAPIKeyByID returns API key info by ID
func (db *APIKeyDB) GetAPIKeyByID(id int64) (*APIKey, error) {
	var key APIKey
	var lastUsed, expires sql.NullTime

	err := db.db.QueryRow(`
		SELECT id, key_hash, name, email, description, is_active, created_at, last_used_at, expires_at, permissions, metadata
		FROM api_keys
		WHERE id = ? AND is_active = 1
		LIMIT 1
	`, id).Scan(
		&key.ID, &key.KeyHash, &key.Name, &key.Email, &key.Description,
		&key.IsActive, &key.CreatedAt, &lastUsed, &expires,
		&key.Permissions, &key.Metadata,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("API key not found: %d", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get API key: %w", err)
	}

	if lastUsed.Valid {
		key.LastUsedAt = &lastUsed.Time
	}
	if expires.Valid {
		key.ExpiresAt = &expires.Time
	}

	return &key, nil
}

// Close closes the database connection
func (db *APIKeyDB) Close() error {
	return db.db.Close()
}
