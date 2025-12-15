package repository

import (
	"context"
	"errors"

	"github.com/faisal-990/ProjectInvestApp/internal/platform/models"
	"github.com/faisal-990/ProjectInvestApp/internal/platform/utils"
	"gorm.io/gorm"
)

type AuthRepo interface {
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	GetUserByID(ctx context.Context, id string) (*models.User, error)
	AddUser(ctx context.Context, user *models.User) error
}

type authRepo struct {
	DB *gorm.DB
}

//type User struct {
//ID          uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
//Name        string    `gorm:"type:varchar(100);not null"`
//Email       string    `gorm:"type:varchar(100);uniqueIndex;not null"`
//Password    string    `gorm:"not null"`
//IsActive    bool      `gorm:"default:true"`
//LastLoginAt *time.Time
//Balance     float64   `gorm:"type:numeric(15,2);default:100000"` // Starting simulation balance
//Trades      []Trade   `gorm:"foreignKey:UserID"`
//Holdings    []Holding `gorm:"foreignKey:UserID"`
//Follows     []Follow  `gorm:"foreignKey:FollowerID"`
//CreatedAt   time.Time
//UpdatedAt   time.Time
//DeletedAt   gorm.DeletedAt `gorm:"index"`
//}

func NewAuthRepo(db *gorm.DB) AuthRepo {
	return &authRepo{
		DB: db,
	}
}

func (a *authRepo) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	// query the db for the user with the given email
	result := a.DB.WithContext(ctx).Where("email=?", email).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) == true {
			return nil, nil
		}
		utils.LogError("failed to get userbyemail", result.Error)
		return nil, result.Error
	}
	return &user, nil
}

func (a *authRepo) AddUser(ctx context.Context, user *models.User) error {
	return a.DB.WithContext(ctx).Create(user).Error
}

func (a *authRepo) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	var user models.User
	result := a.DB.WithContext(ctx).Where("id=?", id).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			// user not found
			return nil, nil
		}
		utils.LogError("failed to get user by id", result.Error)
		return nil, result.Error
	}
	return nil, nil
}
