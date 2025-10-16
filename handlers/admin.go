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
	"time"
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
	codeExpirationDays, _ := strconv.Atoi(r.FormValue("code_expiration_days"))
	isActive := r.FormValue("is_active") == "true"

	// Extract color values with defaults
	tileColor := strings.TrimSpace(r.FormValue("tile_color"))
	if tileColor == "" {
		tileColor = "#667eea"
	}
	tileHoverColor := strings.TrimSpace(r.FormValue("tile_hover_color"))
	if tileHoverColor == "" {
		tileHoverColor = "#5568d3"
	}
	backgroundColor := strings.TrimSpace(r.FormValue("background_color"))
	if backgroundColor == "" {
		backgroundColor = "linear-gradient(135deg, #667eea 0%, #764ba2 100%)"
	}
	textColor := strings.TrimSpace(r.FormValue("text_color"))
	if textColor == "" {
		textColor = "#333333"
	}
	buttonTextColor := strings.TrimSpace(r.FormValue("button_text_color"))
	if buttonTextColor == "" {
		buttonTextColor = "#ffffff"
	}
	borderColor := strings.TrimSpace(r.FormValue("border_color"))
	if borderColor == "" {
		borderColor = "#333333"
	}

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
		CompanyID:          company.ID,
		Name:               name,
		Slug:               slug,
		ImagePath:          "", // Will update after saving image
		GridSize:           gridSize,
		DiscountPercent:    discountPercent,
		TimeLimit:          timeLimit,
		TimerMode:          timerMode,
		CountdownTime:      countdownTime,
		ScrambleMoves:      scrambleMoves,
		AutoSolveSpeed:     autoSolveSpeed,
		TestingMode:        testingMode,
		CodeExpirationDays: codeExpirationDays,
		TileColor:          tileColor,
		TileHoverColor:     tileHoverColor,
		BackgroundColor:    backgroundColor,
		TextColor:          textColor,
		ButtonTextColor:    buttonTextColor,
		BorderColor:        borderColor,
		IsActive:           isActive,
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
	puzzle.CodeExpirationDays, _ = strconv.Atoi(r.FormValue("code_expiration_days"))
	puzzle.IsActive = r.FormValue("is_active") == "true"

	// Update color fields with defaults if empty
	tileColor := strings.TrimSpace(r.FormValue("tile_color"))
	if tileColor != "" {
		puzzle.TileColor = tileColor
	} else {
		puzzle.TileColor = "#667eea"
	}
	tileHoverColor := strings.TrimSpace(r.FormValue("tile_hover_color"))
	if tileHoverColor != "" {
		puzzle.TileHoverColor = tileHoverColor
	} else {
		puzzle.TileHoverColor = "#5568d3"
	}
	backgroundColor := strings.TrimSpace(r.FormValue("background_color"))
	if backgroundColor != "" {
		puzzle.BackgroundColor = backgroundColor
	} else {
		puzzle.BackgroundColor = "linear-gradient(135deg, #667eea 0%, #764ba2 100%)"
	}
	textColor := strings.TrimSpace(r.FormValue("text_color"))
	if textColor != "" {
		puzzle.TextColor = textColor
	} else {
		puzzle.TextColor = "#333333"
	}
	buttonTextColor := strings.TrimSpace(r.FormValue("button_text_color"))
	if buttonTextColor != "" {
		puzzle.ButtonTextColor = buttonTextColor
	} else {
		puzzle.ButtonTextColor = "#ffffff"
	}
	borderColor := strings.TrimSpace(r.FormValue("border_color"))
	if borderColor != "" {
		puzzle.BorderColor = borderColor
	} else {
		puzzle.BorderColor = "#333333"
	}

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

// ShowCompanySettings displays the company email settings page
func (h *AdminHandler) ShowCompanySettings(w http.ResponseWriter, r *http.Request) {
	company := middleware.GetCompany(r)
	user := middleware.GetUser(r)

	// Get fresh company data with email settings
	freshCompany, err := h.DB.GetCompanyByID(company.ID)
	if err != nil || freshCompany == nil {
		http.Error(w, "Failed to load company data", http.StatusInternalServerError)
		return
	}

	// Get message/error from query params
	message := r.URL.Query().Get("message")
	errorMsg := r.URL.Query().Get("error")

	component := templates.AdminCompanySettings(company.Name, user.Email, freshCompany, message, errorMsg)
	if err := component.Render(r.Context(), w); err != nil {
		http.Error(w, fmt.Sprintf("Failed to render template: %v", err), http.StatusInternalServerError)
		return
	}
}

