package models

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"time"
)

// Session represents an authenticated user session
type Session struct {
	ID        string    `json:"id"`
	UserID    int       `json:"user_id"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

// SessionWithUser includes user and company information
type SessionWithUser struct {
	Session
	User    *User
	Company *Company
}

// CreateSession creates a new session for a user
func (db *DB) CreateSession(userID int, duration time.Duration) (*Session, error) {
	// Generate a secure random session ID
	sessionID, err := generateSessionID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate session ID: %w", err)
	}

	expiresAt := time.Now().Add(duration)

	query := `
		INSERT INTO sessions (id, user_id, expires_at)
		VALUES (?, ?, ?)
	`

	_, err = db.Exec(query, sessionID, userID, expiresAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return &Session{
		ID:        sessionID,
		UserID:    userID,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now(),
	}, nil
}

// GetSession retrieves a session by ID
func (db *DB) GetSession(sessionID string) (*Session, error) {
	query := `
		SELECT id, user_id, expires_at, created_at
		FROM sessions
		WHERE id = ?
	`

	var s Session
	err := db.QueryRow(query, sessionID).Scan(
		&s.ID,
		&s.UserID,
		&s.ExpiresAt,
		&s.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	return &s, nil
}

// GetSessionWithUser retrieves a session with associated user and company data
func (db *DB) GetSessionWithUser(sessionID string) (*SessionWithUser, error) {
	// Get the session
	session, err := db.GetSession(sessionID)
	if err != nil {
		return nil, err
	}

	if session == nil {
		return nil, nil
	}

	// Check if session is expired
	if time.Now().After(session.ExpiresAt) {
		// Delete expired session
		db.DeleteSession(sessionID)
		return nil, nil
	}

	// Get the user
	user, err := db.GetUserByID(session.UserID)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, nil
	}

	// Get the company
	company, err := db.GetCompanyByID(user.CompanyID)
	if err != nil {
		return nil, err
	}

	if company == nil {
		return nil, nil
	}

	return &SessionWithUser{
		Session: *session,
		User:    user,
		Company: company,
	}, nil
}

// DeleteSession deletes a session (logout)
func (db *DB) DeleteSession(sessionID string) error {
	query := `DELETE FROM sessions WHERE id = ?`

	_, err := db.Exec(query, sessionID)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	return nil
}

// DeleteUserSessions deletes all sessions for a user
func (db *DB) DeleteUserSessions(userID int) error {
	query := `DELETE FROM sessions WHERE user_id = ?`

	_, err := db.Exec(query, userID)
	if err != nil {
		return fmt.Errorf("failed to delete user sessions: %w", err)
	}

	return nil
}

// CleanExpiredSessions removes expired sessions from the database
func (db *DB) CleanExpiredSessions() error {
	query := `DELETE FROM sessions WHERE expires_at < ?`

	_, err := db.Exec(query, time.Now())
	if err != nil {
		return fmt.Errorf("failed to clean expired sessions: %w", err)
	}

	return nil
}

// generateSessionID generates a cryptographically secure random session ID
func generateSessionID() (string, error) {
	// Generate 32 random bytes (64 hex characters)
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	return hex.EncodeToString(bytes), nil
}

