package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// =======================
// User Model
// =======================
// A User is an identity. Money does not live here — it lives on Account, so a
// single user can hold a simulated account and (later) a live-money account
// without schema changes. Engine-driven investor personas are Users with
// IsBot=true plus an Investor profile.
type User struct {
	ID            uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Name          string    `gorm:"type:varchar(100);not null"`
	Email         string    `gorm:"type:varchar(100);uniqueIndex;not null"`
	Password      string    `gorm:"not null"`                                 // bcrypt hash, never plaintext
	Role          string    `gorm:"type:varchar(20);not null;default:'user'"` // user | admin
	IsBot         bool      `gorm:"not null;default:false"`                   // engine-driven investor persona
	IsActive      bool      `gorm:"default:true"`
	EmailVerified bool      `gorm:"not null;default:false"`
	LastLoginAt   *time.Time
	Accounts      []Account `gorm:"foreignKey:UserID"` // sim and/or live trading accounts
	Follows       []Follow  `gorm:"foreignKey:FollowerID"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     gorm.DeletedAt `gorm:"index"`
}

// =======================
// Account Model (sim / live money)
// =======================
// Account isolates money from identity and is the seam for the real↔fake-money
// toggle: the trade service selects a Broker implementation from Account.Mode,
// so flipping a user to live trading is a data change, not a code change.
type Account struct {
	ID        uuid.UUID   `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	UserID    uuid.UUID   `gorm:"type:uuid;not null;uniqueIndex:idx_user_mode"`
	User      User        `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	Mode      AccountMode `gorm:"type:varchar(4);not null;default:'sim';uniqueIndex:idx_user_mode"` // one account per mode per user
	Currency  string      `gorm:"type:varchar(3);not null;default:'USD'"`
	Balance   float64     `gorm:"type:numeric(18,4);not null;default:100000"` // cash available to trade
	IsActive  bool        `gorm:"not null;default:true"`
	Holdings  []Holding   `gorm:"foreignKey:AccountID"`
	Trades    []Trade     `gorm:"foreignKey:AccountID"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

// AccountMode enumerates the money mode of an Account.
type AccountMode string

const (
	ModeSim  AccountMode = "sim"  // virtual money (SimulatedBroker)
	ModeLive AccountMode = "live" // real money (LiveBroker — future)
)

// StartingSimBalance is the virtual cash a new simulated account begins with.
// It is the baseline capital ROI is measured against until deposit/withdrawal
// accounting exists.
const StartingSimBalance = 100000

// =======================
// RefreshToken Model
// =======================
// Stores only the SHA-256 hash of an opaque refresh token (never the token
// itself), supporting rotation and revocation for the auth flow.
type RefreshToken struct {
	ID        uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;index"`
	TokenHash string    `gorm:"type:varchar(64);not null;uniqueIndex"`
	ExpiresAt time.Time `gorm:"not null;index"`
	RevokedAt *time.Time
	CreatedAt time.Time
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
	ID           uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Symbol       string    `gorm:"type:varchar(25);not null;uniqueIndex"`
	Name         string    `gorm:"type:varchar(100);not null"`
	Exchange     string    `gorm:"type:varchar(50)"`
	Sector       string    `gorm:"type:varchar(50);index"`                           // used by strategy universe filters
	AssetClass   string    `gorm:"type:varchar(20);not null;default:'equity';index"` // equity|bond|gold|commodity (for macro allocation)
	CurrentPrice float64   `gorm:"type:numeric(10,2);not null"`
	Currency     string    `gorm:"type:varchar(3);default:'USD'"`
	// Fundamentals is stored as a single JSONB column. The engine loads it
	// per-stock into memory to evaluate strategies (it does not screen on
	// individual metrics in SQL), so a JSONB blob beats a wide column table.
	Fundamentals Fundamentals `gorm:"type:jsonb;serializer:json"`
	Prices       []StockPrice `gorm:"foreignKey:StockID"`
	Trades       []Trade      `gorm:"foreignKey:StockID"`
	Holdings     []Holding    `gorm:"foreignKey:StockID"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    gorm.DeletedAt `gorm:"index"`
}

// =======================
// Fundamentals (value object, JSONB on Stock)
// =======================
// Fundamentals is the per-stock metric set the strategy engine evaluates. It is
// a value object — refreshed periodically by the engine's slow lane and embedded
// in Stock as JSONB. Field names mirror the strategy schema's metric vocabulary.
type Fundamentals struct {
	// Valuation
	PE            float64 `json:"pe"`
	ForwardPE     float64 `json:"forward_pe"`
	PB            float64 `json:"pb"`
	PS            float64 `json:"ps"`
	PEG           float64 `json:"peg"`
	EVEBITDA      float64 `json:"ev_ebitda"`
	EarningsYield float64 `json:"earnings_yield"` // EBIT/EV
	FCFYield      float64 `json:"fcf_yield"`
	DividendYield float64 `json:"dividend_yield"`
	EPSTTM        float64 `json:"eps_ttm"`
	BVPS          float64 `json:"bvps"`

	// Quality
	ROE             float64 `json:"roe"`
	ROIC            float64 `json:"roic"`
	GrossMargin     float64 `json:"gross_margin"`
	OperatingMargin float64 `json:"operating_margin"`
	NetMargin       float64 `json:"net_margin"`
	DebtToEquity    float64 `json:"debt_to_equity"`
	CurrentRatio    float64 `json:"current_ratio"`
	InterestCover   float64 `json:"interest_coverage"`
	FCFPositive     bool    `json:"fcf_positive"`

	// Growth
	RevenueGrowthYoY float64 `json:"revenue_growth_yoy"`
	EPSGrowthYoY     float64 `json:"eps_growth_yoy"`
	RevenueCAGR3Y    float64 `json:"revenue_cagr_3y"`
	EPSGrowth5Y      float64 `json:"eps_growth_5y"`

	// Stability & size
	EPSPositiveYears int     `json:"eps_positive_years"`
	DividendYears    int     `json:"dividend_years"`
	Beta             float64 `json:"beta"`
	MarketCap        float64 `json:"market_cap"`
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
// A Trade belongs to an Account (sim or live), not directly to a User — so the
// same trade machinery serves human accounts and engine bot accounts alike.
type Trade struct {
	ID         uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	AccountID  uuid.UUID `gorm:"type:uuid;not null;index:idx_account_executed"`
	Account    Account   `gorm:"foreignKey:AccountID;constraint:OnDelete:CASCADE"`
	StockID    uuid.UUID `gorm:"type:uuid;not null;index"`
	Stock      Stock     `gorm:"foreignKey:StockID;constraint:OnDelete:CASCADE"`
	Type       string    `gorm:"type:varchar(4);not null" validate:"oneof=buy sell"`
	Quantity   float64   `gorm:"type:numeric(18,4);not null" validate:"gt=0"`
	Price      float64   `gorm:"type:numeric(18,4);not null" validate:"gt=0"`
	TotalValue float64   `gorm:"type:numeric(18,4)"` // Price * Quantity (for easier queries)
	ExecutedAt time.Time `gorm:"not null;index:idx_account_executed"`
	Status     string    `gorm:"type:varchar(20);default:'executed'"` // pending, executed, cancelled
	// IdempotencyKey makes a buy/sell safe to retry: a duplicate key returns the
	// original trade instead of executing twice. Nullable (Postgres allows many
	// NULLs under a unique index), so non-idempotent trades are unaffected.
	IdempotencyKey *string    `gorm:"type:varchar(80);uniqueIndex"`
	InvestorID     *uuid.UUID `gorm:"type:uuid"` // nullable: who was followed for this copy trade
	CreatedAt      time.Time
	DeletedAt      gorm.DeletedAt `gorm:"index"`
}

// =======================
// Holding Model (Portfolio)
// =======================
// A Holding is an account's position in a stock — unique per (account, stock).
type Holding struct {
	ID        uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	AccountID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_account_stock"`
	StockID   uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_account_stock"`
	Account   Account   `gorm:"foreignKey:AccountID;constraint:OnDelete:CASCADE"`
	Stock     Stock     `gorm:"foreignKey:StockID;constraint:OnDelete:CASCADE"`
	Quantity  float64   `gorm:"type:numeric(18,4);not null"`
	AvgPrice  float64   `gorm:"type:numeric(18,4);not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
}
