package handlers

import (
	"fmt"
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
}

// NewEmbedHandler creates a new embed handler
func NewEmbedHandler(config *models.Config) *EmbedHandler {
	return &EmbedHandler{Config: config}
}

// ServeHTTP handles the embed page request
func (h *EmbedHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Detect language from URL parameter
	lang := r.URL.Query().Get("lang")
	if lang == "" {
		lang = "en" // Default to English
	}

	// Load translations
	translations, err := models.LoadTranslations(lang)
	if err != nil {
		// Fallback to English on error
		translations, _ = models.LoadTranslations("en")
		lang = "en"
	}

	// Get primary color from URL parameter
	primaryColor := r.URL.Query().Get("primaryColor")
	if primaryColor == "" {
		primaryColor = "667eea" // Default color
	}

	// Ensure it doesn't have # prefix
	primaryColor = strings.TrimPrefix(primaryColor, "#")

	// Validate hex color format (6 characters)
	if !isValidHexColor(primaryColor) {
		primaryColor = "667eea" // Fallback to default
	}

	// Add # prefix for CSS
	primaryColor = "#" + primaryColor

	// Select a random image from the images directory
	imageURL, err := h.getRandomImage()
	if err != nil {
		http.Error(w, translations.Error.NoImages, http.StatusInternalServerError)
		return
	}

	// Render the embed template
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
