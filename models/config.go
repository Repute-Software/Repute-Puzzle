package models

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config holds all application configuration
type Config struct {
	Puzzle   PuzzleConfig   `yaml:"puzzle"`
	Email    EmailConfig    `yaml:"email"`
	Database DatabaseConfig `yaml:"database"`
	Server   ServerConfig   `yaml:"server"`
	Images   ImagesConfig   `yaml:"images"`
}

// PuzzleConfig contains game settings
type PuzzleConfig struct {
	GridSize        int  `yaml:"grid_size"`
	DiscountPercent int  `yaml:"discount_percent"`
	TimeLimit       int  `yaml:"time_limit"`       // in seconds, 0 = no limit
	TestingMode     bool `yaml:"testing_mode"`     // highlight clickable tiles for testing
	ScrambleMoves   int  `yaml:"scramble_moves"`   // number of moves to scramble from solved state
	AutoSolveSpeed  int  `yaml:"auto_solve_speed"` // milliseconds per move for auto-solve
}

// DatabaseConfig contains database settings
type DatabaseConfig struct {
	Path string `yaml:"path"`
}

// ServerConfig contains HTTP server settings
type ServerConfig struct {
	Port int `yaml:"port"`
}

// ImagesConfig contains image directory settings
type ImagesConfig struct {
	Directory string `yaml:"directory"`
}

// EmailConfig contains email service settings
type EmailConfig struct {
	Enabled   bool   `yaml:"enabled"`
	Provider  string `yaml:"provider"`
	APIKey    string `yaml:"api_key"`
	FromEmail string `yaml:"from_email"`
	FromName  string `yaml:"from_name"`
	SubjectEN string `yaml:"subject_en"`
	SubjectNL string `yaml:"subject_nl"`
}

// LoadConfig reads and parses the config.yaml file
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &config, nil
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.Puzzle.GridSize < 2 || c.Puzzle.GridSize > 10 {
		return fmt.Errorf("grid_size must be between 2 and 10, got %d", c.Puzzle.GridSize)
	}
	if c.Puzzle.DiscountPercent < 0 || c.Puzzle.DiscountPercent > 100 {
		return fmt.Errorf("discount_percent must be between 0 and 100, got %d", c.Puzzle.DiscountPercent)
	}
	if c.Puzzle.TimeLimit < 0 {
		return fmt.Errorf("time_limit must be >= 0, got %d", c.Puzzle.TimeLimit)
	}
	if c.Puzzle.ScrambleMoves < 1 || c.Puzzle.ScrambleMoves > 1000 {
		return fmt.Errorf("scramble_moves must be between 1 and 1000, got %d", c.Puzzle.ScrambleMoves)
	}
	if c.Puzzle.AutoSolveSpeed < 1 || c.Puzzle.AutoSolveSpeed > 10000 {
		return fmt.Errorf("auto_solve_speed must be between 1 and 10000 ms, got %d", c.Puzzle.AutoSolveSpeed)
	}
	if c.Server.Port < 1 || c.Server.Port > 65535 {
		return fmt.Errorf("server port must be between 1 and 65535, got %d", c.Server.Port)
	}
	if c.Images.Directory == "" {
		return fmt.Errorf("images directory cannot be empty")
	}

	// Email validation
	if c.Email.Enabled {
		if c.Email.FromEmail == "" {
			return fmt.Errorf("email.from_email is required")
		}
		// Note: API key can be set via config.yaml or RESEND_API_KEY env var
		// Validation happens at runtime in main.go
	}

	return nil
}
