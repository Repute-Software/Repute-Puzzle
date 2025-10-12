package models

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

// Completion represents a puzzle completion record
type Completion struct {
	ID           int       `json:"id"`
	Email        string    `json:"email"`
	DiscountCode string    `json:"discount_code"`
	GridSize     int       `json:"grid_size"`
	Moves        int       `json:"moves"`
	TimeSeconds  int       `json:"time_seconds"`
	CreatedAt    time.Time `json:"created_at"`
}

// GenerateDiscountCode creates a unique discount code
// Format: PUZZLE{discount%}-{random}
func GenerateDiscountCode(discountPercent int) (string, error) {
	// Generate 6 random bytes (12 hex characters)
	bytes := make([]byte, 6)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	randomStr := strings.ToUpper(hex.EncodeToString(bytes))
	code := fmt.Sprintf("PUZZLE%d-%s", discountPercent, randomStr)
	return code, nil
}

// SaveCompletion stores a completion record in the database
func (db *DB) SaveCompletion(email string, discountCode string, gridSize, moves, timeSeconds int) error {
	query := `
		INSERT INTO completions (email, discount_code, grid_size, moves, time_seconds)
		VALUES (?, ?, ?, ?, ?)
	`

	_, err := db.Exec(query, email, discountCode, gridSize, moves, timeSeconds)
	if err != nil {
		return fmt.Errorf("failed to save completion: %w", err)
	}

	return nil
}

// GetCompletionByCode retrieves a completion by discount code
func (db *DB) GetCompletionByCode(code string) (*Completion, error) {
	query := `
		SELECT id, email, discount_code, grid_size, moves, time_seconds, created_at
		FROM completions
		WHERE discount_code = ?
	`

	var c Completion
	err := db.QueryRow(query, code).Scan(
		&c.ID,
		&c.Email,
		&c.DiscountCode,
		&c.GridSize,
		&c.Moves,
		&c.TimeSeconds,
		&c.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get completion: %w", err)
	}

	return &c, nil
}

// GetAllCompletions retrieves all completion records
func (db *DB) GetAllCompletions() ([]Completion, error) {
	query := `
		SELECT id, email, discount_code, grid_size, moves, time_seconds, created_at
		FROM completions
		ORDER BY created_at DESC
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query completions: %w", err)
	}
	defer rows.Close()

	var completions []Completion
	for rows.Next() {
		var c Completion
		if err := rows.Scan(
			&c.ID,
			&c.Email,
			&c.DiscountCode,
			&c.GridSize,
			&c.Moves,
			&c.TimeSeconds,
			&c.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan completion: %w", err)
		}
		completions = append(completions, c)
	}

	return completions, nil
}
