package handlers

import (
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"puzzle/models"
	"puzzle/templates"
	"time"
)

// GameHandler handles the main game page
type GameHandler struct {
	Config *models.Config
}

// NewGameHandler creates a new game handler
func NewGameHandler(config *models.Config) *GameHandler {
	return &GameHandler{Config: config}
}

// ServeHTTP handles the game page request
func (h *GameHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

	// Select a random image from the images directory
	imageURL, err := h.getRandomImage()
	if err != nil {
		http.Error(w, translations.Error.NoImages, http.StatusInternalServerError)
		return
	}

	// Render the game template
	component := templates.Game(
		h.Config.Puzzle.GridSize,
		imageURL,
		h.Config.Puzzle.TimeLimit,
		h.Config.Puzzle.TestingMode,
		h.Config.Puzzle.ScrambleMoves,
		h.Config.Puzzle.AutoSolveSpeed,
		translations,
		lang,
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
