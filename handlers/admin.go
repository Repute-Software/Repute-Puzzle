package handlers

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"puzzle/middleware"
	"puzzle/models"
	"puzzle/templates"
	"strconv"
	"strings"
)

// AdminHandler handles admin dashboard routes
type AdminHandler struct {
	DB     *models.DB
	Config *models.Config
}

// NewAdminHandler creates a new admin handler
func NewAdminHandler(db *models.DB, config *models.Config) *AdminHandler {
	return &AdminHandler{
		DB:     db,
		Config: config,
	}
}

// ShowDashboard displays the puzzle list
func (h *AdminHandler) ShowDashboard(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r)
	company := middleware.GetCompany(r)

	// Get all puzzles for this company
	puzzles, err := h.DB.GetPuzzlesByCompany(company.ID)
	if err != nil {
		log.Printf("Error getting puzzles: %v", err)
		http.Error(w, "Failed to load puzzles", http.StatusInternalServerError)
		return
	}

	component := templates.AdminDashboard(company.Name, user.Email, puzzles)
	if err := component.Render(r.Context(), w); err != nil {
		http.Error(w, "Failed to render dashboard", http.StatusInternalServerError)
	}
}

// ShowCreatePuzzle displays the create puzzle form
func (h *AdminHandler) ShowCreatePuzzle(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r)
	company := middleware.GetCompany(r)
	errorMsg := r.URL.Query().Get("error")

	component := templates.AdminPuzzleFormCreate(company.Name, user.Email, errorMsg)
	if err := component.Render(r.Context(), w); err != nil {
		http.Error(w, "Failed to render form", http.StatusInternalServerError)
	}
}

// HandleCreatePuzzle processes the create puzzle form
func (h *AdminHandler) HandleCreatePuzzle(w http.ResponseWriter, r *http.Request) {
	company := middleware.GetCompany(r)

	// Parse multipart form (for image upload)
	if err := r.ParseMultipartForm(10 << 20); err != nil { // 10 MB max
		http.Redirect(w, r, "/admin/puzzles/new?error=Failed+to+parse+form", http.StatusSeeOther)
		return
	}

	// Extract form values
	name := strings.TrimSpace(r.FormValue("name"))
	slug := strings.TrimSpace(r.FormValue("slug"))
	gridSize, _ := strconv.Atoi(r.FormValue("grid_size"))
	discountPercent, _ := strconv.Atoi(r.FormValue("discount_percent"))
	timeLimit, _ := strconv.Atoi(r.FormValue("time_limit"))
	timerMode := r.FormValue("timer_mode")
	countdownTime, _ := strconv.Atoi(r.FormValue("countdown_time"))
	scrambleMoves, _ := strconv.Atoi(r.FormValue("scramble_moves"))
	autoSolveSpeed, _ := strconv.Atoi(r.FormValue("auto_solve_speed"))
	testingMode := r.FormValue("testing_mode") == "true"
	isActive := r.FormValue("is_active") == "true"

	// Validate required fields
	if name == "" || slug == "" {
		http.Redirect(w, r, "/admin/puzzles/new?error=Name+and+slug+required", http.StatusSeeOther)
		return
	}

	// Check if slug already exists for this company
	existing, err := h.DB.GetPuzzleBySlug(company.Slug, slug)
	if err != nil {
		log.Printf("Error checking puzzle slug: %v", err)
		http.Redirect(w, r, "/admin/puzzles/new?error=Database+error", http.StatusSeeOther)
		return
	}
	if existing != nil {
		http.Redirect(w, r, "/admin/puzzles/new?error=Slug+already+exists", http.StatusSeeOther)
		return
	}

	// Handle image upload
	file, header, err := r.FormFile("image")
	if err != nil {
		http.Redirect(w, r, "/admin/puzzles/new?error=Image+required", http.StatusSeeOther)
		return
	}
	defer file.Close()

	// Validate image type
	if !strings.HasSuffix(strings.ToLower(header.Filename), ".png") &&
		!strings.HasSuffix(strings.ToLower(header.Filename), ".jpg") &&
		!strings.HasSuffix(strings.ToLower(header.Filename), ".jpeg") {
		http.Redirect(w, r, "/admin/puzzles/new?error=Only+PNG+and+JPG+images+allowed", http.StatusSeeOther)
		return
	}

	// Create puzzle first to get ID
	puzzle := &models.Puzzle{
		CompanyID:       company.ID,
		Name:            name,
		Slug:            slug,
		ImagePath:       "", // Will update after saving image
		GridSize:        gridSize,
		DiscountPercent: discountPercent,
		TimeLimit:       timeLimit,
		TimerMode:       timerMode,
		CountdownTime:   countdownTime,
		ScrambleMoves:   scrambleMoves,
		AutoSolveSpeed:  autoSolveSpeed,
		TestingMode:     testingMode,
		IsActive:        isActive,
	}

	// Temporarily set image path
	puzzle.ImagePath = "temp"
	createdPuzzle, err := h.DB.CreatePuzzle(puzzle)
	if err != nil {
		log.Printf("Error creating puzzle: %v", err)
		http.Redirect(w, r, "/admin/puzzles/new?error=Failed+to+create+puzzle", http.StatusSeeOther)
		return
	}

	// Save image with puzzle ID
	imagePath, err := h.saveImage(file, company.ID, createdPuzzle.ID, header.Filename)
	if err != nil {
		log.Printf("Error saving image: %v", err)
		// Delete the puzzle since image upload failed
		h.DB.DeletePuzzle(createdPuzzle.ID, company.ID)
		http.Redirect(w, r, "/admin/puzzles/new?error=Failed+to+save+image", http.StatusSeeOther)
		return
	}

	// Update puzzle with image path
	createdPuzzle.ImagePath = imagePath
	if err := h.DB.UpdatePuzzle(createdPuzzle); err != nil {
		log.Printf("Error updating puzzle image path: %v", err)
	}

	log.Printf("Puzzle created: %s (ID: %d) by company %s", name, createdPuzzle.ID, company.Name)
	http.Redirect(w, r, fmt.Sprintf("/admin/puzzles/%d", createdPuzzle.ID), http.StatusSeeOther)
}

