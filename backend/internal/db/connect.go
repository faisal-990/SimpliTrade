package db

import (
	"log"
	"os"
    "fmt"
	"github.com/faisal-990/ProjectInvestApp/backend/internal/models"
	"gorm.io/gorm"
    "gorm.io/driver/postgres"
    "github.com/joho/godotenv"
)

func Connect() (*gorm.DB,error){
    log.Print("starting the db connection ....")
	
     if err := godotenv.Load(); err !=nil{
         log.Panic(err)
     }
    dsn := fmt.Sprintf(
    "host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=%s",
    os.Getenv("HOST"),
    os.Getenv("DBUSER"),
    os.Getenv("DBPASS"),
    os.Getenv("DBNAME"),
    os.Getenv("DBPORT"),
    os.Getenv("SSLMODE"),
    os.Getenv("TIMEZONE"),
)
    log.Println("successfully loaded the environment files")
	
	
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect to db: ", err)
        return nil,err
	}
    
    log.Println("Established the gorm connection to postgres db")

	// Check if migration succeeds
    dbmodels := []interface{}{
       models.User{},
       models.Investor{},
       models.Performance{},
       models.Stock{},
       models.StockPrice{},
       models.Holding{},
       models.Trade{},
       models.Follow{},
    }
    

	if err := db.AutoMigrate(dbmodels...); err != nil {
		log.Fatal("failed to migrate database: ", err)
        return nil,err
	}
    log.Println("successfully automigrated the db")
    return db,nil; 


}
