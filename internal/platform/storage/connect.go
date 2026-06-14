// Package storage owns the database connection and schema migration. It is a
// library: it returns errors rather than calling log.Fatal, and it does not read
// the environment directly — callers pass a resolved config.DBConfig.
package storage

import (
	"fmt"

	"github.com/faisal-990/ProjectInvestApp/internal/platform/config"
	"github.com/faisal-990/ProjectInvestApp/internal/platform/models"
	"github.com/faisal-990/ProjectInvestApp/internal/platform/utils"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Models returns every GORM model managed by AutoMigrate, in dependency order
// (parents before children) so foreign keys resolve cleanly.
func Models() []any {
	return []any{
		&models.User{},
		&models.Account{},
		&models.RefreshToken{},
		&models.PasswordReset{},
		&models.Investor{},
		&models.Performance{},
		&models.Stock{},
		&models.StockPrice{},
		&models.Holding{},
		&models.Trade{},
		&models.Follow{},
	}
}

// Connect opens a GORM/Postgres connection from the given config and runs
// AutoMigrate. AutoMigrate is appropriate for development; versioned migrations
// (golang-migrate) replace it during production hardening (T10).
func Connect(cfg config.DBConfig) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("storage: opening connection: %w", err)
	}

	if err := db.AutoMigrate(Models()...); err != nil {
		return nil, fmt.Errorf("storage: auto-migrating schema: %w", err)
	}

	utils.LogInfo("database connected and schema migrated")
	return db, nil
}
