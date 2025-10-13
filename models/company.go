package models

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

// Company represents a tenant organization
type Company struct {
	ID             int       `json:"id"`
	Name           string    `json:"name"`
	Slug           string    `json:"slug"`
	EmailAPIKey    string    `json:"email_api_key,omitempty"`
	EmailFromEmail string    `json:"email_from_email,omitempty"`
	EmailFromName  string    `json:"email_from_name,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	IsActive       bool      `json:"is_active"`
}

// CreateCompany creates a new company
func (db *DB) CreateCompany(name, slug string) (*Company, error) {
	query := `
		INSERT INTO companies (name, slug, is_active)
		VALUES (?, ?, 1)
	`

	result, err := db.Exec(query, name, slug)
	if err != nil {
		return nil, fmt.Errorf("failed to create company: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get company ID: %w", err)
	}

	return db.GetCompanyByID(int(id))
}

// GetCompanyByID retrieves a company by ID
func (db *DB) GetCompanyByID(id int) (*Company, error) {
	query := `
		SELECT id, name, slug, email_api_key, email_from_email, email_from_name, created_at, is_active
		FROM companies
		WHERE id = ?
	`

	var c Company
	var emailAPIKey, emailFromEmail, emailFromName sql.NullString
	err := db.QueryRow(query, id).Scan(
		&c.ID,
		&c.Name,
		&c.Slug,
		&emailAPIKey,
		&emailFromEmail,
		&emailFromName,
		&c.CreatedAt,
		&c.IsActive,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get company: %w", err)
	}

	// Handle NULL values
	if emailAPIKey.Valid {
		c.EmailAPIKey = emailAPIKey.String
	}
	if emailFromEmail.Valid {
		c.EmailFromEmail = emailFromEmail.String
	}
	if emailFromName.Valid {
		c.EmailFromName = emailFromName.String
	}

	if err == sql.ErrNoRows {
		return nil, nil
	}

	return &c, nil
}

// GetCompanyBySlug retrieves a company by slug
func (db *DB) GetCompanyBySlug(slug string) (*Company, error) {
	query := `
		SELECT id, name, slug, email_api_key, email_from_email, email_from_name, created_at, is_active
		FROM companies
		WHERE slug = ?
	`

	var c Company
	var emailAPIKey, emailFromEmail, emailFromName sql.NullString
	err := db.QueryRow(query, slug).Scan(
		&c.ID,
		&c.Name,
		&c.Slug,
		&emailAPIKey,
		&emailFromEmail,
		&emailFromName,
		&c.CreatedAt,
		&c.IsActive,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get company: %w", err)
	}

	// Handle NULL values
	if emailAPIKey.Valid {
		c.EmailAPIKey = emailAPIKey.String
	}
	if emailFromEmail.Valid {
		c.EmailFromEmail = emailFromEmail.String
	}
	if emailFromName.Valid {
		c.EmailFromName = emailFromName.String
	}

	return &c, nil
}

// GetAllCompanies retrieves all companies
func (db *DB) GetAllCompanies() ([]Company, error) {
	query := `
		SELECT id, name, slug, created_at, is_active
		FROM companies
		ORDER BY created_at DESC
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query companies: %w", err)
	}
	defer rows.Close()

	var companies []Company
	for rows.Next() {
		var c Company
		if err := rows.Scan(
			&c.ID,
			&c.Name,
			&c.Slug,
			&c.CreatedAt,
			&c.IsActive,
		); err != nil {
			return nil, fmt.Errorf("failed to scan company: %w", err)
		}
		companies = append(companies, c)
	}

	return companies, nil
}

// UpdateCompany updates company details
func (db *DB) UpdateCompany(id int, name, slug string) error {
	query := `
		UPDATE companies
		SET name = ?, slug = ?
		WHERE id = ?
	`

	_, err := db.Exec(query, name, slug, id)
	if err != nil {
		return fmt.Errorf("failed to update company: %w", err)
	}

	return nil
}

// UpdateCompanyEmailSettings updates company email settings
func (db *DB) UpdateCompanyEmailSettings(id int, apiKey, fromEmail, fromName string) error {
	query := `
		UPDATE companies
		SET email_api_key = ?, email_from_email = ?, email_from_name = ?
		WHERE id = ?
	`

	_, err := db.Exec(query, apiKey, fromEmail, fromName, id)
	if err != nil {
		return fmt.Errorf("failed to update company email settings: %w", err)
	}

	return nil
}

// GenerateSlug generates a URL-friendly slug from company name
func GenerateSlug(name string) string {
	// Convert to lowercase
	slug := strings.ToLower(name)

	// Replace spaces with hyphens
	slug = strings.ReplaceAll(slug, " ", "-")

	// Remove special characters (keep only alphanumeric and hyphens)
	var result []rune
	for _, r := range slug {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			result = append(result, r)
		}
	}

	slug = string(result)

	// Remove consecutive hyphens
	for strings.Contains(slug, "--") {
		slug = strings.ReplaceAll(slug, "--", "-")
	}

	// Trim hyphens from start and end
	slug = strings.Trim(slug, "-")

	return slug
}
