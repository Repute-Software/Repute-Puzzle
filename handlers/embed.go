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

// EmbedHandler handles the embed page
type EmbedHandler struct {
	Config *models.Config
	DB     *models.DB
}

// NewEmbedHandler creates a new embed handler
func NewEmbedHandler(config *models.Config, db *models.DB) *EmbedHandler {
	return &EmbedHandler{
		Config: config,
		DB:     db,
	}
}

// ServeHTTP handles the embed page request (old backward-compatible route)
func (h *EmbedHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

	primaryColor := r.URL.Query().Get("primaryColor")
	if primaryColor == "" {
		primaryColor = "667eea"
	}

	primaryColor = strings.TrimPrefix(primaryColor, "#")
	if !isValidHexColor(primaryColor) {
		primaryColor = "667eea"
	}
	primaryColor = "#" + primaryColor

	imageURL, err := h.getRandomImage()
	if err != nil {
		http.Error(w, translations.Error.NoImages, http.StatusInternalServerError)
		return
	}

	component := templates.Embed(
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
		primaryColor,
		"/complete", // Old route uses /complete
	)

	if err := component.Render(r.Context(), w); err != nil {
		http.Error(w, fmt.Sprintf("Failed to render template: %v", err), http.StatusInternalServerError)
		return
	}
}

// ServeEmbedForPuzzle serves the embed page for a specific puzzle from database
func (h *EmbedHandler) ServeEmbedForPuzzle(w http.ResponseWriter, r *http.Request) {
	// Extract company and puzzle slugs from URL path
	// Expected format: /embed/:companySlug/:puzzleSlug
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

	// Get primary color from URL parameter
	primaryColor := r.URL.Query().Get("primaryColor")
	if primaryColor == "" {
		primaryColor = "667eea"
	}

	primaryColor = strings.TrimPrefix(primaryColor, "#")
	if !isValidHexColor(primaryColor) {
		primaryColor = "667eea"
	}
	primaryColor = "#" + primaryColor

	// Build image URL from puzzle data
	imageURL := "/images" + puzzle.ImagePath

	// Build completion URL for this puzzle
	completionURL := fmt.Sprintf("/complete/%s/%s", companySlug, puzzleSlug)

	// Render the embed template with puzzle settings
	component := templates.Embed(
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
		primaryColor,
		completionURL,
	)

	if err := component.Render(r.Context(), w); err != nil {
		http.Error(w, fmt.Sprintf("Failed to render template: %v", err), http.StatusInternalServerError)
		return
	}
}

// getRandomImage selects a random PNG image from the images directory
func (h *EmbedHandler) getRandomImage() (string, error) {
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

// isValidHexColor validates a hex color string (without #)
func isValidHexColor(color string) bool {
	if len(color) != 6 {
		return false
	}
	for _, c := range color {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}
