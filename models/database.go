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
	db.SetMaxOpenConns(10)                      // Maximum 10 open connections
	db.SetMaxIdleConns(5)                       // Keep 5 idle connections
	db.SetConnMaxLifetime(0)                    // Connections don't expire (SQLite is local)
	db.SetConnMaxIdleTime(5 * time.Minute)     // Close idle connections after 5 minutes

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
	CREATE TABLE IF NOT EXISTS completions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		email TEXT NOT NULL,
		discount_code TEXT NOT NULL UNIQUE,
		grid_size INTEGER NOT NULL,
		moves INTEGER NOT NULL,
		time_seconds INTEGER NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	
	CREATE INDEX IF NOT EXISTS idx_email ON completions(email);
	CREATE INDEX IF NOT EXISTS idx_discount_code ON completions(discount_code);
	CREATE INDEX IF NOT EXISTS idx_created_at ON completions(created_at);
	`

	_, err := db.Exec(schema)
	return err
}
