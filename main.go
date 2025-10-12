package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"puzzle/handlers"
	"puzzle/models"
)

func main() {
	// Load configuration
	config, err := models.LoadConfig("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize database
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

	// Create handlers
	gameHandler := handlers.NewGameHandler(config)
	completionHandler := handlers.NewCompletionHandler(db, config, emailService)

	// Set up routes
	mux := http.NewServeMux()

	// Game page
	mux.Handle("/", gameHandler)

	// Completion endpoint
	mux.Handle("/complete", completionHandler)

	// Static files (CSS, JS)
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Image files
	mux.Handle("/images/", http.StripPrefix("/images/", http.FileServer(http.Dir(config.Images.Directory))))

	// Start server
	addr := fmt.Sprintf(":%d", config.Server.Port)
	log.Printf("Starting server on http://localhost%s", addr)
	log.Printf("Puzzle configuration: %dx%d grid, %d%% discount",
		config.Puzzle.GridSize,
		config.Puzzle.GridSize,
		config.Puzzle.DiscountPercent)

	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
