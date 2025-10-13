package handlers

import (
	"log"
	"net/http"
	"puzzle/models"
	"puzzle/templates"
	"strconv"
	"strings"
	"sync"
	"time"
)

// EmailJob represents an email to be sent
type EmailJob struct {
	Email           string
	DiscountCode    string
	DiscountPercent int
	Lang            string
	CompanySettings *models.CompanyEmailSettings
}

// CompletionHandler handles puzzle completion and discount code generation
type CompletionHandler struct {
	DB           *models.DB
	Config       *models.Config
	EmailService *models.EmailService
	emailQueue   chan EmailJob
	workerWg     sync.WaitGroup
}

// startEmailWorkers starts a pool of workers to process email jobs
func (h *CompletionHandler) startEmailWorkers(numWorkers int) {
	h.emailQueue = make(chan EmailJob, 100) // Buffer up to 100 email jobs

	for i := 0; i < numWorkers; i++ {
		h.workerWg.Add(1)
		go func(workerID int) {
			defer h.workerWg.Done()
			for job := range h.emailQueue {
				// Send email (email service has its own timeout)
				if h.EmailService != nil {
					if err := h.EmailService.SendDiscountCodeWithSettings(job.Email, job.DiscountCode, job.DiscountPercent, job.Lang, job.CompanySettings); err != nil {
						log.Printf("Worker %d: Failed to send email to %s: %v", workerID, job.Email, err)
					} else {
						log.Printf("Worker %d: Successfully sent discount code email to %s", workerID, job.Email)
					}
				}
			}
		}(i)
	}
}

// getTranslations loads translations from form data or defaults to English
func (h *CompletionHandler) getTranslations(r *http.Request) *models.Translations {
	lang := r.FormValue("lang")
	if lang == "" {
		lang = "en"
	}
	translations, err := models.LoadTranslations(lang)
	if err != nil {
		translations, _ = models.LoadTranslations("en")
	}
	return translations
}

// NewCompletionHandler creates a new completion handler
func NewCompletionHandler(db *models.DB, config *models.Config, emailService *models.EmailService) *CompletionHandler {
	handler := &CompletionHandler{
		DB:           db,
		Config:       config,
		EmailService: emailService,
	}

	// Start 3 worker goroutines to handle email sending
	// This prevents unbounded goroutine growth
	handler.startEmailWorkers(3)

	return handler
}

// ServeHTTP handles the completion form submission (old backward-compatible route)
func (h *CompletionHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// For backward compatibility - uses puzzle_id = 1 and config.yaml discount
	h.handleCompletion(w, r, 1, h.Config.Puzzle.DiscountPercent)
}

// ServeCompletionForPuzzle handles completion for a specific puzzle from database
func (h *CompletionHandler) ServeCompletionForPuzzle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract company and puzzle slugs from URL path
	// Expected format: /complete/:companySlug/:puzzleSlug
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

	// Handle completion with puzzle's settings
	h.handleCompletion(w, r, puzzle.ID, puzzle.DiscountPercent)
}

