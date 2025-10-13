package handlers

import (
	"log"
	"net/http"
	"puzzle/models"
	"puzzle/templates"
	"strconv"
	"strings"
	"sync"
)

// EmailJob represents an email to be sent
type EmailJob struct {
	Email           string
	DiscountCode    string
	DiscountPercent int
	Lang            string
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
					if err := h.EmailService.SendDiscountCode(job.Email, job.DiscountCode, job.DiscountPercent, job.Lang); err != nil {
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

// ServeHTTP handles the completion form submission
func (h *CompletionHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

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

	// TODO: This is temporary - will be replaced with proper puzzle lookup
	// For now, use puzzle_id = 1 (default puzzle) for backward compatibility
	puzzleID := 1

	// Generate unique discount code
	var discountCode string
	maxRetries := 10

	for i := 0; i < maxRetries; i++ {
		code, err := models.GenerateDiscountCode(h.Config.Puzzle.DiscountPercent, puzzleID)
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

	// Save completion to database
	err = h.DB.SaveCompletion(
		puzzleID,
		email,
		discountCode,
		moves,
		timeSeconds,
	)
	if err != nil {
		h.renderErrorWithTranslations(w, r, "Failed to save completion record", translations)
		return
	}

	// Queue email for sending (non-blocking)
	lang := r.FormValue("lang")
	if h.EmailService != nil && h.emailQueue != nil {
		// Send to worker queue with non-blocking select
		select {
		case h.emailQueue <- EmailJob{
			Email:           email,
			DiscountCode:    discountCode,
			DiscountPercent: h.Config.Puzzle.DiscountPercent,
			Lang:            lang,
		}:
			log.Printf("Email job queued for %s", email)
		default:
			// Queue is full, log warning but don't block the response
			log.Printf("WARNING: Email queue full, could not queue email for %s", email)
		}
	}

	// Render success response
	component := templates.CompletionSuccess(discountCode, h.Config.Puzzle.DiscountPercent, translations)
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