// ShowPuzzleDetail displays puzzle details with embed code
func (h *AdminHandler) ShowPuzzleDetail(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r)
	company := middleware.GetCompany(r)

	// Extract puzzle ID from URL
	puzzleID, err := strconv.Atoi(strings.TrimPrefix(r.URL.Path, "/admin/puzzles/"))
	if err != nil {
		http.Error(w, "Invalid puzzle ID", http.StatusBadRequest)
		return
	}

	// Get puzzle
	puzzle, err := h.DB.GetPuzzleByID(puzzleID)
	if err != nil {
		log.Printf("Error getting puzzle: %v", err)
		http.Error(w, "Failed to load puzzle", http.StatusInternalServerError)
		return
	}

	if puzzle == nil || puzzle.CompanyID != company.ID {
		http.Error(w, "Puzzle not found", http.StatusNotFound)
		return
	}

	// Get statistics
	stats, err := h.DB.GetPuzzleCompletionStats(puzzleID)
	if err != nil {
		log.Printf("Error getting puzzle stats: %v", err)
		stats = map[string]interface{}{
			"total_completions": 0,
			"avg_moves":         0.0,
			"avg_time":          0.0,
		}
	}

	component := templates.AdminPuzzleDetail(company.Name, user.Email, puzzle, company, stats)
	if err := component.Render(r.Context(), w); err != nil {
		http.Error(w, "Failed to render puzzle detail", http.StatusInternalServerError)
	}
}

// ShowEditPuzzle displays the edit puzzle form
func (h *AdminHandler) ShowEditPuzzle(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r)
	company := middleware.GetCompany(r)

	// Extract puzzle ID
	puzzleIDStr := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/admin/puzzles/"), "/edit")
	puzzleID, err := strconv.Atoi(puzzleIDStr)
	if err != nil {
		http.Error(w, "Invalid puzzle ID", http.StatusBadRequest)
		return
	}

	// Get puzzle
	puzzle, err := h.DB.GetPuzzleByID(puzzleID)
	if err != nil {
		log.Printf("Error getting puzzle: %v", err)
		http.Error(w, "Failed to load puzzle", http.StatusInternalServerError)
		return
	}

	if puzzle == nil || puzzle.CompanyID != company.ID {
		http.Error(w, "Puzzle not found", http.StatusNotFound)
		return
	}

	errorMsg := r.URL.Query().Get("error")
	component := templates.AdminPuzzleFormEdit(company.Name, user.Email, puzzle, errorMsg)
	if err := component.Render(r.Context(), w); err != nil {
		http.Error(w, "Failed to render form", http.StatusInternalServerError)
	}
}

