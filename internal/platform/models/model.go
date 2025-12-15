package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// =======================
// User Model
// =======================
type User struct {
	ID          uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Name        string    `gorm:"type:varchar(100);not null"`
	Email       string    `gorm:"type:varchar(100);uniqueIndex;not null"`
	Password    string    `gorm:"not null"`
	IsActive    bool      `gorm:"default:true"`
	LastLoginAt *time.Time
	Balance     float64   `gorm:"type:numeric(15,2);default:100000"` // Starting simulation balance
	Trades      []Trade   `gorm:"foreignKey:UserID"`
	Holdings    []Holding `gorm:"foreignKey:UserID"`
	Follows     []Follow  `gorm:"foreignKey:FollowerID"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}

// =======================
// Performance Summary Model
// =======================
type Performance struct {
	ID         uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	InvestorID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex"`
	// Investor   Investor  `gorm:"foreignKey:InvestorID;constraint:OnDelete:CASCADE"`
	ROI        float64   `gorm:"type:numeric(5,2);not null"`
	Rank       int       `gorm:"not null"`
	LastUpdate time.Time `gorm:"not null"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// =======================
// Investor Model (1:1 with User)
// =======================
type Investor struct {
	ID          uuid.UUID   `gorm:"type:uuid;primaryKey"` // same as User ID
	Bio         string      `gorm:"type:text"`
	Strategy    string      `gorm:"type:text"`
	Followers   []Follow    `gorm:"foreignKey:InvestorID"`
	Performance Performance `gorm:"foreignKey:InvestorID"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// =======================
// Follow Model (Join Table)
// =======================
type Follow struct {
	ID         uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	InvestorID uuid.UUID `gorm:"type:uuid;not null;index"`
	FollowerID uuid.UUID `gorm:"type:uuid;not null;index"`
	CreatedAt  time.Time
	Investor   Investor `gorm:"foreignKey:InvestorID;constraint:OnDelete:CASCADE"`
	Follower   User     `gorm:"foreignKey:FollowerID;constraint:OnDelete:CASCADE"`
}

// =======================
// Stock Model
// =======================
type Stock struct {
	ID           uuid.UUID    `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Symbol       string       `gorm:"type:varchar(25);not null;uniqueIndex"`
	Name         string       `gorm:"type:varchar(100);not null"`
	Exchange     string       `gorm:"type:varchar(50)"`
	CurrentPrice float64      `gorm:"type:numeric(10,2);not null"`
	Currency     string       `gorm:"type:varchar(3);default:'USD'"`
	Prices       []StockPrice `gorm:"foreignKey:StockID"`
	Trades       []Trade      `gorm:"foreignKey:StockID"`
	Holdings     []Holding    `gorm:"foreignKey:StockID"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    gorm.DeletedAt `gorm:"index"`
}

// =======================
// StockPrice Model (Time Series Data)
// =======================
type StockPrice struct {
	ID        uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	StockID   uuid.UUID `gorm:"type:uuid;not null;index"`
	Stock     Stock     `gorm:"foreignKey:StockID;constraint:OnDelete:CASCADE"`
	Timestamp time.Time `gorm:"not null;index"`
	Open      float64   `gorm:"type:numeric(10,2)"`
	Close     float64   `gorm:"type:numeric(10,2)"`
	High      float64   `gorm:"type:numeric(10,2)"`
	Low       float64   `gorm:"type:numeric(10,2)"`
	Volume    int64     `gorm:"type:bigint"`
	Interval  string    `gorm:"type:varchar(10);not null"`
	CreatedAt time.Time
}

// =======================
// Trade Model
// =======================
type Trade struct {
	ID          uuid.UUID  `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	UserID      uuid.UUID  `gorm:"type:uuid;not null;index:idx_user_executed"`
	User        User       `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	StockID     uuid.UUID  `gorm:"type:uuid;not null;index"`
	Stock       Stock      `gorm:"foreignKey:StockID;constraint:OnDelete:CASCADE"`
	Type        string     `gorm:"type:varchar(4);not null" validate:"oneof=buy sell"`
	Quantity    float64    `gorm:"type:numeric(10,2);not null" validate:"gt=0"`
	Price       float64    `gorm:"type:numeric(10,2);not null" validate:"gt=0"`
	TotalValue  float64    `gorm:"type:numeric(15,2)"` // Price * Quantity (for easier queries)
	ExecutedAt  time.Time  `gorm:"not null;index:idx_user_executed"`
	Status      string     `gorm:"type:varchar(20);default:'executed'"` // pending, executed, cancelled
	IsSimulated bool       `gorm:"default:true"`
	InvestorID  *uuid.UUID `gorm:"type:uuid"` // nullable: who was followed for this copy trade
	CreatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}

// =======================
// Holding Model (Portfolio)
// =======================
type Holding struct {
	ID        uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_user_stock"`
	StockID   uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_user_stock"`
	User      User      `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	Stock     Stock     `gorm:"foreignKey:StockID;constraint:OnDelete:CASCADE"`
	Quantity  float64   `gorm:"type:numeric(10,2);not null"`
	AvgPrice  float64   `gorm:"type:numeric(10,2);not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
}
