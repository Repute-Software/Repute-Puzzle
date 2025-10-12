package main

import (
	"fmt"
	"log"
	"net/http"
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

	// Create handlers
	gameHandler := handlers.NewGameHandler(config)
	completionHandler := handlers.NewCompletionHandler(db, config)

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
