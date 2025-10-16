package handlers

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"puzzle/models"
	"puzzle/templates"
	"strings"
	"time"
)

// GameHandler handles the main game page
type GameHandler struct {
	Config *models.Config
	DB     *models.DB
}

// NewGameHandler creates a new game handler
func NewGameHandler(config *models.Config, db *models.DB) *GameHandler {
	return &GameHandler{
		Config: config,
		DB:     db,
	}
}

// ServeHTTP handles the game page request (old backward-compatible route)
func (h *GameHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// This is for backward compatibility - uses config.yaml settings
	lang := r.URL.Query().Get("lang")
	if lang == "" {
		lang = "en"
	}

	translations, err := models.LoadTranslations(lang)
	if err != nil {
		translations, _ = models.LoadTranslations("en")
		lang = "en"
	}

	imageURL, err := h.getRandomImage()
	if err != nil {
		http.Error(w, translations.Error.NoImages, http.StatusInternalServerError)
		return
	}

	component := templates.Game(
		h.Config.Puzzle.GridSize,
		imageURL,
		h.Config.Puzzle.TimeLimit,
		h.Config.Puzzle.TimerMode,
		h.Config.Puzzle.CountdownTime,
		h.Config.Puzzle.TestingMode,
		h.Config.Puzzle.ScrambleMoves,
		h.Config.Puzzle.AutoSolveSpeed,
		translations,
		lang,
		"/complete", // Old route uses /complete
		"#667eea",   // Default tile color
		"#5568d3",   // Default tile hover color
		"linear-gradient(135deg, #667eea 0%, #764ba2 100%)", // Default background
		"#333333", // Default text color
		"#ffffff", // Default button text color
		"#333333", // Default border color
	)

	if err := component.Render(r.Context(), w); err != nil {
		http.Error(w, fmt.Sprintf("Failed to render template: %v", err), http.StatusInternalServerError)
		return
	}
}

// ServeGameForPuzzle serves the game page for a specific puzzle from database
func (h *GameHandler) ServeGameForPuzzle(w http.ResponseWriter, r *http.Request) {
	// Extract company and puzzle slugs from URL path
	// Expected format: /play/:companySlug/:puzzleSlug
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 3 {
		http.Error(w, "Invalid URL format", http.StatusBadRequest)
		return
	}

	companySlug := pathParts[1]
	puzzleSlug := pathParts[2]

	// Get puzzle from database
	puzzle, err := h.DB.GetPuzzleBySlug(companySlug, puzzleSlug)
	if err != nil {
		log.Printf("Error loading puzzle %s/%s: %v", companySlug, puzzleSlug, err)
		http.Error(w, "Failed to load puzzle", http.StatusInternalServerError)
		return
	}

	if puzzle == nil {
		http.Error(w, "Puzzle not found", http.StatusNotFound)
		return
	}

	if !puzzle.IsActive {
		http.Error(w, "This puzzle is not currently active", http.StatusForbidden)
		return
	}

	// Get language
	lang := r.URL.Query().Get("lang")
	if lang == "" {
		lang = "en"
	}

	translations, err := models.LoadTranslations(lang)
	if err != nil {
		translations, _ = models.LoadTranslations("en")
		lang = "en"
	}

	// Build image URL from puzzle data
	imageURL := "/images" + puzzle.ImagePath

	// Build completion URL for this puzzle
	completionURL := fmt.Sprintf("/complete/%s/%s", companySlug, puzzleSlug)

	// Render the game template with puzzle settings
	component := templates.Game(
		puzzle.GridSize,
		imageURL,
		puzzle.TimeLimit,
		puzzle.TimerMode,
		puzzle.CountdownTime,
		puzzle.TestingMode, // Use puzzle's testing mode setting
		puzzle.ScrambleMoves,
		puzzle.AutoSolveSpeed,
		translations,
		lang,
		completionURL,
		puzzle.TileColor,
		puzzle.TileHoverColor,
		puzzle.BackgroundColor,
		puzzle.TextColor,
		puzzle.ButtonTextColor,
		puzzle.BorderColor,
	)

	if err := component.Render(r.Context(), w); err != nil {
		http.Error(w, fmt.Sprintf("Failed to render template: %v", err), http.StatusInternalServerError)
		return
	}
}

// getRandomImage selects a random PNG image from the images directory
func (h *GameHandler) getRandomImage() (string, error) {
	files, err := os.ReadDir(h.Config.Images.Directory)
	if err != nil {
		return "", fmt.Errorf("failed to read images directory: %w", err)
	}

	// Filter for PNG files
	var pngFiles []string
	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".png" {
			pngFiles = append(pngFiles, file.Name())
		}
	}

	if len(pngFiles) == 0 {
		return "", fmt.Errorf("no PNG images found in %s", h.Config.Images.Directory)
	}

	// Select random image
	rand.Seed(time.Now().UnixNano())
	selectedImage := pngFiles[rand.Intn(len(pngFiles))]

	return "/images/" + selectedImage, nil
}
