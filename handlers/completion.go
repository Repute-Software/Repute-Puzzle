package handlers

import (
	"log"
	"net/http"
	"puzzle/models"
	"puzzle/templates"
	"strconv"
	"strings"
)

// CompletionHandler handles puzzle completion and discount code generation
type CompletionHandler struct {
	DB           *models.DB
	Config       *models.Config
	EmailService *models.EmailService
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
	return &CompletionHandler{
		DB:           db,
		Config:       config,
		EmailService: emailService,
	}
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

	// Generate unique discount code
	var discountCode string
	maxRetries := 10

	for i := 0; i < maxRetries; i++ {
		code, err := models.GenerateDiscountCode(h.Config.Puzzle.DiscountPercent)
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
		email,
		discountCode,
		h.Config.Puzzle.GridSize,
		moves,
		timeSeconds,
	)
	if err != nil {
		h.renderErrorWithTranslations(w, r, "Failed to save completion record", translations)
		return
	}

	// Send email with discount code (in background)
	lang := r.FormValue("lang")
	if h.EmailService != nil {
		go func() {
			// Send email in background (don't block response)
			if err := h.EmailService.SendDiscountCode(email, discountCode, h.Config.Puzzle.DiscountPercent, lang); err != nil {
				// Log error but don't fail the request
				log.Printf("Failed to send email to %s: %v\n", email, err)
			} else {
				log.Printf("Successfully sent discount code email to %s\n", email)
			}
		}()
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
