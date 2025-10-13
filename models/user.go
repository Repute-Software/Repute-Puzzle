package models

import (
	"database/sql"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// User represents a user account
type User struct {
	ID           int       `json:"id"`
	CompanyID    int       `json:"company_id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"` // Never expose in JSON
	Role         string    `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
}

// CreateUser creates a new user with hashed password
func (db *DB) CreateUser(companyID int, email, password, role string) (*User, error) {
	// Hash the password
	hashedPassword, err := HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	query := `
		INSERT INTO users (company_id, email, password_hash, role)
		VALUES (?, ?, ?, ?)
	`

	result, err := db.Exec(query, companyID, email, hashedPassword, role)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get user ID: %w", err)
	}

	return db.GetUserByID(int(id))
}

// GetUserByID retrieves a user by ID
func (db *DB) GetUserByID(id int) (*User, error) {
	query := `
		SELECT id, company_id, email, password_hash, role, created_at
		FROM users
		WHERE id = ?
	`

	var u User
	err := db.QueryRow(query, id).Scan(
		&u.ID,
		&u.CompanyID,
		&u.Email,
		&u.PasswordHash,
		&u.Role,
		&u.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &u, nil
}

// GetUserByEmail retrieves a user by email
func (db *DB) GetUserByEmail(email string) (*User, error) {
	query := `
		SELECT id, company_id, email, password_hash, role, created_at
		FROM users
		WHERE email = ?
	`

	var u User
	err := db.QueryRow(query, email).Scan(
		&u.ID,
		&u.CompanyID,
		&u.Email,
		&u.PasswordHash,
		&u.Role,
		&u.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &u, nil
}

// GetUsersByCompany retrieves all users for a company
func (db *DB) GetUsersByCompany(companyID int) ([]User, error) {
	query := `
		SELECT id, company_id, email, password_hash, role, created_at
		FROM users
		WHERE company_id = ?
		ORDER BY created_at DESC
	`

	rows, err := db.Query(query, companyID)
	if err != nil {
		return nil, fmt.Errorf("failed to query users: %w", err)
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var u User
		if err := rows.Scan(
			&u.ID,
			&u.CompanyID,
			&u.Email,
			&u.PasswordHash,
			&u.Role,
			&u.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, u)
	}

	return users, nil
}

// AuthenticateUser checks if email and password match
func (db *DB) AuthenticateUser(email, password string) (*User, error) {
	user, err := db.GetUserByEmail(email)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, nil // User not found
	}

	// Check password
	if !CheckPassword(password, user.PasswordHash) {
		return nil, nil // Invalid password
	}

	return user, nil
}

// HashPassword hashes a plain text password using bcrypt
func HashPassword(password string) (string, error) {
	// Cost of 12 is a good balance between security and performance
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}

// CheckPassword compares a plain text password with a hashed password
func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// ValidatePassword checks if password meets requirements
func ValidatePassword(password string) error {
	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters")
	}
	
	// Add more validation rules as needed
	hasLetter := false
	hasNumber := false
	
	for _, char := range password {
		if char >= 'a' && char <= 'z' || char >= 'A' && char <= 'Z' {
			hasLetter = true
		}
		if char >= '0' && char <= '9' {
			hasNumber = true
		}
	}
	
	if !hasLetter {
		return fmt.Errorf("password must contain at least one letter")
	}
	if !hasNumber {
		return fmt.Errorf("password must contain at least one number")
	}
	
	return nil
}

