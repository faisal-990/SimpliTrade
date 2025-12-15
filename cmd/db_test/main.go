package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/joho/godotenv"

	// Import your Twin Towers modules
	"github.com/faisal-990/ProjectInvestApp/internal/platform/models"
	"github.com/faisal-990/ProjectInvestApp/internal/platform/repository"
	"github.com/faisal-990/ProjectInvestApp/internal/platform/storage"
)

func main() {
	fmt.Println("🛠️  Starting Database Health Check...")

	// 1. Load .env (Critical for DB credentials)
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️  Warning: No .env file found (relying on system env vars)")
	}

	// 2. Connect to DB (Using your Platform Storage)
	db, err := storage.Connect()
	if err != nil {
		log.Fatalf("❌ Failed to connect to DB: %v", err)
	}
	fmt.Println("✅ Database Connected")

	// 3. AUTO-MIGRATE (Critical: Creates the 'users' table if missing)
	fmt.Println("🔄 Running AutoMigrate for User model...")
	if err := db.AutoMigrate(&models.User{}); err != nil {
		log.Fatalf("❌ Migration Failed: %v", err)
	}
	fmt.Println("✅ User Table Migrated")

	// 4. Initialize Repository
	authRepo := repository.NewAuthRepo(db)
	ctx := context.Background()

	// 5. TEST: Add a User
	testEmail := "veryrandomemail@gmail.com"

	// Check if user exists first to avoid duplicate key error
	existing, _ := authRepo.GetUserByEmail(ctx, testEmail)
	if existing != nil {
		fmt.Println("ℹ️  Test user already exists. Skipping creation.")
	} else {
		fmt.Println("📝 Creating new test user...")
		newUser := &models.User{
			Name:      "Benjamin Graham",
			Email:     testEmail,
			Password:  "securepassword123", // In real app, verify this is hashed!
			IsActive:  true,
			Balance:   100000.00,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		err = authRepo.AddUser(ctx, newUser)
		if err != nil {
			log.Fatalf("❌ AddUser Failed: %v", err)
		}
		fmt.Println("✅ AddUser Success")
	}

	// 6. TEST: Fetch the User back
	fmt.Println("🔍 Verifying data by fetching from DB...")
	fetchedUser, err := authRepo.GetUserByEmail(ctx, testEmail)
	if err != nil {
		log.Fatalf("❌ Fetch Failed: %v", err)
	}

	if fetchedUser == nil {
		log.Fatalf("❌ Error: User was added but not found!")
	}

	// 7. Visual Confirmation
	fmt.Println("---------------------------------------------------")
	fmt.Printf("🎉 SUCCESS! User Found in DB:\n")
	fmt.Printf("UUID:    %v\n", fetchedUser.ID)
	fmt.Printf("Name:    %s\n", fetchedUser.Name)
	fmt.Printf("Email:   %s\n", fetchedUser.Email)
	fmt.Printf("Balance: $%.2f\n", fetchedUser.Balance)
	fmt.Println("---------------------------------------------------")
}
