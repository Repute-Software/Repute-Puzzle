package middleware

import (
	"context"
	"log"
	"net/http"
	"puzzle/models"
)

// ContextKey is a custom type for context keys to avoid collisions
type ContextKey string

const (
	// UserContextKey is the context key for the authenticated user
	UserContextKey ContextKey = "user"
	// CompanyContextKey is the context key for the user's company
	CompanyContextKey ContextKey = "company"
	// SessionContextKey is the context key for the session
	SessionContextKey ContextKey = "session"
)

// AuthMiddleware validates session and injects user/company into request context
type AuthMiddleware struct {
	DB *models.DB
}

// NewAuthMiddleware creates a new authentication middleware
func NewAuthMiddleware(db *models.DB) *AuthMiddleware {
	return &AuthMiddleware{DB: db}
}

// RequireAuth wraps a handler and requires authentication
func (m *AuthMiddleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get session cookie
		cookie, err := r.Cookie("session_id")
		if err != nil {
			// No session cookie, redirect to login
			http.Redirect(w, r, "/login?redirect="+r.URL.Path, http.StatusSeeOther)
			return
		}

		// Validate session
		sessionWithUser, err := m.DB.GetSessionWithUser(cookie.Value)
		if err != nil {
			log.Printf("Error getting session: %v", err)
			http.Redirect(w, r, "/login?redirect="+r.URL.Path, http.StatusSeeOther)
			return
		}

		if sessionWithUser == nil {
			// Session not found or expired
			// Clear the invalid cookie
			http.SetCookie(w, &http.Cookie{
				Name:     "session_id",
				Value:    "",
				Path:     "/",
				MaxAge:   -1,
				HttpOnly: true,
			})
			http.Redirect(w, r, "/login?redirect="+r.URL.Path, http.StatusSeeOther)
			return
		}

		// Inject user, company, and session into request context
		ctx := r.Context()
		ctx = context.WithValue(ctx, UserContextKey, sessionWithUser.User)
		ctx = context.WithValue(ctx, CompanyContextKey, sessionWithUser.Company)
		ctx = context.WithValue(ctx, SessionContextKey, &sessionWithUser.Session)

		// Call next handler with updated context
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// OptionalAuth checks for authentication but doesn't require it
func (m *AuthMiddleware) OptionalAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get session cookie
		cookie, err := r.Cookie("session_id")
		if err != nil {
			// No session, continue without auth
			next.ServeHTTP(w, r)
			return
		}

		// Try to validate session
		sessionWithUser, err := m.DB.GetSessionWithUser(cookie.Value)
		if err != nil || sessionWithUser == nil {
			// Invalid session, continue without auth
			next.ServeHTTP(w, r)
			return
		}

		// Inject user and company into request context
		ctx := r.Context()
		ctx = context.WithValue(ctx, UserContextKey, sessionWithUser.User)
		ctx = context.WithValue(ctx, CompanyContextKey, sessionWithUser.Company)
		ctx = context.WithValue(ctx, SessionContextKey, &sessionWithUser.Session)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetUser extracts the authenticated user from request context
func GetUser(r *http.Request) *models.User {
	user, ok := r.Context().Value(UserContextKey).(*models.User)
	if !ok {
		return nil
	}
	return user
}

// GetCompany extracts the company from request context
func GetCompany(r *http.Request) *models.Company {
	company, ok := r.Context().Value(CompanyContextKey).(*models.Company)
	if !ok {
		return nil
	}
	return company
}

// GetSession extracts the session from request context
func GetSession(r *http.Request) *models.Session {
	session, ok := r.Context().Value(SessionContextKey).(*models.Session)
	if !ok {
		return nil
	}
	return session
}

