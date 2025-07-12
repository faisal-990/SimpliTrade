package main

import (
	"fmt"
	"log"
	"os"

	"github.com/faisal-990/ProjectInvestApp/backend/internal/db"
	"github.com/faisal-990/ProjectInvestApp/backend/internal/middlewares"
	"github.com/faisal-990/ProjectInvestApp/backend/internal/repository"
	"github.com/faisal-990/ProjectInvestApp/backend/internal/router"
	"github.com/faisal-990/ProjectInvestApp/backend/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env files")
	}
	log.Println("✅ Loaded .env files")

	db, err := db.Connect()
	if err != nil {
		log.Fatalf("❌ Failed to connect to DB: %s", err)
	}
	log.Println("✅ Connected to DB")

	//if err := repository.CreateUser(db); err != nil {
	//log.Fatalf("❌ Failed to seed/test DB: %s", err)
	//}
	//delete a user tesst
	userId := "41efa6f6-09ad-4724-9ce3-58de02ee6a96"
	parsedUuid, err := uuid.Parse(userId)
	//if(err !=nil){
	//utils.LogError("failed to parse uuid",err)
	//}
	//if err := repository.DeleteUser(db,parsedUuid); err != nil {
	//utils.LogError("failed to delete user",err)
	//}
	//
	//utils.LogInfoF("succesfully deleted user",userId)
	//
	//search a userbased on userId
	results, err := repository.GetUser(db, parsedUuid)
	if err != nil {
		utils.LogError("look this message", err)
	}
	fmt.Printf("username:%s \n email:%s \n", results.Name, results.Email)
	// update a user

	if err := repository.UpdateUser(db, results); err != nil {
		utils.LogError("failed to update user", err)
	}
	utils.LogInfoF("sucessfully updated user name and email of id:%s", userId)
	fmt.Printf("update name:%s \n updated email:%s \n", results.Name, results.Email)

	r := gin.Default()
	log.Println("✅ Created Gin engine")

	r.Use(middlewares.CORSMiddleware())

	router.InitializeRoutes(r)
	log.Println("✅ Initialized routes")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("🚀 Server running at :%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("❌ Failed to start server: %s", err)
	}
}
