package database

import (
	"github.com/Jennifer-ra/accounting/services/go-api/internal/config"
	"github.com/Jennifer-ra/accounting/services/go-api/internal/model"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func Open(cfg config.Config) (*gorm.DB, error) {
	return gorm.Open(mysql.Open(cfg.MySQLDSN), &gorm.Config{TranslateError: true})
}

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&model.User{},
		&model.Book{},
		&model.BookMember{},
		&model.Category{},
		&model.Account{},
		&model.Transaction{},
		&model.Budget{},
	)
}
