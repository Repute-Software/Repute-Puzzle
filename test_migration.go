package main

import (
	"fmt"
	"log"
	"os"
	"puzzle/models"
)

// This is a simple test script to verify database migration
func main() {
	// Use a test database
	testDB := "./data/test_migration.db"

	// Remove test database if it exists
	os.Remove(testDB)

	log.Println("Testing database migration...")

	// Initialize database (this will create all tables)
	db, err := models.InitDB(testDB)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	log.Println("✓ Database initialized successfully")

	// Test 1: Create a company
	company, err := db.CreateCompany("Test Company", "test-company")
	if err != nil {
		log.Fatalf("Failed to create company: %v", err)
	}
	log.Printf("✓ Created company: %s (ID: %d, Slug: %s)", company.Name, company.ID, company.Slug)

	// Test 2: Create a user
	user, err := db.CreateUser(company.ID, "admin@testcompany.com", "password123", "admin")
	if err != nil {
		log.Fatalf("Failed to create user: %v", err)
	}
	log.Printf("✓ Created user: %s (ID: %d, Role: %s)", user.Email, user.ID, user.Role)

	// Test 3: Test authentication
	authenticatedUser, err := db.AuthenticateUser("admin@testcompany.com", "password123")
	if err != nil {
		log.Fatalf("Failed to authenticate user: %v", err)
	}
	if authenticatedUser == nil {
		log.Fatal("Authentication failed!")
	}
	log.Printf("✓ User authenticated successfully")

	// Test 4: Create a session
	session, err := db.CreateSession(user.ID, 24*60*60*1000000000) // 24 hours
	if err != nil {
		log.Fatalf("Failed to create session: %v", err)
	}
	log.Printf("✓ Created session: %s", session.ID)

	// Test 5: Get session with user
	sessionWithUser, err := db.GetSessionWithUser(session.ID)
	if err != nil {
		log.Fatalf("Failed to get session with user: %v", err)
	}
	if sessionWithUser == nil {
		log.Fatal("Session not found!")
	}
	log.Printf("✓ Retrieved session with user: %s and company: %s", sessionWithUser.User.Email, sessionWithUser.Company.Name)

	// Test 6: Create a puzzle
	puzzle := &models.Puzzle{
		CompanyID:       company.ID,
		Name:            "Summer Sale Puzzle",
		Slug:            "summer-sale",
		ImagePath:       "images/1/1.png",
		GridSize:        3,
		DiscountPercent: 15,
		TimeLimit:       90,
		TimerMode:       "first_move",
		CountdownTime:   15,
		ScrambleMoves:   25,
		AutoSolveSpeed:  50,
		IsActive:        true,
	}
	createdPuzzle, err := db.CreatePuzzle(puzzle)
	if err != nil {
		log.Fatalf("Failed to create puzzle: %v", err)
	}
	log.Printf("✓ Created puzzle: %s (ID: %d, Slug: %s)", createdPuzzle.Name, createdPuzzle.ID, createdPuzzle.Slug)

	// Test 7: Get puzzle by slug
	foundPuzzle, err := db.GetPuzzleBySlug("test-company", "summer-sale")
	if err != nil {
		log.Fatalf("Failed to get puzzle by slug: %v", err)
	}
	if foundPuzzle == nil {
		log.Fatal("Puzzle not found by slug!")
	}
	log.Printf("✓ Retrieved puzzle by slug: %s", foundPuzzle.Name)

	// Test 8: Save a completion
	discountCode, err := models.GenerateDiscountCode(foundPuzzle.DiscountPercent, foundPuzzle.ID)
	if err != nil {
		log.Fatalf("Failed to generate discount code: %v", err)
	}
	err = db.SaveCompletion(foundPuzzle.ID, "player@example.com", discountCode, 42, 85)
	if err != nil {
		log.Fatalf("Failed to save completion: %v", err)
	}
	log.Printf("✓ Saved completion with discount code: %s", discountCode)

	// Test 9: Get completions by puzzle
	completions, err := db.GetCompletionsByPuzzle(foundPuzzle.ID)
	if err != nil {
		log.Fatalf("Failed to get completions: %v", err)
	}
	log.Printf("✓ Retrieved %d completion(s) for puzzle", len(completions))

	// Test 10: Get completions by company
	companyCompletions, err := db.GetCompletionsByCompany(company.ID)
	if err != nil {
		log.Fatalf("Failed to get company completions: %v", err)
	}
	log.Printf("✓ Retrieved %d completion(s) for company", len(companyCompletions))

	// Test 11: Get puzzle stats
	stats, err := db.GetPuzzleCompletionStats(foundPuzzle.ID)
	if err != nil {
		log.Fatalf("Failed to get puzzle stats: %v", err)
	}
	log.Printf("✓ Puzzle stats: %d completions, avg moves: %.1f, avg time: %.1f",
		stats["total_completions"], stats["avg_moves"], stats["avg_time"])

	log.Println("\n✅ All migration tests passed!")

	// Clean up test database
	os.Remove(testDB)
	fmt.Println("\nTest database cleaned up.")
}

