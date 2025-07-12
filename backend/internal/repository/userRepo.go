package repository

import (
	"github.com/faisal-990/ProjectInvestApp/backend/internal/models"
	"github.com/faisal-990/ProjectInvestApp/backend/internal/utils"
	"github.com/google/uuid"
	"gorm.io/gorm"
)



func CreateUser(db *gorm.DB)(error){

       
    users := []models.User{
    {Name: "Sawez Faisal", Email: "sawez.faisal@example.com", Password: "password1"},
    {Name: "Arjun Mehta", Email: "arjun.mehta@example.com", Password: "password2"},
    {Name: "Priya Sharma", Email: "priya.sharma@example.com", Password: "password3"},
    {Name: "Rahul Nair", Email: "rahul.nair@example.com", Password: "password4"},
    {Name: "Neha Verma", Email: "neha.verma@example.com", Password: "password5"},
    {Name: "Kabir Bansal", Email: "kabir.bansal@example.com", Password: "password6"},
    {Name: "Divya Menon", Email: "divya.menon@example.com", Password: "password7"},
    {Name: "Rohan Gupta", Email: "rohan.gupta@example.com", Password: "password8"},
    {Name: "Ananya Reddy", Email: "ananya.reddy@example.com", Password: "password9"},
    {Name: "Ishaan Kapoor", Email: "ishaan.kapoor@example.com", Password: "password10"},
    {Name: "Tanvi Joshi", Email: "tanvi.joshi@example.com", Password: "password11"},
    {Name: "Aarav Singh", Email: "aarav.singh@example.com", Password: "password12"},
    {Name: "Meera Dutta", Email: "meera.dutta@example.com", Password: "password13"},
    {Name: "Aditya Roy", Email: "aditya.roy@example.com", Password: "password14"},
    {Name: "Sneha Patil", Email: "sneha.patil@example.com", Password: "password15"},
}

    result := db.Create(&users)
    if result.Error != nil {
        utils.LogError("failed to create user",result.Error)
        return result.Error
    }
   utils.LogInfo("user succesfully created") 
    return nil;
}

func DeleteUser(db *gorm.DB, userId uuid.UUID)(error){
    
    result := db.Delete(&models.User{} , userId)
    if result.Error != nil{
        utils.LogError("failed to find user with id",result.Error)
    }
    utils.LogInfoF("successfully Deleted user with id",userId.String())
    
    return nil
}

func GetUser(db *gorm.DB , userId uuid.UUID)(*models.User,error){
    var user models.User
    result := db.First(&user,userId);
    
    if result.Error != nil{
        utils.LogError("failed to get user information",result.Error)
    }
    
    return &user,nil
}

func UpdateUser(db *gorm.DB, user *models.User) error {
    user.Name = "New Name"
    user.Email = "new@example.com"
    return db.Save(user).Error
}