// handleCompletion is the shared completion logic
func (h *CompletionHandler) handleCompletion(w http.ResponseWriter, r *http.Request, puzzleID int, discountPercent int) {
	// Parse form data
	if err := r.ParseForm(); err != nil {
		h.renderError(w, r, "Failed to parse form data")
		return
	}

	// Load translations
	translations := h.getTranslations(r)

	// Extract form values
	email := strings.TrimSpace(r.FormValue("email"))
	movesStr := r.FormValue("moves")
	timeStr := r.FormValue("time")

	// Validate email
	if email == "" || !isValidEmail(email) {
		h.renderErrorWithTranslations(w, r, translations.Error.InvalidEmail, translations)
		return
	}

	// Parse moves and time
	moves, err := strconv.Atoi(movesStr)
	if err != nil || moves < 0 {
		h.renderErrorWithTranslations(w, r, "Invalid moves count", translations)
		return
	}

	timeSeconds, err := strconv.Atoi(timeStr)
	if err != nil || timeSeconds < 0 {
		h.renderErrorWithTranslations(w, r, "Invalid time value", translations)
		return
	}

	// Generate unique discount code
	var discountCode string
	maxRetries := 10

	for i := 0; i < maxRetries; i++ {
		code, err := models.GenerateDiscountCode(discountPercent, puzzleID)
		if err != nil {
			h.renderErrorWithTranslations(w, r, translations.Error.FailedGenerate, translations)
			return
		}

		// Check if code already exists
		existing, err := h.DB.GetCompletionByCode(code)
		if err != nil {
			h.renderErrorWithTranslations(w, r, "Database error", translations)
			return
		}

		if existing == nil {
			discountCode = code
			break
		}
	}

	if discountCode == "" {
		h.renderErrorWithTranslations(w, r, translations.Error.FailedGenerate, translations)
		return
	}

	// Calculate expiration date if puzzle has expiration set
	var expiresAt *time.Time
	if puzzleID > 0 { // Only for database puzzles, not old config puzzles
		puzzle, err := h.DB.GetPuzzleByID(puzzleID)
		if err == nil && puzzle != nil && puzzle.CodeExpirationDays > 0 {
			expiration := time.Now().AddDate(0, 0, puzzle.CodeExpirationDays)
			expiresAt = &expiration
		}
	}

	// Save completion to database
	err = h.DB.SaveCompletion(
		puzzleID,
		email,
		discountCode,
		moves,
		timeSeconds,
		expiresAt,
	)
	if err != nil {
		h.renderErrorWithTranslations(w, r, "Failed to save completion record", translations)
		return
	}

	// Queue email for sending (non-blocking)
	lang := r.FormValue("lang")
	if h.EmailService != nil && h.emailQueue != nil {
		// Get company email settings for this puzzle
		var companySettings *models.CompanyEmailSettings
		if puzzleID > 0 { // Only for database puzzles, not old config puzzles
			puzzle, err := h.DB.GetPuzzleByID(puzzleID)
			if err == nil && puzzle != nil {
				settings, err := h.DB.GetEmailSettingsForCompany(puzzle.CompanyID)
				if err == nil {
					companySettings = settings
				}
			}
		}

		// Send to worker queue with non-blocking select
		select {
		case h.emailQueue <- EmailJob{
			Email:           email,
			DiscountCode:    discountCode,
			DiscountPercent: discountPercent,
			Lang:            lang,
			CompanySettings: companySettings,
		}:
			log.Printf("Email job queued for %s (puzzle %d)", email, puzzleID)
		default:
			// Queue is full, log warning but don't block the response
			log.Printf("WARNING: Email queue full, could not queue email for %s", email)
		}
	}

	// Render success response
	component := templates.CompletionSuccess(discountCode, discountPercent, translations)
	if err := component.Render(r.Context(), w); err != nil {
		http.Error(w, "Failed to render response", http.StatusInternalServerError)
	}
}

// renderError renders an error message (kept for backward compatibility)
func (h *CompletionHandler) renderError(w http.ResponseWriter, r *http.Request, message string) {
	translations := h.getTranslations(r)
	h.renderErrorWithTranslations(w, r, message, translations)
}

// renderErrorWithTranslations renders an error message with translations
func (h *CompletionHandler) renderErrorWithTranslations(w http.ResponseWriter, r *http.Request, message string, translations *models.Translations) {
	w.WriteHeader(http.StatusBadRequest)
	component := templates.CompletionError(message, translations)
	if err := component.Render(r.Context(), w); err != nil {
		http.Error(w, message, http.StatusBadRequest)
	}
}

// isValidEmail performs basic email validation
func isValidEmail(email string) bool {
	// Basic validation: contains @ and at least one dot after @
	if len(email) < 3 {
		return false
	}
	atIndex := strings.Index(email, "@")
	if atIndex <= 0 || atIndex >= len(email)-1 {
		return false
	}
	dotIndex := strings.LastIndex(email, ".")
	return dotIndex > atIndex && dotIndex < len(email)-1
}
