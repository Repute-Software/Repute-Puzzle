package models

import (
	"database/sql"
	"fmt"
	"time"
)

// Puzzle represents a puzzle configuration
type Puzzle struct {
	ID                 int       `json:"id"`
	CompanyID          int       `json:"company_id"`
	Name               string    `json:"name"`
	Slug               string    `json:"slug"`
	ImagePath          string    `json:"image_path"`
	GridSize           int       `json:"grid_size"`
	DiscountPercent    int       `json:"discount_percent"`
	TimeLimit          int       `json:"time_limit"`
	TimerMode          string    `json:"timer_mode"`
	CountdownTime      int       `json:"countdown_time"`
	ScrambleMoves      int       `json:"scramble_moves"`
	AutoSolveSpeed     int       `json:"auto_solve_speed"`
	TestingMode        bool      `json:"testing_mode"`
	CodeExpirationDays int       `json:"code_expiration_days"`
	IsActive           bool      `json:"is_active"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

// CreatePuzzle creates a new puzzle for a company
func (db *DB) CreatePuzzle(p *Puzzle) (*Puzzle, error) {
	query := `
		INSERT INTO puzzles (
			company_id, name, slug, image_path, grid_size, discount_percent,
			time_limit, timer_mode, countdown_time, scramble_moves, auto_solve_speed, testing_mode, code_expiration_days, is_active
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	result, err := db.Exec(query,
		p.CompanyID, p.Name, p.Slug, p.ImagePath, p.GridSize, p.DiscountPercent,
		p.TimeLimit, p.TimerMode, p.CountdownTime, p.ScrambleMoves, p.AutoSolveSpeed, p.TestingMode, p.CodeExpirationDays, p.IsActive,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create puzzle: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get puzzle ID: %w", err)
	}

	return db.GetPuzzleByID(int(id))
}

// GetPuzzleByID retrieves a puzzle by ID
func (db *DB) GetPuzzleByID(id int) (*Puzzle, error) {
	query := `
		SELECT id, company_id, name, slug, image_path, grid_size, discount_percent,
		       time_limit, timer_mode, countdown_time, scramble_moves, auto_solve_speed, testing_mode, code_expiration_days,
		       is_active, created_at, updated_at
		FROM puzzles
		WHERE id = ?
	`

	var p Puzzle
	err := db.QueryRow(query, id).Scan(
		&p.ID, &p.CompanyID, &p.Name, &p.Slug, &p.ImagePath, &p.GridSize, &p.DiscountPercent,
		&p.TimeLimit, &p.TimerMode, &p.CountdownTime, &p.ScrambleMoves, &p.AutoSolveSpeed, &p.TestingMode, &p.CodeExpirationDays,
		&p.IsActive, &p.CreatedAt, &p.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get puzzle: %w", err)
	}

	return &p, nil
}

// GetPuzzleBySlug retrieves a puzzle by company slug and puzzle slug
func (db *DB) GetPuzzleBySlug(companySlug, puzzleSlug string) (*Puzzle, error) {
	query := `
		SELECT p.id, p.company_id, p.name, p.slug, p.image_path, p.grid_size, p.discount_percent,
		       p.time_limit, p.timer_mode, p.countdown_time, p.scramble_moves, p.auto_solve_speed, p.testing_mode, p.code_expiration_days,
		       p.is_active, p.created_at, p.updated_at
		FROM puzzles p
		JOIN companies c ON c.id = p.company_id
		WHERE c.slug = ? AND p.slug = ? AND p.is_active = 1
	`

	var p Puzzle
	err := db.QueryRow(query, companySlug, puzzleSlug).Scan(
		&p.ID, &p.CompanyID, &p.Name, &p.Slug, &p.ImagePath, &p.GridSize, &p.DiscountPercent,
		&p.TimeLimit, &p.TimerMode, &p.CountdownTime, &p.ScrambleMoves, &p.AutoSolveSpeed, &p.TestingMode, &p.CodeExpirationDays,
		&p.IsActive, &p.CreatedAt, &p.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get puzzle: %w", err)
	}

	return &p, nil
}

// GetPuzzlesByCompany retrieves all puzzles for a company
func (db *DB) GetPuzzlesByCompany(companyID int) ([]Puzzle, error) {
	query := `
		SELECT id, company_id, name, slug, image_path, grid_size, discount_percent,
		       time_limit, timer_mode, countdown_time, scramble_moves, auto_solve_speed, testing_mode, code_expiration_days,
		       is_active, created_at, updated_at
		FROM puzzles
		WHERE company_id = ?
		ORDER BY created_at DESC
	`

	rows, err := db.Query(query, companyID)
	if err != nil {
		return nil, fmt.Errorf("failed to query puzzles: %w", err)
	}
	defer rows.Close()

	var puzzles []Puzzle
	for rows.Next() {
		var p Puzzle
		if err := rows.Scan(
			&p.ID, &p.CompanyID, &p.Name, &p.Slug, &p.ImagePath, &p.GridSize, &p.DiscountPercent,
			&p.TimeLimit, &p.TimerMode, &p.CountdownTime, &p.ScrambleMoves, &p.AutoSolveSpeed, &p.TestingMode, &p.CodeExpirationDays,
			&p.IsActive, &p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan puzzle: %w", err)
		}
		puzzles = append(puzzles, p)
	}

	return puzzles, nil
}

// UpdatePuzzle updates a puzzle's details
func (db *DB) UpdatePuzzle(p *Puzzle) error {
	query := `
		UPDATE puzzles
		SET name = ?, slug = ?, image_path = ?, grid_size = ?, discount_percent = ?,
		    time_limit = ?, timer_mode = ?, countdown_time = ?, scramble_moves = ?,
		    auto_solve_speed = ?, testing_mode = ?, code_expiration_days = ?, is_active = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ? AND company_id = ?
	`

	_, err := db.Exec(query,
		p.Name, p.Slug, p.ImagePath, p.GridSize, p.DiscountPercent,
		p.TimeLimit, p.TimerMode, p.CountdownTime, p.ScrambleMoves,
		p.AutoSolveSpeed, p.TestingMode, p.CodeExpirationDays, p.IsActive,
		p.ID, p.CompanyID,
	)
	if err != nil {
		return fmt.Errorf("failed to update puzzle: %w", err)
	}

	return nil
}

// DeletePuzzle deletes a puzzle (company-scoped)
func (db *DB) DeletePuzzle(id, companyID int) error {
	query := `DELETE FROM puzzles WHERE id = ? AND company_id = ?`

	_, err := db.Exec(query, id, companyID)
	if err != nil {
		return fmt.Errorf("failed to delete puzzle: %w", err)
	}

	return nil
}