// HandleUpdatePuzzle processes the edit puzzle form
func (h *AdminHandler) HandleUpdatePuzzle(w http.ResponseWriter, r *http.Request) {
	company := middleware.GetCompany(r)

	// Extract puzzle ID
	puzzleIDStr := strings.TrimPrefix(r.URL.Path, "/admin/puzzles/")
	puzzleID, err := strconv.Atoi(puzzleIDStr)
	if err != nil {
		http.Error(w, "Invalid puzzle ID", http.StatusBadRequest)
		return
	}

	// Get existing puzzle
	puzzle, err := h.DB.GetPuzzleByID(puzzleID)
	if err != nil || puzzle == nil || puzzle.CompanyID != company.ID {
		http.Error(w, "Puzzle not found", http.StatusNotFound)
		return
	}

	// Parse form
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Redirect(w, r, fmt.Sprintf("/admin/puzzles/%d/edit?error=Failed+to+parse+form", puzzleID), http.StatusSeeOther)
		return
	}

	// Update fields
	puzzle.Name = strings.TrimSpace(r.FormValue("name"))
	puzzle.Slug = strings.TrimSpace(r.FormValue("slug"))
	puzzle.GridSize, _ = strconv.Atoi(r.FormValue("grid_size"))
	puzzle.DiscountPercent, _ = strconv.Atoi(r.FormValue("discount_percent"))
	puzzle.TimeLimit, _ = strconv.Atoi(r.FormValue("time_limit"))
	puzzle.TimerMode = r.FormValue("timer_mode")
	puzzle.CountdownTime, _ = strconv.Atoi(r.FormValue("countdown_time"))
	puzzle.ScrambleMoves, _ = strconv.Atoi(r.FormValue("scramble_moves"))
	puzzle.AutoSolveSpeed, _ = strconv.Atoi(r.FormValue("auto_solve_speed"))
	puzzle.TestingMode = r.FormValue("testing_mode") == "true"
	puzzle.IsActive = r.FormValue("is_active") == "true"

	// Handle image upload if provided
	file, header, err := r.FormFile("image")
	if err == nil {
		defer file.Close()

		// Validate image type
		if strings.HasSuffix(strings.ToLower(header.Filename), ".png") ||
			strings.HasSuffix(strings.ToLower(header.Filename), ".jpg") ||
			strings.HasSuffix(strings.ToLower(header.Filename), ".jpeg") {

			// Save new image
			imagePath, err := h.saveImage(file, company.ID, puzzle.ID, header.Filename)
			if err != nil {
				log.Printf("Error saving new image: %v", err)
			} else {
				// Delete old image
				if puzzle.ImagePath != "" {
					oldPath := filepath.Join(h.Config.Images.Directory, puzzle.ImagePath)
					os.Remove(oldPath)
				}
				puzzle.ImagePath = imagePath
			}
		}
	}

	// Update puzzle in database
	if err := h.DB.UpdatePuzzle(puzzle); err != nil {
		log.Printf("Error updating puzzle: %v", err)
		http.Redirect(w, r, fmt.Sprintf("/admin/puzzles/%d/edit?error=Failed+to+update+puzzle", puzzleID), http.StatusSeeOther)
		return
	}

	log.Printf("Puzzle updated: %s (ID: %d)", puzzle.Name, puzzle.ID)
	http.Redirect(w, r, fmt.Sprintf("/admin/puzzles/%d", puzzleID), http.StatusSeeOther)
}

// HandleDeletePuzzle deletes a puzzle
func (h *AdminHandler) HandleDeletePuzzle(w http.ResponseWriter, r *http.Request) {
	company := middleware.GetCompany(r)

	// Extract puzzle ID
	puzzleIDStr := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/admin/puzzles/"), "/delete")
	puzzleID, err := strconv.Atoi(puzzleIDStr)
	if err != nil {
		http.Error(w, "Invalid puzzle ID", http.StatusBadRequest)
		return
	}

	// Get puzzle to verify ownership and get image path
	puzzle, err := h.DB.GetPuzzleByID(puzzleID)
	if err != nil || puzzle == nil || puzzle.CompanyID != company.ID {
		http.Error(w, "Puzzle not found", http.StatusNotFound)
		return
	}

	// Delete image file
	if puzzle.ImagePath != "" {
		imagePath := filepath.Join(h.Config.Images.Directory, puzzle.ImagePath)
		if err := os.Remove(imagePath); err != nil {
			log.Printf("Warning: Failed to delete image file: %v", err)
		}
	}

	// Delete puzzle from database
	if err := h.DB.DeletePuzzle(puzzleID, company.ID); err != nil {
		log.Printf("Error deleting puzzle: %v", err)
		http.Error(w, "Failed to delete puzzle", http.StatusInternalServerError)
		return
	}

	log.Printf("Puzzle deleted: ID %d by company %s", puzzleID, company.Name)
	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

// saveImage saves an uploaded image file
func (h *AdminHandler) saveImage(file io.Reader, companyID, puzzleID int, filename string) (string, error) {
	// Create company directory
	companyDir := filepath.Join(h.Config.Images.Directory, fmt.Sprintf("%d", companyID))
	if err := os.MkdirAll(companyDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create company directory: %w", err)
	}

	// Get file extension
	ext := filepath.Ext(filename)
	if ext == "" {
		ext = ".png"
	}

	// Create image filename
	imageFilename := fmt.Sprintf("%d%s", puzzleID, ext)
	imagePath := filepath.Join(companyDir, imageFilename)

	// Create file
	dst, err := os.Create(imagePath)
	if err != nil {
		return "", fmt.Errorf("failed to create image file: %w", err)
	}
	defer dst.Close()

	// Copy image data
	if _, err := io.Copy(dst, file); err != nil {
		return "", fmt.Errorf("failed to save image: %w", err)
	}

	// Return relative path for database
	relativePath := fmt.Sprintf("/%d/%s", companyID, imageFilename)
	return relativePath, nil
}

