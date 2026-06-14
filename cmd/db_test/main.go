// Command db_test is a manual database health check: it connects, migrates, and
// round-trips a user + sim account. It requires a live Postgres (via env/.env)
// and is not part of the automated test suite.
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/faisal-990/ProjectInvestApp/internal/platform/config"
	"github.com/faisal-990/ProjectInvestApp/internal/platform/models"
	"github.com/faisal-990/ProjectInvestApp/internal/platform/repository"
	"github.com/faisal-990/ProjectInvestApp/internal/platform/storage"
)

func main() {
	fmt.Println("🛠️  Starting Database Health Check...")

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("❌ Failed to load config: %v", err)
	}

	db, err := storage.Connect(cfg.DB)
	if err != nil {
		log.Fatalf("❌ Failed to connect to DB: %v", err)
	}
	fmt.Println("✅ Database connected and migrated")

	authRepo := repository.NewAuthRepo(db)
	ctx := context.Background()

	testEmail := "veryrandomemail@gmail.com"

	existing, _ := authRepo.GetUserByEmail(ctx, testEmail)
	if existing != nil {
		fmt.Println("ℹ️  Test user already exists. Skipping creation.")
	} else {
		fmt.Println("📝 Creating new test user + sim account...")
		newUser := &models.User{
			Name:          "Benjamin Graham",
			Email:         testEmail,
			Password:      "PLACEHOLDER_HASH", // real signup hashes via bcrypt (T1)
			Role:          "user",
			IsActive:      true,
			EmailVerified: true,
			Accounts: []models.Account{
				{Mode: models.ModeSim, Currency: "USD", Balance: 100000, IsActive: true},
			},
		}
		if err := authRepo.AddUser(ctx, newUser); err != nil {
			log.Fatalf("❌ AddUser failed: %v", err)
		}
		fmt.Println("✅ AddUser success")
	}

	fmt.Println("🔍 Verifying by fetching from DB...")
	fetchedUser, err := authRepo.GetUserByEmail(ctx, testEmail)
	if err != nil {
		log.Fatalf("❌ Fetch failed: %v", err)
	}
	if fetchedUser == nil {
		log.Fatalf("❌ Error: user was added but not found!")
	}

	fmt.Println("---------------------------------------------------")
	fmt.Printf("🎉 SUCCESS! User found in DB:\n")
	fmt.Printf("UUID:  %v\n", fetchedUser.ID)
	fmt.Printf("Name:  %s\n", fetchedUser.Name)
	fmt.Printf("Email: %s\n", fetchedUser.Email)
	fmt.Println("---------------------------------------------------")
}
