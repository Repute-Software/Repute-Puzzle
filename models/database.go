package models

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// DB is the database instance
type DB struct {
	*sql.DB
}

// InitDB initializes the SQLite database and creates tables
func InitDB(dbPath string) (*DB, error) {
	// Ensure the directory exists
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	// Open database connection with proper settings for SQLite
	// Add WAL mode and other pragmas for better concurrency
	connStr := fmt.Sprintf("%s?cache=shared&mode=rwc&_journal_mode=WAL", dbPath)
	db, err := sql.Open("sqlite3", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool
	// SQLite has limited concurrency, so keep these numbers low
	db.SetMaxOpenConns(10)                 // Maximum 10 open connections
	db.SetMaxIdleConns(5)                  // Keep 5 idle connections
	db.SetConnMaxLifetime(0)               // Connections don't expire (SQLite is local)
	db.SetConnMaxIdleTime(5 * time.Minute) // Close idle connections after 5 minutes

	// Test the connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Create tables
	if err := createTables(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	return &DB{db}, nil
}

// createTables creates the necessary database tables
func createTables(db *sql.DB) error {
	schema := `
	-- Companies (tenants)
	CREATE TABLE IF NOT EXISTS companies (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		slug TEXT UNIQUE NOT NULL,
		email_api_key TEXT,
		email_from_email TEXT,
		email_from_name TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		is_active BOOLEAN DEFAULT 1
	);
	
	-- Users (company admins/members)
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		company_id INTEGER NOT NULL,
		email TEXT NOT NULL UNIQUE,
		password_hash TEXT NOT NULL,
		role TEXT DEFAULT 'admin',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (company_id) REFERENCES companies(id)
	);
	
	CREATE INDEX IF NOT EXISTS idx_users_company ON users(company_id);
	CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
	
	-- Sessions (authentication)
	CREATE TABLE IF NOT EXISTS sessions (
		id TEXT PRIMARY KEY,
		user_id INTEGER NOT NULL,
		expires_at DATETIME NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id)
	);
	
	CREATE INDEX IF NOT EXISTS idx_sessions_user ON sessions(user_id);
	CREATE INDEX IF NOT EXISTS idx_sessions_expires ON sessions(expires_at);
	
	-- Puzzles (each company can have multiple)
	CREATE TABLE IF NOT EXISTS puzzles (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		company_id INTEGER NOT NULL,
		name TEXT NOT NULL,
		slug TEXT NOT NULL,
		image_path TEXT NOT NULL,
		grid_size INTEGER DEFAULT 3,
		discount_percent INTEGER DEFAULT 15,
		time_limit INTEGER DEFAULT 90,
		timer_mode TEXT DEFAULT 'first_move',
		countdown_time INTEGER DEFAULT 15,
		scramble_moves INTEGER DEFAULT 25,
		auto_solve_speed INTEGER DEFAULT 50,
		testing_mode BOOLEAN DEFAULT 0,
		code_expiration_days INTEGER DEFAULT 0,
		is_active BOOLEAN DEFAULT 1,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (company_id) REFERENCES companies(id),
		UNIQUE(company_id, slug)
	);
	
	CREATE INDEX IF NOT EXISTS idx_puzzles_company ON puzzles(company_id);
	CREATE INDEX IF NOT EXISTS idx_puzzles_slug ON puzzles(company_id, slug);
	CREATE INDEX IF NOT EXISTS idx_puzzles_active ON puzzles(is_active);
	
	-- Completions (updated to reference puzzle)
	CREATE TABLE IF NOT EXISTS completions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		puzzle_id INTEGER NOT NULL,
		email TEXT NOT NULL,
		discount_code TEXT NOT NULL UNIQUE,
		moves INTEGER NOT NULL,
		time_seconds INTEGER NOT NULL,
		expires_at DATETIME,
		is_used BOOLEAN DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (puzzle_id) REFERENCES puzzles(id)
	);
	
	CREATE INDEX IF NOT EXISTS idx_completions_puzzle ON completions(puzzle_id);
	CREATE INDEX IF NOT EXISTS idx_completions_email ON completions(puzzle_id, email);
	CREATE INDEX IF NOT EXISTS idx_completions_code ON completions(discount_code);
	CREATE INDEX IF NOT EXISTS idx_completions_created ON completions(created_at);
	CREATE INDEX IF NOT EXISTS idx_completions_expires ON completions(expires_at);
	`

	_, err := db.Exec(schema)
	return err
}
