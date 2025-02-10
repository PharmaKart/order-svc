package utils

import (
	"github.com/PharmaKart/order-svc/pkg/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// ConnectDB connects to the database
func ConnectDB(cfg *config.Config) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(cfg.DBConnString), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	return db, nil
}
