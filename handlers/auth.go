package handlers

import (
	"log"
	"net/http"
	"puzzle/middleware"
	"puzzle/models"
	"puzzle/templates"
	"strings"
	"time"
)

// AuthHandler handles authentication routes
type AuthHandler struct {
	DB     *models.DB
	Config *models.Config
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(db *models.DB, config *models.Config) *AuthHandler {
	return &AuthHandler{
		DB:     db,
		Config: config,
	}
}

// ShowLogin displays the login page
func (h *AuthHandler) ShowLogin(w http.ResponseWriter, r *http.Request) {
	// Check if already logged in
	user := middleware.GetUser(r)
	if user != nil {
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
		return
	}

	// Get redirect URL if provided
	redirectURL := r.URL.Query().Get("redirect")
	errorMsg := r.URL.Query().Get("error")

	component := templates.Login(redirectURL, errorMsg)
	if err := component.Render(r.Context(), w); err != nil {
		http.Error(w, "Failed to render login page", http.StatusInternalServerError)
	}
}

// HandleLogin processes login form submission
func (h *AuthHandler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse form data
	if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, "/login?error=Invalid+form+data", http.StatusSeeOther)
		return
	}

	email := strings.TrimSpace(r.FormValue("email"))
	password := r.FormValue("password")
	redirectURL := r.FormValue("redirect")

	// Validate inputs
	if email == "" || password == "" {
		http.Redirect(w, r, "/login?error=Email+and+password+required", http.StatusSeeOther)
		return
	}

	// Authenticate user
	log.Printf("Login attempt for: %s", email)
	user, err := h.DB.AuthenticateUser(email, password)
	if err != nil {
		log.Printf("Authentication error for %s: %v", email, err)
		http.Redirect(w, r, "/login?error=Authentication+failed", http.StatusSeeOther)
		return
	}

	if user == nil {
		log.Printf("Invalid credentials for: %s", email)
		http.Redirect(w, r, "/login?error=Invalid+email+or+password", http.StatusSeeOther)
		return
	}

	// Create session (7 days)
	log.Printf("User %s authenticated successfully, creating session...", email)
	session, err := h.DB.CreateSession(user.ID, 7*24*time.Hour)
	if err != nil {
		log.Printf("Failed to create session for %s: %v", email, err)
		http.Redirect(w, r, "/login?error=Failed+to+create+session", http.StatusSeeOther)
		return
	}

	// Set session cookie
	log.Printf("Session created for %s: %s", email, session.ID)
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    session.ID,
		Path:     "/",
		MaxAge:   7 * 24 * 60 * 60, // 7 days in seconds
		HttpOnly: true,
		Secure:   r.TLS != nil, // Only set Secure flag if using HTTPS
		SameSite: http.SameSiteLaxMode,
	})

	// Redirect to destination
	if redirectURL != "" && strings.HasPrefix(redirectURL, "/") {
		log.Printf("Redirecting %s to: %s", email, redirectURL)
		http.Redirect(w, r, redirectURL, http.StatusSeeOther)
	} else {
		log.Printf("Redirecting %s to: /admin", email)
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
	}
}

// ShowSignup displays the signup page
func (h *AuthHandler) ShowSignup(w http.ResponseWriter, r *http.Request) {
	// Check if already logged in
	user := middleware.GetUser(r)
	if user != nil {
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
		return
	}

	errorMsg := r.URL.Query().Get("error")

	component := templates.Signup(errorMsg)
	if err := component.Render(r.Context(), w); err != nil {
		http.Error(w, "Failed to render signup page", http.StatusInternalServerError)
	}
}

// HandleSignup processes signup form submission
func (h *AuthHandler) HandleSignup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse form data
	if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, "/signup?error=Invalid+form+data", http.StatusSeeOther)
		return
	}

	companyName := strings.TrimSpace(r.FormValue("company_name"))
	email := strings.TrimSpace(r.FormValue("email"))
	password := r.FormValue("password")
	confirmPassword := r.FormValue("confirm_password")

	// Validate inputs
	if companyName == "" {
		http.Redirect(w, r, "/signup?error=Company+name+required", http.StatusSeeOther)
		return
	}

	if email == "" || !isValidEmail(email) {
		http.Redirect(w, r, "/signup?error=Valid+email+required", http.StatusSeeOther)
		return
	}

	if password != confirmPassword {
		http.Redirect(w, r, "/signup?error=Passwords+do+not+match", http.StatusSeeOther)
		return
	}

	// Validate password strength
	if err := models.ValidatePassword(password); err != nil {
		http.Redirect(w, r, "/signup?error="+err.Error(), http.StatusSeeOther)
		return
	}

	// Check if email already exists
	existingUser, err := h.DB.GetUserByEmail(email)
	if err != nil {
		log.Printf("Error checking existing user: %v", err)
		http.Redirect(w, r, "/signup?error=Database+error", http.StatusSeeOther)
		return
	}

	if existingUser != nil {
		http.Redirect(w, r, "/signup?error=Email+already+registered", http.StatusSeeOther)
		return
	}

	// Generate company slug
	companySlug := models.GenerateSlug(companyName)

	// Check if slug already exists (add number if needed)
	existingCompany, err := h.DB.GetCompanyBySlug(companySlug)
	if err != nil {
		log.Printf("Error checking existing company: %v", err)
		http.Redirect(w, r, "/signup?error=Database+error", http.StatusSeeOther)
		return
	}

	if existingCompany != nil {
		// Append timestamp to make slug unique
		companySlug = companySlug + "-" + time.Now().Format("20060102")
	}

	// Create company
	company, err := h.DB.CreateCompany(companyName, companySlug)
	if err != nil {
		log.Printf("Failed to create company: %v", err)
		http.Redirect(w, r, "/signup?error=Failed+to+create+company", http.StatusSeeOther)
		return
	}

	// Create admin user
	user, err := h.DB.CreateUser(company.ID, email, password, "admin")
	if err != nil {
		log.Printf("Failed to create user: %v", err)
		http.Redirect(w, r, "/signup?error=Failed+to+create+user", http.StatusSeeOther)
		return
	}

	// Create session
	session, err := h.DB.CreateSession(user.ID, 7*24*time.Hour)
	if err != nil {
		log.Printf("Failed to create session: %v", err)
		http.Redirect(w, r, "/signup?error=Failed+to+create+session", http.StatusSeeOther)
		return
	}

	// Set session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    session.ID,
		Path:     "/",
		MaxAge:   7 * 24 * 60 * 60, // 7 days in seconds
		HttpOnly: true,
		Secure:   r.TLS != nil,
		SameSite: http.SameSiteLaxMode,
	})

	log.Printf("New company registered: %s (slug: %s) by %s", company.Name, company.Slug, user.Email)

	// Redirect to admin dashboard
	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

// HandleLogout processes logout
func (h *AuthHandler) HandleLogout(w http.ResponseWriter, r *http.Request) {
	// Get session from context
	session := middleware.GetSession(r)
	if session != nil {
		// Delete session from database
		if err := h.DB.DeleteSession(session.ID); err != nil {
			log.Printf("Failed to delete session: %v", err)
		}
	}

	// Clear session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})

	// Redirect to login
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
