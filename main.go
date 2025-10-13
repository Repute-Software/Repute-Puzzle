package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"puzzle/handlers"
	"puzzle/models"
	"runtime/debug"
	"syscall"
	"time"
)

// panicRecovery is middleware that recovers from panics and logs them
func panicRecovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("PANIC RECOVERED: %v\nStack trace:\n%s", err, debug.Stack())
				http.Error(w, "Internal server error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// loggingMiddleware logs all requests
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s - %v", r.Method, r.URL.Path, time.Since(start))
	})
}

func main() {
	log.Println("Starting Repute Puzzle application...")

	// Load configuration
	config, err := models.LoadConfig("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize database with connection pooling
	db, err := models.InitDB(config.Database.Path)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	log.Println("Database initialized successfully")

	// Override email API key from environment variable (more secure)
	if apiKey := os.Getenv("RESEND_API_KEY"); apiKey != "" {
		config.Email.APIKey = apiKey
		log.Println("Email API key loaded from environment variable")
	}

	// Initialize email service
	var emailService *models.EmailService
	if config.Email.Enabled {
		if config.Email.APIKey != "" {
			emailService = models.NewEmailService(&config.Email)
			log.Printf("Email service enabled (sending from: %s)", config.Email.FromEmail)
		} else {
			log.Println("Email enabled but no API key provided - emails will not be sent")
		}
	} else {
		log.Println("Email service disabled")
	}

	// Initialize translation cache
	if err := models.InitTranslationCache(); err != nil {
		log.Fatalf("Failed to initialize translation cache: %v", err)
	}
	log.Println("Translation cache initialized")

	// Create handlers
	gameHandler := handlers.NewGameHandler(config)
	embedHandler := handlers.NewEmbedHandler(config)
	completionHandler := handlers.NewCompletionHandler(db, config, emailService)

	// Set up routes
	mux := http.NewServeMux()

	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "OK")
	})

	// Game page
	mux.Handle("/", gameHandler)

	// Embed page
	mux.Handle("/embed", embedHandler)

	// Completion endpoint
	mux.Handle("/complete", completionHandler)

	// Static files (CSS, JS)
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Image files
	mux.Handle("/images/", http.StripPrefix("/images/", http.FileServer(http.Dir(config.Images.Directory))))

	// Wrap with middleware
	handler := panicRecovery(loggingMiddleware(mux))

	// Configure server with proper timeouts
	addr := fmt.Sprintf(":%d", config.Server.Port)
	server := &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		MaxHeaderBytes:    1 << 20, // 1 MB
	}

	log.Printf("Starting server on http://localhost%s", addr)
	log.Printf("Puzzle configuration: %dx%d grid, %d%% discount",
		config.Puzzle.GridSize,
		config.Puzzle.GridSize,
		config.Puzzle.DiscountPercent)

	// Start server in goroutine for graceful shutdown
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit

	log.Printf("Received signal: %v - initiating graceful shutdown...", sig)

	// Graceful shutdown with 10 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
		os.Exit(1)
	}

	log.Println("Server stopped gracefully")
	os.Exit(0)
}
