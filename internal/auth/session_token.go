// Copyright 2025 Bedrock Proxy Authors
// SPDX-License-Identifier: Apache-2.0

package auth

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"fmt"
	"time"
)

// SessionToken represents a temporary session token
type SessionToken struct {
	ID         int64
	Token      string
	APIKeyID   int64
	CreatedAt  time.Time
	ExpiresAt  time.Time
	LastUsedAt *time.Time
	IPAddress  string
	UserAgent  string
	IsActive   bool
}

// SessionManager manages session tokens
type SessionManager struct {
	db *sql.DB
}

// NewSessionManager creates a new session manager
func NewSessionManager(db *sql.DB) *SessionManager {
	// Create session tokens table
	schema := `
	CREATE TABLE IF NOT EXISTS session_tokens (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		token TEXT NOT NULL UNIQUE,
		api_key_id INTEGER NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		expires_at TIMESTAMP NOT NULL,
		last_used_at TIMESTAMP,
		ip_address TEXT,
		user_agent TEXT,
		is_active BOOLEAN DEFAULT 1,
		FOREIGN KEY (api_key_id) REFERENCES api_keys(id) ON DELETE CASCADE
	);

	CREATE INDEX IF NOT EXISTS idx_session_token ON session_tokens(token);
	CREATE INDEX IF NOT EXISTS idx_session_active ON session_tokens(is_active, expires_at);
	`

	db.Exec(schema)

	return &SessionManager{db: db}
}

// GenerateSessionToken creates a new session token after successful auth
func (m *SessionManager) GenerateSessionToken(
	apiKeyID int64,
	duration time.Duration,
	ipAddress, userAgent string,
) (string, error) {
	// Generate secure random token (32 bytes = 44 chars base64)
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}

	// Create alphanumeric token with prefix
	token := "bdrk_sess_" + base64.URLEncoding.EncodeToString(tokenBytes)

	// Calculate expiration
	expiresAt := time.Now().Add(duration)

	// Insert into database
	_, err := m.db.Exec(`
		INSERT INTO session_tokens (token, api_key_id, expires_at, ip_address, user_agent)
		VALUES (?, ?, ?, ?, ?)
	`, token, apiKeyID, expiresAt, ipAddress, userAgent)

	if err != nil {
		return "", fmt.Errorf("failed to store session token: %w", err)
	}

	return token, nil
}

// ValidateSessionToken checks if a session token is valid
func (m *SessionManager) ValidateSessionToken(token string) (*SessionToken, int64, error) {
	var session SessionToken
	var apiKeyID int64
	var lastUsed sql.NullTime

	err := m.db.QueryRow(`
		SELECT
			st.id, st.token, st.api_key_id, st.created_at, st.expires_at,
			st.last_used_at, st.ip_address, st.user_agent, st.is_active,
			ak.id as api_key_id
		FROM session_tokens st
		JOIN api_keys ak ON st.api_key_id = ak.id
		WHERE st.token = ? AND st.is_active = 1
	`, token).Scan(
		&session.ID, &session.Token, &session.APIKeyID,
		&session.CreatedAt, &session.ExpiresAt, &lastUsed,
		&session.IPAddress, &session.UserAgent, &session.IsActive,
		&apiKeyID,
	)

	if err == sql.ErrNoRows {
		return nil, 0, fmt.Errorf("invalid session token")
	}
	if err != nil {
		return nil, 0, fmt.Errorf("failed to validate token: %w", err)
	}

	if lastUsed.Valid {
		session.LastUsedAt = &lastUsed.Time
	}

	// Check expiration
	if time.Now().After(session.ExpiresAt) {
		return nil, 0, fmt.Errorf("session token expired")
	}

	// Update last used timestamp
	m.db.Exec("UPDATE session_tokens SET last_used_at = ? WHERE id = ?", time.Now(), session.ID)

	return &session, apiKeyID, nil
}

// RevokeSessionToken invalidates a session token
func (m *SessionManager) RevokeSessionToken(token string) error {
	_, err := m.db.Exec("UPDATE session_tokens SET is_active = 0 WHERE token = ?", token)
	return err
}

// RevokeAllUserSessions revokes all sessions for a specific API key
func (m *SessionManager) RevokeAllUserSessions(apiKeyID int64) error {
	_, err := m.db.Exec("UPDATE session_tokens SET is_active = 0 WHERE api_key_id = ?", apiKeyID)
	return err
}

// CleanupExpiredSessions removes expired session tokens
func (m *SessionManager) CleanupExpiredSessions() error {
	_, err := m.db.Exec("DELETE FROM session_tokens WHERE expires_at < ?", time.Now())
	return err
}

// ListUserSessions returns active sessions for an API key
func (m *SessionManager) ListUserSessions(apiKeyID int64) ([]SessionToken, error) {
	rows, err := m.db.Query(`
		SELECT id, token, api_key_id, created_at, expires_at, last_used_at, ip_address, user_agent, is_active
		FROM session_tokens
		WHERE api_key_id = ? AND is_active = 1 AND expires_at > ?
		ORDER BY created_at DESC
	`, apiKeyID, time.Now())

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []SessionToken
	for rows.Next() {
		var s SessionToken
		var lastUsed sql.NullTime

		err := rows.Scan(
			&s.ID, &s.Token, &s.APIKeyID, &s.CreatedAt, &s.ExpiresAt,
			&lastUsed, &s.IPAddress, &s.UserAgent, &s.IsActive,
		)
		if err != nil {
			continue
		}

		if lastUsed.Valid {
			s.LastUsedAt = &lastUsed.Time
		}

		sessions = append(sessions, s)
	}

	return sessions, nil
}
