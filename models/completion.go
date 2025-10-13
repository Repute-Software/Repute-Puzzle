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
	PuzzleID     int       `json:"puzzle_id"`
	Email        string    `json:"email"`
	DiscountCode string    `json:"discount_code"`
	Moves        int       `json:"moves"`
	TimeSeconds  int       `json:"time_seconds"`
	CreatedAt    time.Time `json:"created_at"`
}

// GenerateDiscountCode creates a unique discount code
// Format: PUZZLE{discount%}-{puzzle_id}-{random}
func GenerateDiscountCode(discountPercent, puzzleID int) (string, error) {
	// Generate 6 random bytes (12 hex characters)
	bytes := make([]byte, 6)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	randomStr := strings.ToUpper(hex.EncodeToString(bytes))
	code := fmt.Sprintf("PUZZLE%d-%d-%s", discountPercent, puzzleID, randomStr)
	return code, nil
}

// SaveCompletion stores a completion record in the database
func (db *DB) SaveCompletion(puzzleID int, email string, discountCode string, moves, timeSeconds int) error {
	query := `
		INSERT INTO completions (puzzle_id, email, discount_code, moves, time_seconds)
		VALUES (?, ?, ?, ?, ?)
	`

	_, err := db.Exec(query, puzzleID, email, discountCode, moves, timeSeconds)
	if err != nil {
		return fmt.Errorf("failed to save completion: %w", err)
	}

	return nil
}

// GetCompletionByCode retrieves a completion by discount code
func (db *DB) GetCompletionByCode(code string) (*Completion, error) {
	query := `
		SELECT id, puzzle_id, email, discount_code, moves, time_seconds, created_at
		FROM completions
		WHERE discount_code = ?
	`

	var c Completion
	err := db.QueryRow(query, code).Scan(
		&c.ID,
		&c.PuzzleID,
		&c.Email,
		&c.DiscountCode,
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
		SELECT id, puzzle_id, email, discount_code, moves, time_seconds, created_at
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
			&c.PuzzleID,
			&c.Email,
			&c.DiscountCode,
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

// GetCompletionsByPuzzle retrieves all completions for a specific puzzle
func (db *DB) GetCompletionsByPuzzle(puzzleID int) ([]Completion, error) {
	query := `
		SELECT id, puzzle_id, email, discount_code, moves, time_seconds, created_at
		FROM completions
		WHERE puzzle_id = ?
		ORDER BY created_at DESC
	`

	rows, err := db.Query(query, puzzleID)
	if err != nil {
		return nil, fmt.Errorf("failed to query completions: %w", err)
	}
	defer rows.Close()

	var completions []Completion
	for rows.Next() {
		var c Completion
		if err := rows.Scan(
			&c.ID,
			&c.PuzzleID,
			&c.Email,
			&c.DiscountCode,
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

// GetCompletionsByCompany retrieves all completions for a company's puzzles
func (db *DB) GetCompletionsByCompany(companyID int) ([]Completion, error) {
	query := `
		SELECT c.id, c.puzzle_id, c.email, c.discount_code, c.moves, c.time_seconds, c.created_at
		FROM completions c
		JOIN puzzles p ON p.id = c.puzzle_id
		WHERE p.company_id = ?
		ORDER BY c.created_at DESC
	`

	rows, err := db.Query(query, companyID)
	if err != nil {
		return nil, fmt.Errorf("failed to query completions: %w", err)
	}
	defer rows.Close()

	var completions []Completion
	for rows.Next() {
		var c Completion
		if err := rows.Scan(
			&c.ID,
			&c.PuzzleID,
			&c.Email,
			&c.DiscountCode,
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