// HandleUpdateCompanySettings processes the company settings form
func (h *AdminHandler) HandleUpdateCompanySettings(w http.ResponseWriter, r *http.Request) {
	company := middleware.GetCompany(r)

	// Parse form
	if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, "/admin/settings?error=Failed+to+parse+form", http.StatusSeeOther)
		return
	}

	// Get form values
	apiKey := strings.TrimSpace(r.FormValue("email_api_key"))
	fromEmail := strings.TrimSpace(r.FormValue("email_from_email"))
	fromName := strings.TrimSpace(r.FormValue("email_from_name"))

	// Update company email settings
	err := h.DB.UpdateCompanyEmailSettings(company.ID, apiKey, fromEmail, fromName)
	if err != nil {
		http.Redirect(w, r, "/admin/settings?error=Failed+to+update+settings", http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/admin/settings?message=Settings+updated+successfully", http.StatusSeeOther)
}

// HandleTestEmail sends a test email with the provided settings
func (h *AdminHandler) HandleTestEmail(w http.ResponseWriter, r *http.Request) {
	// Parse form
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	// Get form values
	apiKey := strings.TrimSpace(r.FormValue("email_api_key"))
	fromEmail := strings.TrimSpace(r.FormValue("email_from_email"))
	fromName := strings.TrimSpace(r.FormValue("email_from_name"))

	// Validate required fields
	if apiKey == "" || fromEmail == "" || fromName == "" {
		http.Error(w, "All email settings are required for testing", http.StatusBadRequest)
		return
	}

	// Send test email using the email service
	// Note: This would need access to the email service, which we don't have in the admin handler
	// For now, we'll simulate the test by validating the settings
	w.Header().Set("Content-Type", "text/plain")

	// Basic validation of the settings
	if len(apiKey) < 10 {
		w.Write([]byte("Error: API key appears to be too short"))
		return
	}

	if !strings.Contains(fromEmail, "@") {
		w.Write([]byte("Error: Invalid email format"))
		return
	}

	w.Write([]byte("Test email settings validated successfully!\n\n" +
		"Settings:\n" +
		"API Key: " + apiKey[:10] + "...\n" +
		"From Email: " + fromEmail + "\n" +
		"From Name: " + fromName + "\n\n" +
		"Note: To actually send a test email, the email service would need to be integrated into the admin handler."))
}

// ShowPuzzleCodes displays the codes list for a puzzle
func (h *AdminHandler) ShowPuzzleCodes(w http.ResponseWriter, r *http.Request) {
	company := middleware.GetCompany(r)
	user := middleware.GetUser(r)

	// Extract puzzle ID
	puzzleIDStr := strings.TrimPrefix(r.URL.Path, "/admin/puzzles/")
	puzzleIDStr = strings.TrimSuffix(puzzleIDStr, "/codes")
	puzzleID, err := strconv.Atoi(puzzleIDStr)
	if err != nil {
		http.Error(w, "Invalid puzzle ID", http.StatusBadRequest)
		return
	}

	// Get puzzle
	puzzle, err := h.DB.GetPuzzleByID(puzzleID)
	if err != nil || puzzle == nil || puzzle.CompanyID != company.ID {
		http.Error(w, "Puzzle not found", http.StatusNotFound)
		return
	}

	// Get completions
	completions, err := h.DB.GetCompletionsByPuzzle(puzzleID)
	if err != nil {
		http.Error(w, "Failed to load completions", http.StatusInternalServerError)
		return
	}

	// Apply filter if specified
	filter := r.URL.Query().Get("filter")
	if filter != "" {
		completions = filterCompletions(completions, filter)
	}

	// Get message/error from query params
	message := r.URL.Query().Get("message")
	errorMsg := r.URL.Query().Get("error")

	component := templates.AdminPuzzleCodes(company.Name, user.Email, puzzle, completions, filter, message, errorMsg)
	if err := component.Render(r.Context(), w); err != nil {
		http.Error(w, fmt.Sprintf("Failed to render template: %v", err), http.StatusInternalServerError)
		return
	}
}

// HandleExportCodes exports codes as CSV
func (h *AdminHandler) HandleExportCodes(w http.ResponseWriter, r *http.Request) {
	company := middleware.GetCompany(r)

	// Extract puzzle ID
	puzzleIDStr := strings.TrimPrefix(r.URL.Path, "/admin/puzzles/")
	puzzleIDStr = strings.TrimSuffix(puzzleIDStr, "/codes/export")
	puzzleID, err := strconv.Atoi(puzzleIDStr)
	if err != nil {
		http.Error(w, "Invalid puzzle ID", http.StatusBadRequest)
		return
	}

	// Get puzzle
	puzzle, err := h.DB.GetPuzzleByID(puzzleID)
	if err != nil || puzzle == nil || puzzle.CompanyID != company.ID {
		http.Error(w, "Puzzle not found", http.StatusNotFound)
		return
	}

	// Get completions
	completions, err := h.DB.GetCompletionsByPuzzle(puzzleID)
	if err != nil {
		http.Error(w, "Failed to load completions", http.StatusInternalServerError)
		return
	}

	// Set CSV headers
	filename := fmt.Sprintf("puzzle_%d_codes_%s.csv", puzzleID, time.Now().Format("2006-01-02"))
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))

	// Write CSV header
	w.Write([]byte("Discount Code,Email,Date Generated,Expires At,Status,Moves,Time (seconds)\n"))

	// Write CSV data
	for _, completion := range completions {
		status := "Active"
		if completion.IsUsed {
			status = "Used"
		} else if completion.ExpiresAt != nil && time.Now().After(*completion.ExpiresAt) {
			status = "Expired"
		}

		expiresAt := "Never"
		if completion.ExpiresAt != nil {
			expiresAt = completion.ExpiresAt.Format("2006-01-02 15:04:05")
		}

		line := fmt.Sprintf("%s,%s,%s,%s,%s,%d,%d\n",
			completion.DiscountCode,
			completion.Email,
			completion.CreatedAt.Format("2006-01-02 15:04:05"),
			expiresAt,
			status,
			completion.Moves,
			completion.TimeSeconds,
		)
		w.Write([]byte(line))
	}
}

// filterCompletions filters completions by status
func filterCompletions(completions []models.Completion, filter string) []models.Completion {
	var filtered []models.Completion
	now := time.Now()

	for _, completion := range completions {
		switch filter {
		case "active":
			if !completion.IsUsed && (completion.ExpiresAt == nil || now.Before(*completion.ExpiresAt)) {
				filtered = append(filtered, completion)
			}
		case "expired":
			if !completion.IsUsed && completion.ExpiresAt != nil && now.After(*completion.ExpiresAt) {
				filtered = append(filtered, completion)
			}
		case "used":
			if completion.IsUsed {
				filtered = append(filtered, completion)
			}
		default:
			filtered = append(filtered, completion)
		}
	}

	return filtered
}
